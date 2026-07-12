package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/source/tweedekamer"
)

const (
	tweedeKamerSourceKey = "tweedekamer-odata-v2"
	motionsPipeline      = "motions.raw"
	zaakCollection       = "Zaak"
)

type TweedeKamerMotionIngest struct {
	Pool          *pgxpool.Pool
	Client        *tweedekamer.Client
	BatchSize     int
	MaxPages      int
	InitialSince  time.Time
	CursorOverlap time.Duration
	SinceOverride *time.Time
	ResetCursor   bool
}

func (ingest TweedeKamerMotionIngest) Run(ctx context.Context) error {
	releaseLock, err := ingest.acquireLock(ctx)
	if err != nil {
		return err
	}
	defer releaseLock()

	source, err := ingest.getSource(ctx)
	if err != nil {
		return err
	}

	if ingest.ResetCursor {
		if err := ingest.resetCursor(ctx); err != nil {
			return err
		}
	}

	cursorBefore, err := ingest.getCursor(ctx)
	if err != nil {
		return err
	}
	if ingest.SinceOverride != nil {
		cursorBefore = Cursor{ApiUpdatedAt: ingest.SinceOverride}
	}

	since := ingest.cursorSince(cursorBefore)
	if ingest.SinceOverride != nil {
		since = *ingest.SinceOverride
	}
	runID, err := ingest.startRun(ctx, cursorBefore)
	if err != nil {
		return err
	}

	recordsSeen := 0
	recordsChanged := 0
	maxUpdatedAt := cursorBefore.ApiUpdatedAt
	nextURL := ""
	stopReason := ""
	skip := 0

	pagesProcessed := 0
	for page := 1; ; page++ {
		result, err := ingest.Client.FetchChangedMotions(ctx, since, ingest.BatchSize, skip, nextURL)
		if err != nil {
			_ = ingest.finishRun(ctx, runID, "failed", cursorBefore, recordsSeen, recordsChanged, false, "error", err.Error())
			return err
		}

		nextURL = result.NextURL
		recordsSeen += len(result.Records)
		skip += len(result.Records)
		pagesProcessed = page

		for _, record := range result.Records {
			changed, err := ingest.storeMotionRecord(ctx, source.JurisdictionKey, record)
			if err != nil {
				_ = ingest.finishRun(ctx, runID, "failed", cursorBefore, recordsSeen, recordsChanged, false, "error", err.Error())
				return err
			}
			if changed {
				recordsChanged++
			}

			apiUpdatedAt := timePtr(record.ApiGewijzigdOp)
			if apiUpdatedAt != nil && (maxUpdatedAt == nil || apiUpdatedAt.After(*maxUpdatedAt)) {
				value := *apiUpdatedAt
				maxUpdatedAt = &value
			}
		}

		hasMore := nextURL != "" || len(result.Records) == ingest.BatchSize

		fmt.Printf("page=%d seen=%d changed=%d next=%t\n", page, recordsSeen, recordsChanged, hasMore)

		if !hasMore {
			stopReason = "complete"
			break
		}
		if ingest.MaxPages > 0 && page >= ingest.MaxPages {
			stopReason = "max_pages"
			break
		}
	}

	cursorAfter := Cursor{ApiUpdatedAt: maxUpdatedAt}
	cursorSaved := stopReason == "complete"

	if cursorSaved {
		if err := ingest.saveCursor(ctx, cursorAfter); err != nil {
			_ = ingest.finishRun(ctx, runID, "failed", cursorAfter, recordsSeen, recordsChanged, false, stopReason, err.Error())
			return err
		}
	} else {
		cursorAfter = cursorBefore
	}

	if err := ingest.finishRun(ctx, runID, "succeeded", cursorAfter, recordsSeen, recordsChanged, cursorSaved, stopReason, ""); err != nil {
		return err
	}

	fmt.Printf(
		"ingestion complete run_id=%d pages=%d seen=%d changed=%d cursor_before=%s cursor_after=%s cursor_saved=%t stop_reason=%s\n",
		runID,
		pagesProcessed,
		recordsSeen,
		recordsChanged,
		formatCursor(cursorBefore),
		formatCursor(cursorAfter),
		cursorSaved,
		stopReason,
	)
	return nil
}

type source struct {
	SourceKey       string
	JurisdictionKey string
	BaseURL         string
}

func (ingest TweedeKamerMotionIngest) getSource(ctx context.Context) (source, error) {
	var result source
	err := ingest.Pool.QueryRow(ctx, `
		SELECT source_key, jurisdiction_key, base_url
		FROM data_sources
		WHERE source_key = $1 AND enabled = true
	`, tweedeKamerSourceKey).Scan(&result.SourceKey, &result.JurisdictionKey, &result.BaseURL)
	if err != nil {
		return source{}, fmt.Errorf("get data source %s: %w", tweedeKamerSourceKey, err)
	}
	return result, nil
}

func (ingest TweedeKamerMotionIngest) acquireLock(ctx context.Context) (func(), error) {
	conn, err := ingest.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	var acquired bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", tweedeKamerSourceKey+":"+motionsPipeline).Scan(&acquired); err != nil {
		conn.Release()
		return nil, err
	}
	if !acquired {
		conn.Release()
		return nil, fmt.Errorf("ingestion pipeline already running: %s/%s", tweedeKamerSourceKey, motionsPipeline)
	}

	return func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", tweedeKamerSourceKey+":"+motionsPipeline)
		conn.Release()
	}, nil
}

type Cursor struct {
	ApiUpdatedAt *time.Time `json:"apiUpdatedAt"`
}

func (ingest TweedeKamerMotionIngest) getCursor(ctx context.Context) (Cursor, error) {
	var raw []byte
	err := ingest.Pool.QueryRow(ctx, `
		SELECT cursor
		FROM source_cursors
		WHERE source_key = $1 AND pipeline = $2
	`, tweedeKamerSourceKey, motionsPipeline).Scan(&raw)
	if err == pgx.ErrNoRows {
		initial := ingest.InitialSince
		return Cursor{ApiUpdatedAt: &initial}, nil
	}
	if err != nil {
		return Cursor{}, err
	}

	var cursor Cursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return Cursor{}, err
	}
	return cursor, nil
}

func (ingest TweedeKamerMotionIngest) resetCursor(ctx context.Context) error {
	_, err := ingest.Pool.Exec(ctx, `
		DELETE FROM source_cursors
		WHERE source_key = $1 AND pipeline = $2
	`, tweedeKamerSourceKey, motionsPipeline)
	return err
}

func (ingest TweedeKamerMotionIngest) cursorSince(cursor Cursor) time.Time {
	if cursor.ApiUpdatedAt == nil {
		return ingest.InitialSince
	}
	if cursor.ApiUpdatedAt.Equal(ingest.InitialSince) {
		return ingest.InitialSince
	}
	return cursor.ApiUpdatedAt.Add(-ingest.CursorOverlap)
}

func (ingest TweedeKamerMotionIngest) startRun(ctx context.Context, cursorBefore Cursor) (int64, error) {
	raw, err := json.Marshal(cursorBefore)
	if err != nil {
		return 0, err
	}

	var runID int64
	err = ingest.Pool.QueryRow(ctx, `
		INSERT INTO ingestion_runs (source_key, pipeline, status, cursor_before)
		VALUES ($1, $2, 'running', $3)
		RETURNING id
	`, tweedeKamerSourceKey, motionsPipeline, string(raw)).Scan(&runID)
	return runID, err
}

func (ingest TweedeKamerMotionIngest) finishRun(
	ctx context.Context,
	runID int64,
	status string,
	cursorAfter Cursor,
	recordsSeen int,
	recordsChanged int,
	cursorSaved bool,
	stopReason string,
	errorMessage string,
) error {
	raw, err := json.Marshal(cursorAfter)
	if err != nil {
		return err
	}

	var nullableError *string
	if errorMessage != "" {
		nullableError = &errorMessage
	}

	_, err = ingest.Pool.Exec(ctx, `
		UPDATE ingestion_runs
		SET status = $2,
		    cursor_after = $3,
		    records_seen = $4,
		    records_changed = $5,
		    error_message = $6,
		    cursor_saved = $7,
		    stop_reason = $8,
		    finished_at = now()
		WHERE id = $1
	`, runID, status, string(raw), recordsSeen, recordsChanged, nullableError, cursorSaved, stopReason)
	return err
}

func (ingest TweedeKamerMotionIngest) saveCursor(ctx context.Context, cursorAfter Cursor) error {
	raw, err := json.Marshal(cursorAfter)
	if err != nil {
		return err
	}

	_, err = ingest.Pool.Exec(ctx, `
		INSERT INTO source_cursors (source_key, pipeline, cursor, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (source_key, pipeline)
		DO UPDATE SET cursor = EXCLUDED.cursor,
		              updated_at = now()
	`, tweedeKamerSourceKey, motionsPipeline, string(raw))
	return err
}

func (ingest TweedeKamerMotionIngest) storeMotionRecord(
	ctx context.Context,
	jurisdictionKey string,
	record tweedekamer.MotionRecord,
) (bool, error) {
	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	raw := projectMotionRaw(record)
	motion := projectMotion(jurisdictionKey, record)

	tag, err := tx.Exec(ctx, `
		INSERT INTO raw_records (
			source_key,
			collection,
			source_id,
			source_updated_at,
			source_deleted,
			payload,
			payload_hash
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_key, collection, source_id)
		DO UPDATE SET source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              payload = EXCLUDED.payload,
		              payload_hash = EXCLUDED.payload_hash,
		              last_seen_at = now()
		WHERE raw_records.payload_hash IS DISTINCT FROM EXCLUDED.payload_hash
		   OR raw_records.source_updated_at IS DISTINCT FROM EXCLUDED.source_updated_at
		   OR raw_records.source_deleted IS DISTINCT FROM EXCLUDED.source_deleted
	`, tweedeKamerSourceKey, raw.Collection, raw.SourceID, raw.SourceUpdatedAt, raw.SourceDeleted, string(raw.Payload), raw.PayloadHash)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO motions (
			motion_key,
			source_key,
			jurisdiction_key,
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
			raw_collection,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now())
		ON CONFLICT (source_key, source_id)
		DO UPDATE SET number = EXCLUDED.number,
		              title = EXCLUDED.title,
		              subject = EXCLUDED.subject,
		              status = EXCLUDED.status,
		              kind = EXCLUDED.kind,
		              parliamentary_year = EXCLUDED.parliamentary_year,
		              proposed_at = EXCLUDED.proposed_at,
		              source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              updated_at = now()
	`, motion.MotionKey, motion.SourceKey, motion.JurisdictionKey, motion.SourceID, motion.Number, motion.Title, motion.Subject, motion.Status, motion.Kind, motion.ParliamentaryYear, motion.ProposedAt, motion.SourceUpdatedAt, motion.SourceDeleted, motion.RawCollection)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return tag.RowsAffected() > 0, nil
}

func title(record tweedekamer.MotionRecord) *string {
	for _, value := range []*string{record.Titel, record.Citeertitel, record.Onderwerp} {
		if value != nil && *value != "" {
			return value
		}
	}
	return nil
}

func proposedAt(record tweedekamer.MotionRecord) *time.Time {
	return timePtr(record.GestartOp)
}

func timePtr(value *tweedekamer.Time) *time.Time {
	if value == nil {
		return nil
	}
	return &value.Time
}

func motionKey(sourceID string) string {
	return tweedeKamerSourceKey + ":" + sourceID
}

func hashBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func formatCursor(cursor Cursor) string {
	if cursor.ApiUpdatedAt == nil {
		return "null"
	}
	return cursor.ApiUpdatedAt.UTC().Format(time.RFC3339Nano)
}
