package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const shutdownTimeout = 10 * time.Second

type Server struct {
	Pool *pgxpool.Pool
}

func (server Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("GET /api/ingestion-runs", server.listIngestionRuns)
	mux.HandleFunc("GET /api/motions", server.listMotions)
	mux.HandleFunc("GET /api/motions/{motionKey}", server.getMotion)
	return mux
}

func (server Server) health(response http.ResponseWriter, request *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (server Server) listIngestionRuns(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	limit := clamp(parseInt(query.Get("limit"), 10), 1, 100)

	rows, err := server.Pool.Query(request.Context(), `
		SELECT id,
		       source_key,
		       pipeline,
		       status,
		       cursor_before,
		       cursor_after,
		       cursor_saved,
		       stop_reason,
		       records_seen,
		       records_changed,
		       error_message,
		       started_at,
		       finished_at
		FROM ingestion_runs
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		writeError(response, err)
		return
	}
	defer rows.Close()

	runs := []map[string]any{}
	for rows.Next() {
		var run ingestionRunRow
		if err := rows.Scan(
			&run.ID,
			&run.SourceKey,
			&run.Pipeline,
			&run.Status,
			&run.CursorBefore,
			&run.CursorAfter,
			&run.CursorSaved,
			&run.StopReason,
			&run.RecordsSeen,
			&run.RecordsChanged,
			&run.ErrorMessage,
			&run.StartedAt,
			&run.FinishedAt,
		); err != nil {
			writeError(response, err)
			return
		}
		runs = append(runs, run.mapValue())
	}
	if err := rows.Err(); err != nil {
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"runs":  runs,
		"limit": limit,
	})
}

func (server Server) listMotions(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	query := request.URL.Query()
	limit := clamp(parseInt(query.Get("limit"), 25), 1, 100)
	offset := max(parseInt(query.Get("offset"), 0), 0)
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	search := query.Get("search")
	var searchPtr *string
	if strings.TrimSpace(search) != "" {
		searchPtr = &search
	}

	rows, err := server.Pool.Query(ctx, `
		SELECT motion_key,
		       source_id,
		       number,
		       title,
		       subject,
		       status,
		       kind,
		       parliamentary_year,
		       proposed_at,
		       source_updated_at,
		       source_deleted,
		       count(*) OVER ()::int AS total
		FROM motions
		WHERE jurisdiction_key = $1
		  AND source_deleted = false
		  AND (
		    $2::text IS NULL
		    OR title ILIKE '%' || $2 || '%'
		    OR subject ILIKE '%' || $2 || '%'
		    OR number ILIKE '%' || $2 || '%'
		  )
		ORDER BY proposed_at DESC NULLS LAST, source_updated_at DESC NULLS LAST
		LIMIT $3
		OFFSET $4
	`, jurisdiction, searchPtr, limit, offset)
	if err != nil {
		writeError(response, err)
		return
	}
	defer rows.Close()

	motions := []map[string]any{}
	total := 0
	for rows.Next() {
		var motion motionRow
		if err := motion.scan(rows.Scan); err != nil {
			writeError(response, err)
			return
		}
		total = motion.Total
		motions = append(motions, motion.mapValue())
	}
	if err := rows.Err(); err != nil {
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"motions": motions,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"hasMore": offset+limit < total,
	})
}

func (server Server) getMotion(response http.ResponseWriter, request *http.Request) {
	motionKey := request.PathValue("motionKey")

	var motion motionRow
	err := server.Pool.QueryRow(request.Context(), `
		SELECT motion_key,
		       source_id,
		       number,
		       title,
		       subject,
		       status,
		       kind,
		       parliamentary_year,
		       proposed_at,
		       source_updated_at,
		       source_deleted,
		       1 AS total
		FROM motions
		WHERE motion_key = $1
	`, motionKey).Scan(
		&motion.MotionKey,
		&motion.SourceID,
		&motion.Number,
		&motion.Title,
		&motion.Subject,
		&motion.Status,
		&motion.Kind,
		&motion.ParliamentaryYear,
		&motion.ProposedAt,
		&motion.SourceUpdatedAt,
		&motion.SourceDeleted,
		&motion.Total,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(response, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusOK, motion.mapValue())
}

type scanner func(dest ...any) error

type ingestionRunRow struct {
	ID             int64
	SourceKey      string
	Pipeline       string
	Status         string
	CursorBefore   []byte
	CursorAfter    []byte
	CursorSaved    bool
	StopReason     *string
	RecordsSeen    int
	RecordsChanged int
	ErrorMessage   *string
	StartedAt      time.Time
	FinishedAt     *time.Time
}

func (row ingestionRunRow) mapValue() map[string]any {
	return map[string]any{
		"id":             row.ID,
		"sourceKey":      row.SourceKey,
		"pipeline":       row.Pipeline,
		"status":         row.Status,
		"cursorBefore":   json.RawMessage(row.CursorBefore),
		"cursorAfter":    json.RawMessage(row.CursorAfter),
		"cursorSaved":    row.CursorSaved,
		"stopReason":     row.StopReason,
		"recordsSeen":    row.RecordsSeen,
		"recordsChanged": row.RecordsChanged,
		"errorMessage":   row.ErrorMessage,
		"startedAt":      row.StartedAt,
		"finishedAt":     row.FinishedAt,
	}
}

type motionRow struct {
	MotionKey         string
	SourceID          string
	Number            *string
	Title             *string
	Subject           *string
	Status            *string
	Kind              *string
	ParliamentaryYear *string
	ProposedAt        *time.Time
	SourceUpdatedAt   *time.Time
	SourceDeleted     bool
	Total             int
}

func (row *motionRow) scan(scan scanner) error {
	return scan(
		&row.MotionKey,
		&row.SourceID,
		&row.Number,
		&row.Title,
		&row.Subject,
		&row.Status,
		&row.Kind,
		&row.ParliamentaryYear,
		&row.ProposedAt,
		&row.SourceUpdatedAt,
		&row.SourceDeleted,
		&row.Total,
	)
}

func (row motionRow) mapValue() map[string]any {
	return map[string]any{
		"motionKey":         row.MotionKey,
		"sourceId":          row.SourceID,
		"number":            row.Number,
		"title":             row.Title,
		"subject":           row.Subject,
		"status":            row.Status,
		"kind":              row.Kind,
		"parliamentaryYear": row.ParliamentaryYear,
		"proposedAt":        row.ProposedAt,
		"sourceUpdatedAt":   row.SourceUpdatedAt,
		"sourceDeleted":     row.SourceDeleted,
	}
}

func ListenAndServe(ctx context.Context, address string, handler http.Handler) error {
	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	errs := make(chan error, 1)
	go func() {
		errs <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errs:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, err error) {
	fmt.Println(err)
	writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func clamp(value int, minValue int, maxValue int) int {
	return min(max(value, minValue), maxValue)
}
