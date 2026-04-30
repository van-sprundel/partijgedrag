package web

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/rewrite/internal/analysis"
	"partijgedrag/rewrite/internal/politics"
	"partijgedrag/rewrite/internal/status"
)

//go:embed templates/*.html static/styles.css
var files embed.FS

type Server struct {
	Pool      *pgxpool.Pool
	templates map[string]*template.Template
}

func MustNew(pool *pgxpool.Pool) Server {
	server, err := New(pool)
	if err != nil {
		panic(err)
	}
	return server
}

func New(pool *pgxpool.Pool) (Server, error) {
	templates := make(map[string]*template.Template)
	for _, name := range []string{"home", "motions", "motion", "party_likeness"} {
		parsed, err := parseTemplate(name)
		if err != nil {
			return Server{}, err
		}
		templates[name] = parsed
	}

	return Server{
		Pool:      pool,
		templates: templates,
	}, nil
}

func (server Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /static/styles.css", server.styles)
	mux.HandleFunc("GET /", server.home)
	mux.HandleFunc("GET /party-likeness", server.partyLikeness)
	mux.HandleFunc("GET /motions", server.motions)
	mux.HandleFunc("GET /motions/{motionKey}", server.motion)
}

func parseTemplate(name string) (*template.Template, error) {
	base, err := files.ReadFile("templates/base.html")
	if err != nil {
		return nil, err
	}
	page, err := files.ReadFile("templates/" + name + ".html")
	if err != nil {
		return nil, err
	}

	tmpl := template.New(name).Funcs(template.FuncMap{
		"date":     dateValue,
		"fallback": fallback,
		"likeness": likenessValue,
		"time":     timeValue,
	})
	return tmpl.Parse(string(base) + "\n" + string(page))
}

func (server Server) styles(response http.ResponseWriter, request *http.Request) {
	content, err := files.ReadFile("static/styles.css")
	if err != nil {
		http.Error(response, "static asset not found", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = response.Write(content)
}

func (server Server) home(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(response, request)
		return
	}

	summary, err := status.LoadSummary(request.Context(), server.Pool)
	if err != nil {
		writeError(response, err)
		return
	}
	runs, err := loadIngestionRuns(request.Context(), server.Pool, 10)
	if err != nil {
		writeError(response, err)
		return
	}

	server.render(response, "home", homePage{
		Summary: summary,
		Runs:    runs,
	})
}

func (server Server) motions(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	limit := clamp(parseInt(query.Get("limit"), 25), 1, 100)
	offset := max(parseInt(query.Get("offset"), 0), 0)
	search := strings.TrimSpace(query.Get("search"))
	withVotes := query.Get("withVotes") == "true"

	motions, total, err := loadMotions(request.Context(), server.Pool, motionListOptions{
		Jurisdiction: "nl-tweede-kamer",
		Search:       search,
		WithVotes:    withVotes,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	page := motionsPage{
		Motions:   motions,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		Search:    search,
		WithVotes: withVotes,
	}
	if offset > 0 {
		page.PrevURL = motionsURL(search, withVotes, limit, max(offset-limit, 0))
	}
	if offset+limit < total {
		page.NextURL = motionsURL(search, withVotes, limit, offset+limit)
	}

	server.render(response, "motions", page)
}

func (server Server) motion(response http.ResponseWriter, request *http.Request) {
	motionKey := request.PathValue("motionKey")

	motion, err := loadMotion(request.Context(), server.Pool, motionKey)
	if err == pgx.ErrNoRows {
		http.NotFound(response, request)
		return
	}
	if err != nil {
		writeError(response, err)
		return
	}
	decisions, err := loadDecisions(request.Context(), server.Pool, motionKey)
	if err != nil {
		writeError(response, err)
		return
	}
	positions, err := loadPartyPositions(request.Context(), server.Pool, motionKey)
	if err != nil {
		writeError(response, err)
		return
	}

	server.render(response, "motion", motionPage{
		Motion:    motion,
		Decisions: decisions,
		Positions: positions,
	})
}

func (server Server) partyLikeness(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	dateFrom, err := parseDate(query.Get("dateFrom"))
	if err != nil {
		http.Error(response, "invalid dateFrom", http.StatusBadRequest)
		return
	}
	dateTo, err := parseDate(query.Get("dateTo"))
	if err != nil {
		http.Error(response, "invalid dateTo", http.StatusBadRequest)
		return
	}
	minCommon := clamp(parseInt(query.Get("minCommon"), 10), 1, 1000)

	rows, err := analysis.LoadPartyLikeness(request.Context(), server.Pool, analysis.PartyLikenessOptions{
		Jurisdiction: "nl-tweede-kamer",
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		MinCommon:    minCommon,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	server.render(response, "party_likeness", partyLikenessPage{
		Parties:   likenessParties(rows),
		Rows:      rows,
		Matrix:    likenessMatrix(rows),
		DateFrom:  query.Get("dateFrom"),
		DateTo:    query.Get("dateTo"),
		MinCommon: minCommon,
	})
}

func (server Server) render(response http.ResponseWriter, name string, data any) {
	var buffer bytes.Buffer
	if err := server.templates[name].ExecuteTemplate(&buffer, "base", data); err != nil {
		writeError(response, err)
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = response.Write(buffer.Bytes())
}

func loadIngestionRuns(ctx context.Context, pool *pgxpool.Pool, limit int) ([]ingestionRun, error) {
	rows, err := pool.Query(ctx, `
		SELECT id,
		       pipeline,
		       status,
		       COALESCE(stop_reason, '') AS stop_reason,
		       records_seen,
		       records_changed,
		       started_at
		FROM ingestion_runs
		ORDER BY started_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []ingestionRun{}
	for rows.Next() {
		var run ingestionRun
		if err := rows.Scan(&run.ID, &run.Pipeline, &run.Status, &run.StopReason, &run.RecordsSeen, &run.RecordsChanged, &run.StartedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func loadMotions(ctx context.Context, pool *pgxpool.Pool, options motionListOptions) ([]motion, int, error) {
	var search *string
	if options.Search != "" {
		search = &options.Search
	}

	rows, err := pool.Query(ctx, `
		WITH subset AS (
			SELECT m.motion_key,
			       m.source_id,
			       m.number,
			       m.title,
			       m.subject,
			       m.status,
			       m.kind,
			       m.parliamentary_year,
			       m.proposed_at,
			       m.source_updated_at,
			       m.source_deleted,
			       m.votes_synced_at,
			       COALESCE(count(DISTINCT d.decision_key), 0)::int AS decision_count,
			       COALESCE(count(v.vote_key) FILTER (WHERE v.source_deleted = false), 0)::int AS vote_count
			FROM motions m
			LEFT JOIN decisions d ON d.motion_key = m.motion_key AND d.source_deleted = false
			LEFT JOIN votes v ON v.motion_key = m.motion_key AND v.source_deleted = false
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND (
			    $2::text IS NULL
			    OR m.title ILIKE '%' || $2 || '%'
			    OR m.subject ILIKE '%' || $2 || '%'
			    OR m.number ILIKE '%' || $2 || '%'
			  )
			GROUP BY m.motion_key
			HAVING $5::boolean = false OR COALESCE(count(v.vote_key) FILTER (WHERE v.source_deleted = false), 0) > 0
		)
		SELECT *, count(*) OVER ()::int AS total
		FROM subset
		ORDER BY proposed_at DESC NULLS LAST, source_updated_at DESC NULLS LAST
		LIMIT $3
		OFFSET $4
	`, options.Jurisdiction, search, options.Limit, options.Offset, options.WithVotes)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	motions := []motion{}
	total := 0
	for rows.Next() {
		var row motion
		if err := scanMotion(rows.Scan, &row); err != nil {
			return nil, 0, err
		}
		total = row.Total
		motions = append(motions, row)
	}
	return motions, total, rows.Err()
}

func loadMotion(ctx context.Context, pool *pgxpool.Pool, motionKey string) (motion, error) {
	var motion motion
	err := pool.QueryRow(ctx, `
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
		       votes_synced_at,
		       (SELECT count(*)::int FROM decisions d WHERE d.motion_key = motions.motion_key AND d.source_deleted = false) AS decision_count,
		       (SELECT count(*)::int FROM votes v WHERE v.motion_key = motions.motion_key AND v.source_deleted = false) AS vote_count,
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
		&motion.VotesSyncedAt,
		&motion.DecisionCount,
		&motion.VoteCount,
		&motion.Total,
	)
	return motion, err
}

func loadDecisions(ctx context.Context, pool *pgxpool.Pool, motionKey string) ([]decision, error) {
	rows, err := pool.Query(ctx, `
		SELECT d.decision_key,
		       d.source_id,
		       d.decision_type,
		       d.decision_text,
		       d.status,
		       d.decision_order,
		       count(v.vote_key) FILTER (WHERE v.source_deleted = false)::int AS vote_count,
		       count(v.vote_key) FILTER (WHERE v.source_deleted = false AND v.mistake = true)::int AS mistake_count
		FROM decisions d
		LEFT JOIN votes v ON v.decision_key = d.decision_key
		WHERE d.motion_key = $1
		  AND d.source_deleted = false
		GROUP BY d.decision_key
		ORDER BY d.decision_order NULLS LAST, d.source_updated_at NULLS LAST
	`, motionKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decisions := []decision{}
	for rows.Next() {
		var decision decision
		if err := rows.Scan(
			&decision.DecisionKey,
			&decision.SourceID,
			&decision.DecisionType,
			&decision.DecisionText,
			&decision.Status,
			&decision.DecisionOrder,
			&decision.VoteCount,
			&decision.MistakeCount,
		); err != nil {
			return nil, err
		}
		decisions = append(decisions, decision)
	}
	return decisions, rows.Err()
}

func loadPartyPositions(ctx context.Context, pool *pgxpool.Pool, motionKey string) ([]partyPosition, error) {
	rows, err := pool.Query(ctx, `
		SELECT COALESCE(party_name, actor_name, party_source_id, 'unknown') AS party_name,
		       party_source_id,
		       SUM(CASE WHEN vote_type = 'Voor' THEN 1 ELSE 0 END)::int AS votes_for,
		       SUM(CASE WHEN vote_type = 'Tegen' THEN 1 ELSE 0 END)::int AS votes_against,
		       COUNT(*)::int AS total_votes
		FROM votes
		WHERE motion_key = $1
		  AND source_deleted = false
		  AND mistake = false
		  AND vote_type IN ('Voor', 'Tegen')
		GROUP BY COALESCE(party_name, actor_name, party_source_id, 'unknown'), party_source_id
		ORDER BY party_name
	`, motionKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []partyPosition{}
	for rows.Next() {
		var position partyPosition
		if err := rows.Scan(&position.PartyName, &position.PartySourceID, &position.VotesFor, &position.VotesAgainst, &position.TotalVotes); err != nil {
			return nil, err
		}
		position.Position = string(politics.PartyPosition(position.VotesFor, position.VotesAgainst))
		positions = append(positions, position)
	}
	return positions, rows.Err()
}

type scanner func(dest ...any) error

func scanMotion(scan scanner, motion *motion) error {
	return scan(
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
		&motion.VotesSyncedAt,
		&motion.DecisionCount,
		&motion.VoteCount,
		&motion.Total,
	)
}

type homePage struct {
	Summary status.Summary
	Runs    []ingestionRun
}

type motionsPage struct {
	Motions   []motion
	Total     int
	Limit     int
	Offset    int
	Search    string
	WithVotes bool
	PrevURL   string
	NextURL   string
}

type motionPage struct {
	Motion    motion
	Decisions []decision
	Positions []partyPosition
}

type partyLikenessPage struct {
	Parties   []likenessParty
	Rows      []analysis.PartyLikeness
	Matrix    map[string]map[string]analysis.PartyLikeness
	DateFrom  string
	DateTo    string
	MinCommon int
}

type likenessParty struct {
	SourceID  string
	ShortName string
}

type ingestionRun struct {
	ID             int64
	Pipeline       string
	Status         string
	StopReason     string
	RecordsSeen    int
	RecordsChanged int
	StartedAt      time.Time
}

type motionListOptions struct {
	Jurisdiction string
	Search       string
	WithVotes    bool
	Limit        int
	Offset       int
}

type motion struct {
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
	VotesSyncedAt     *time.Time
	DecisionCount     int
	VoteCount         int
	Total             int
}

type decision struct {
	DecisionKey   string
	SourceID      string
	DecisionType  *string
	DecisionText  *string
	Status        *string
	DecisionOrder *int
	VoteCount     int
	MistakeCount  int
}

type partyPosition struct {
	PartyName     string
	PartySourceID *string
	Position      string
	VotesFor      int
	VotesAgainst  int
	TotalVotes    int
}

func motionsURL(search string, withVotes bool, limit int, offset int) string {
	query := url.Values{}
	if search != "" {
		query.Set("search", search)
	}
	if withVotes {
		query.Set("withVotes", "true")
	}
	if limit != 25 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	if encoded := query.Encode(); encoded != "" {
		return "/motions?" + encoded
	}
	return "/motions"
}

func likenessParties(rows []analysis.PartyLikeness) []likenessParty {
	seen := map[string]string{}
	for _, row := range rows {
		seen[row.Party1SourceID] = row.Party1Name
		seen[row.Party2SourceID] = row.Party2Name
	}

	parties := make([]likenessParty, 0, len(seen))
	for sourceID, name := range seen {
		parties = append(parties, likenessParty{
			SourceID:  sourceID,
			ShortName: name,
		})
	}
	sort.Slice(parties, func(i, j int) bool {
		return strings.ToLower(parties[i].ShortName) < strings.ToLower(parties[j].ShortName)
	})
	return parties
}

func likenessMatrix(rows []analysis.PartyLikeness) map[string]map[string]analysis.PartyLikeness {
	matrix := map[string]map[string]analysis.PartyLikeness{}
	for _, row := range rows {
		if matrix[row.Party1SourceID] == nil {
			matrix[row.Party1SourceID] = map[string]analysis.PartyLikeness{}
		}
		if matrix[row.Party2SourceID] == nil {
			matrix[row.Party2SourceID] = map[string]analysis.PartyLikeness{}
		}
		matrix[row.Party1SourceID][row.Party2SourceID] = row
		matrix[row.Party2SourceID][row.Party1SourceID] = analysis.PartyLikeness{
			Party1SourceID: row.Party2SourceID,
			Party1Name:     row.Party2Name,
			Party2SourceID: row.Party1SourceID,
			Party2Name:     row.Party1Name,
			CommonMotions:  row.CommonMotions,
			SameVotes:      row.SameVotes,
			Similarity:     row.Similarity,
		}
	}
	return matrix
}

func likenessValue(matrix map[string]map[string]analysis.PartyLikeness, rowID string, columnID string) string {
	if rowID == columnID {
		return "-"
	}
	row := matrix[rowID]
	if row == nil {
		return ""
	}
	cell, ok := row[columnID]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%.0f%%", cell.Similarity)
}

func fallback(values ...any) string {
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(typed) != "" {
				return typed
			}
		case *string:
			if typed != nil && strings.TrimSpace(*typed) != "" {
				return *typed
			}
		default:
			text := fmt.Sprint(value)
			if strings.TrimSpace(text) != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func dateValue(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func timeValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case time.Time:
		return typed.Format("2006-01-02 15:04")
	case *time.Time:
		if typed == nil {
			return ""
		}
		return typed.Format("2006-01-02 15:04")
	default:
		return fmt.Sprint(value)
	}
}

func parseDate(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func writeError(response http.ResponseWriter, err error) {
	fmt.Println(err)
	http.Error(response, "internal server error", http.StatusInternalServerError)
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
