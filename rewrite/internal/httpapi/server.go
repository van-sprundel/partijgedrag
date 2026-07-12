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

	"partijgedrag/rewrite/internal/analysis"
	"partijgedrag/rewrite/internal/categorize"
	"partijgedrag/rewrite/internal/politics"
	"partijgedrag/rewrite/internal/status"
	"partijgedrag/rewrite/internal/web"
)

const shutdownTimeout = 10 * time.Second

type Server struct {
	Pool *pgxpool.Pool
}

func (server Server) Handler() http.Handler {
	mux := http.NewServeMux()
	web.MustNew(server.Pool).Register(mux)
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("GET /api/summary", server.summary)
	mux.HandleFunc("GET /api/cabinet-periods", server.listCabinetPeriods)
	mux.HandleFunc("GET /api/coalition-analysis", server.getCoalitionAnalysis)
	mux.HandleFunc("GET /api/coalition-analysis/motions", server.listCoalitionMotions)
	mux.HandleFunc("GET /api/ingestion-runs", server.listIngestionRuns)
	mux.HandleFunc("GET /api/categories", server.listCategories)
	mux.HandleFunc("GET /api/parties", server.listParties)
	mux.HandleFunc("GET /api/party-likeness", server.listPartyLikeness)
	mux.HandleFunc("GET /api/party-focus", server.getPartyFocus)
	mux.HandleFunc("GET /api/voting-compass/motions", server.listVotingCompassMotions)
	mux.HandleFunc("POST /api/compass-sessions", server.createCompassSession)
	mux.HandleFunc("GET /api/compass-sessions/{sessionKey}", server.getCompassSession)
	mux.HandleFunc("GET /api/motions", server.listMotions)
	mux.HandleFunc("GET /api/motions/{motionKey}/party-positions", server.getMotionPartyPositions)
	mux.HandleFunc("GET /api/motions/{motionKey}", server.getMotion)
	return mux
}

func (server Server) health(response http.ResponseWriter, request *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (server Server) summary(response http.ResponseWriter, request *http.Request) {
	summary, err := status.LoadSummary(request.Context(), server.Pool)
	if err != nil {
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusOK, summary)
}

func (server Server) listIngestionRuns(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	limit := clamp(parseInt(query.Get("limit"), 10), 1, 100)
	pipeline := query.Get("pipeline")
	failedOnly := query.Get("failed") == "true"

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
		WHERE ($1::text = '' OR pipeline = $1)
		  AND ($2::boolean = false OR status = 'failed')
		ORDER BY started_at DESC
		LIMIT $3
	`, pipeline, failedOnly, limit)
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

func (server Server) listCabinetPeriods(response http.ResponseWriter, request *http.Request) {
	jurisdiction := request.URL.Query().Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	periods, err := analysis.LoadCabinetPeriods(request.Context(), server.Pool, jurisdiction)
	if err != nil {
		writeError(response, err)
		return
	}

	items := make([]map[string]any, 0, len(periods))
	for _, period := range periods {
		items = append(items, map[string]any{
			"periodKey":    period.PeriodKey,
			"jurisdiction": period.Jurisdiction,
			"name":         period.Name,
			"startedOn":    period.StartedOn.Format("2006-01-02"),
			"endedOn":      dateString(period.EndedOn),
			"parties":      period.Parties,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"cabinetPeriods": items,
	})
}

func (server Server) listCategories(response http.ResponseWriter, request *http.Request) {
	jurisdiction := request.URL.Query().Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	categories, err := categorize.LoadCategories(request.Context(), server.Pool, jurisdiction)
	if err != nil {
		writeError(response, err)
		return
	}

	items := make([]map[string]any, 0, len(categories))
	for _, category := range categories {
		items = append(items, map[string]any{
			"categoryKey": category.CategoryKey,
			"name":        category.Name,
			"kind":        category.Kind,
			"keywords":    category.Keywords,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"categories": items,
	})
}

func (server Server) listParties(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	activeOnly := query.Get("activeOnly") != "false"
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	parties, err := analysis.LoadParties(request.Context(), server.Pool, analysis.PartyListOptions{
		Jurisdiction: jurisdiction,
		ActiveOnly:   activeOnly,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	items := make([]map[string]any, 0, len(parties))
	for _, party := range parties {
		items = append(items, map[string]any{
			"partyKey":   party.PartyKey,
			"sourceId":   party.SourceID,
			"shortName":  party.ShortName,
			"name":       party.Name,
			"seats":      party.Seats,
			"activeFrom": party.ActiveFrom,
			"activeTo":   party.ActiveTo,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"parties": items,
	})
}

func (server Server) getCoalitionAnalysis(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	minCommon := clamp(parseInt(query.Get("minCommon"), 5), 1, 1000)

	period, err := selectedCabinetPeriod(request.Context(), server.Pool, jurisdiction, query.Get("period"))
	if err != nil {
		if analysis.IsNotFound(err) {
			writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_period"})
			return
		}
		writeError(response, err)
		return
	}

	coalition, err := analysis.LoadCoalitionAnalysis(request.Context(), server.Pool, analysis.CoalitionAnalysisOptions{
		Period:    period,
		MinCommon: minCommon,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	parties := make([]map[string]any, 0, len(coalition.Parties))
	for _, party := range coalition.Parties {
		parties = append(parties, map[string]any{
			"partySourceId":    party.PartySourceID,
			"partyName":        party.PartyName,
			"coalitionParty":   party.CoalitionParty,
			"commonMotions":    party.CommonMotions,
			"withCoalition":    party.WithCoalition,
			"againstCoalition": party.AgainstCoalition,
			"alignment":        party.Alignment,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"period": map[string]any{
			"periodKey":    period.PeriodKey,
			"jurisdiction": period.Jurisdiction,
			"name":         period.Name,
			"startedOn":    period.StartedOn.Format("2006-01-02"),
			"endedOn":      dateString(period.EndedOn),
			"parties":      period.Parties,
		},
		"minCommon": minCommon,
		"summary": map[string]any{
			"motionsWithCoalitionVotes": coalition.Summary.MotionsWithCoalitionVotes,
			"clearBlocPosition":         coalition.Summary.ClearBlocPosition,
			"unanimousFor":              coalition.Summary.UnanimousFor,
			"unanimousAgainst":          coalition.Summary.UnanimousAgainst,
			"split":                     coalition.Summary.Split,
		},
		"parties": parties,
	})
}

func (server Server) listCoalitionMotions(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	partySourceID := query.Get("partySourceId")
	if partySourceID == "" {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "missing_party_source_id"})
		return
	}
	relation, ok := analysis.NormalizeCoalitionRelation(query.Get("relation"))
	if !ok {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_relation"})
		return
	}

	period, err := selectedCabinetPeriod(request.Context(), server.Pool, jurisdiction, query.Get("period"))
	if err != nil {
		if analysis.IsNotFound(err) {
			writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_period"})
			return
		}
		writeError(response, err)
		return
	}

	limit := clamp(parseInt(query.Get("limit"), 100), 1, 500)
	offset := max(parseInt(query.Get("offset"), 0), 0)
	motions, err := analysis.LoadCoalitionMotions(request.Context(), server.Pool, analysis.CoalitionMotionOptions{
		Period:        period,
		PartySourceID: partySourceID,
		Relation:      relation,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	items := make([]map[string]any, 0, len(motions))
	for _, motion := range motions {
		items = append(items, map[string]any{
			"motionKey":         motion.MotionKey,
			"number":            motion.Number,
			"title":             motion.Title,
			"subject":           motion.Subject,
			"proposedAt":        motion.ProposedAt,
			"partySourceId":     motion.PartySourceID,
			"partyName":         motion.PartyName,
			"partyPosition":     motion.PartyPosition,
			"coalitionPosition": motion.CoalitionPosition,
			"coalitionFor":      motion.CoalitionFor,
			"coalitionAgainst":  motion.CoalitionAgainst,
			"withCoalition":     motion.WithCoalition,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"period": map[string]any{
			"periodKey":    period.PeriodKey,
			"jurisdiction": period.Jurisdiction,
			"name":         period.Name,
			"startedOn":    period.StartedOn.Format("2006-01-02"),
			"endedOn":      dateString(period.EndedOn),
			"parties":      period.Parties,
		},
		"partySourceId": partySourceID,
		"relation":      relation,
		"limit":         limit,
		"offset":        offset,
		"motions":       items,
	})
}

func (server Server) listPartyLikeness(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	dateFrom, err := parseDate(query.Get("dateFrom"))
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_date_from"})
		return
	}
	dateTo, err := parseDate(query.Get("dateTo"))
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_date_to"})
		return
	}
	minCommon := clamp(parseInt(query.Get("minCommon"), 10), 1, 1000)
	periodKey := query.Get("period")
	if periodKey != "" {
		period, err := analysis.LoadCabinetPeriod(request.Context(), server.Pool, jurisdiction, periodKey)
		if err != nil {
			if analysis.IsNotFound(err) {
				writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_period"})
				return
			}
			writeError(response, err)
			return
		}
		dateFrom = &period.StartedOn
		dateTo = period.EndedOn
	}

	rows, err := analysis.LoadPartyLikeness(request.Context(), server.Pool, analysis.PartyLikenessOptions{
		Jurisdiction: jurisdiction,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		MinCommon:    minCommon,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]any{
			"party1SourceId": row.Party1SourceID,
			"party1Name":     row.Party1Name,
			"party2SourceId": row.Party2SourceID,
			"party2Name":     row.Party2Name,
			"commonMotions":  row.CommonMotions,
			"sameVotes":      row.SameVotes,
			"similarity":     row.Similarity,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"partyLikeness": items,
		"minCommon":     minCommon,
		"period":        periodKey,
		"dateFrom":      dateString(dateFrom),
		"dateTo":        dateString(dateTo),
	})
}

func (server Server) getPartyFocus(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	partySourceID := query.Get("party")
	if partySourceID == "" {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "missing_party"})
		return
	}

	dateFrom, err := parseDate(query.Get("dateFrom"))
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_date_from"})
		return
	}
	dateTo, err := parseDate(query.Get("dateTo"))
	if err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_date_to"})
		return
	}
	minCommon := clamp(parseInt(query.Get("minCommon"), 10), 1, 1000)
	periodKey := query.Get("period")
	if periodKey != "" {
		period, err := analysis.LoadCabinetPeriod(request.Context(), server.Pool, jurisdiction, periodKey)
		if err != nil {
			if analysis.IsNotFound(err) {
				writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_period"})
				return
			}
			writeError(response, err)
			return
		}
		dateFrom = &period.StartedOn
		dateTo = period.EndedOn
	}

	focus, err := analysis.LoadPartyFocus(request.Context(), server.Pool, analysis.PartyFocusOptions{
		Jurisdiction:  jurisdiction,
		PartySourceID: partySourceID,
		DateFrom:      dateFrom,
		DateTo:        dateTo,
		MinCommon:     minCommon,
	})
	if err != nil {
		if analysis.IsNotFound(err) {
			writeJSON(response, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeError(response, err)
		return
	}

	categories := make([]map[string]any, 0, len(focus.Categories))
	for _, category := range focus.Categories {
		categories = append(categories, map[string]any{
			"categoryKey":  category.CategoryKey,
			"name":         category.Name,
			"kind":         category.Kind,
			"motionsVoted": category.MotionsVoted,
			"votedFor":     category.VotedFor,
			"votedAgainst": category.VotedAgainst,
			"forShare":     category.ForShare,
		})
	}
	likeness := make([]map[string]any, 0, len(focus.Likeness))
	for _, row := range focus.Likeness {
		likeness = append(likeness, map[string]any{
			"partySourceId": row.Party2SourceID,
			"partyName":     row.Party2Name,
			"commonMotions": row.CommonMotions,
			"sameVotes":     row.SameVotes,
			"similarity":    row.Similarity,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"party": map[string]any{
			"partyKey":   focus.Party.PartyKey,
			"sourceId":   focus.Party.SourceID,
			"shortName":  focus.Party.ShortName,
			"name":       focus.Party.Name,
			"seats":      focus.Party.Seats,
			"activeFrom": focus.Party.ActiveFrom,
			"activeTo":   focus.Party.ActiveTo,
		},
		"totals": map[string]any{
			"motionsVoted": focus.Totals.MotionsVoted,
			"votedFor":     focus.Totals.VotedFor,
			"votedAgainst": focus.Totals.VotedAgainst,
		},
		"categories": categories,
		"likeness":   likeness,
		"minCommon":  minCommon,
		"period":     periodKey,
		"dateFrom":   dateString(dateFrom),
		"dateTo":     dateString(dateTo),
	})
}

func (server Server) listVotingCompassMotions(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	jurisdiction := query.Get("jurisdiction")
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	var dateFrom *time.Time
	var dateTo *time.Time
	periodKey := query.Get("period")
	if periodKey != "" {
		period, err := analysis.LoadCabinetPeriod(request.Context(), server.Pool, jurisdiction, periodKey)
		if err != nil {
			if analysis.IsNotFound(err) {
				writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_period"})
				return
			}
			writeError(response, err)
			return
		}
		dateFrom = &period.StartedOn
		dateTo = period.EndedOn
	}

	limit := clamp(parseInt(query.Get("limit"), 12), 1, 50)
	minParties := clamp(parseInt(query.Get("minParties"), 8), 1, 50)
	motions, err := analysis.LoadVotingCompassMotions(request.Context(), server.Pool, analysis.VotingCompassOptions{
		Jurisdiction: jurisdiction,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		Limit:        limit,
		MinParties:   minParties,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	items := make([]map[string]any, 0, len(motions))
	for _, motion := range motions {
		positions := make([]map[string]any, 0, len(motion.Positions))
		for _, position := range motion.Positions {
			positions = append(positions, map[string]any{
				"partySourceId": position.PartySourceID,
				"partyName":     position.PartyName,
				"position":      position.Position,
			})
		}
		items = append(items, map[string]any{
			"motionKey":  motion.MotionKey,
			"number":     motion.Number,
			"title":      motion.Title,
			"subject":    motion.Subject,
			"proposedAt": motion.ProposedAt,
			"positions":  positions,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"period":     periodKey,
		"limit":      limit,
		"minParties": minParties,
		"motions":    items,
	})
}

func (server Server) createCompassSession(response http.ResponseWriter, request *http.Request) {
	var input struct {
		Jurisdiction string                   `json:"jurisdiction"`
		Answers      []analysis.CompassAnswer `json:"answers"`
		MinOverlap   int                      `json:"minOverlap"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(response, request.Body, 64*1024)).Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	if input.MinOverlap == 0 {
		input.MinOverlap = 5
	}
	if err := analysis.ValidateCompassAnswers(input.Answers); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{
			"error":  "invalid_answers",
			"detail": err.Error(),
		})
		return
	}

	sessionKey, err := analysis.SaveCompassSession(request.Context(), server.Pool, input.Jurisdiction, input.Answers, input.MinOverlap)
	if err != nil {
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusCreated, map[string]any{
		"sessionKey": sessionKey,
		"url":        "/compass/results/" + sessionKey,
	})
}

func (server Server) getCompassSession(response http.ResponseWriter, request *http.Request) {
	sessionKey := request.PathValue("sessionKey")

	session, err := analysis.LoadCompassSession(request.Context(), server.Pool, sessionKey)
	if err != nil {
		if analysis.IsNotFound(err) {
			writeJSON(response, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeError(response, err)
		return
	}
	results, err := analysis.ScoreCompassSession(request.Context(), server.Pool, session)
	if err != nil {
		writeError(response, err)
		return
	}

	matches := make([]map[string]any, 0, len(results.Matches))
	for _, match := range results.Matches {
		matches = append(matches, compassMatchValue(match))
	}
	inconclusive := make([]map[string]any, 0, len(results.Inconclusive))
	for _, match := range results.Inconclusive {
		inconclusive = append(inconclusive, compassMatchValue(match))
	}
	motions := make([]map[string]any, 0, len(results.Motions))
	for _, motion := range results.Motions {
		positions := make([]map[string]any, 0, len(motion.Positions))
		for _, position := range motion.Positions {
			positions = append(positions, map[string]any{
				"partySourceId":  position.PartySourceID,
				"partyName":      position.PartyName,
				"position":       position.Position,
				"agreesWithUser": position.AgreesWithUser,
			})
		}
		motions = append(motions, map[string]any{
			"motionKey":  motion.Motion.MotionKey,
			"number":     motion.Motion.Number,
			"title":      motion.Motion.Title,
			"subject":    motion.Motion.Subject,
			"proposedAt": motion.Motion.ProposedAt,
			"userAnswer": motion.UserAnswer,
			"positions":  positions,
		})
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"sessionKey":   session.SessionKey,
		"createdAt":    session.CreatedAt,
		"totalAnswers": len(session.Answers),
		"threshold":    results.Threshold,
		"matches":      matches,
		"inconclusive": inconclusive,
		"motions":      motions,
	})
}

func compassMatchValue(match analysis.CompassMatch) map[string]any {
	return map[string]any{
		"partySourceId": match.PartySourceID,
		"partyName":     match.PartyName,
		"match":         match.Match,
		"sameVotes":     match.SameVotes,
		"overlap":       match.Overlap,
	}
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
	withVotes := query.Get("withVotes") == "true"
	category := query.Get("category")

	rows, err := server.Pool.Query(ctx, `
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
			  AND (
			    $6::text = ''
			    OR EXISTS (
			      SELECT 1
			      FROM motion_categories mc
			      WHERE mc.motion_key = m.motion_key
			        AND mc.category_key = $6
			    )
			  )
			GROUP BY m.motion_key
			HAVING $5::boolean = false OR COALESCE(count(v.vote_key) FILTER (WHERE v.source_deleted = false), 0) > 0
		)
		SELECT *, count(*) OVER ()::int AS total
		FROM subset
		ORDER BY proposed_at DESC NULLS LAST, source_updated_at DESC NULLS LAST
		LIMIT $3
		OFFSET $4
	`, jurisdiction, searchPtr, limit, offset, withVotes, category)
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
	if err != nil {
		if err == pgx.ErrNoRows {
			writeJSON(response, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeError(response, err)
		return
	}

	categoryRows, err := server.Pool.Query(request.Context(), `
		SELECT c.category_key, c.name, c.kind
		FROM motion_categories mc
		JOIN categories c ON c.category_key = mc.category_key
		WHERE mc.motion_key = $1
		ORDER BY c.kind, c.name
	`, motionKey)
	if err != nil {
		writeError(response, err)
		return
	}
	defer categoryRows.Close()

	categories := []map[string]any{}
	for categoryRows.Next() {
		var categoryKey, name, kind string
		if err := categoryRows.Scan(&categoryKey, &name, &kind); err != nil {
			writeError(response, err)
			return
		}
		categories = append(categories, map[string]any{
			"categoryKey": categoryKey,
			"name":        name,
			"kind":        kind,
		})
	}
	if err := categoryRows.Err(); err != nil {
		writeError(response, err)
		return
	}

	value := motion.mapValue()
	value["categories"] = categories
	writeJSON(response, http.StatusOK, value)
}

func (server Server) getMotionPartyPositions(response http.ResponseWriter, request *http.Request) {
	motionKey := request.PathValue("motionKey")

	rows, err := server.Pool.Query(request.Context(), `
		SELECT party_source_id,
		       COALESCE(party_name, actor_name, party_source_id, 'unknown') AS party_name,
		       SUM(CASE WHEN vote_type = 'Voor' THEN 1 ELSE 0 END)::int AS votes_for,
		       SUM(CASE WHEN vote_type = 'Tegen' THEN 1 ELSE 0 END)::int AS votes_against,
		       COUNT(*)::int AS total_votes
		FROM votes
		WHERE motion_key = $1
		  AND source_deleted = false
		  AND mistake = false
		  AND vote_type IN ('Voor', 'Tegen')
		GROUP BY party_source_id, COALESCE(party_name, actor_name, party_source_id, 'unknown')
		ORDER BY party_name
	`, motionKey)
	if err != nil {
		writeError(response, err)
		return
	}
	defer rows.Close()

	positions := []map[string]any{}
	for rows.Next() {
		var partySourceID *string
		var partyName string
		var votesFor, votesAgainst, totalVotes int
		if err := rows.Scan(&partySourceID, &partyName, &votesFor, &votesAgainst, &totalVotes); err != nil {
			writeError(response, err)
			return
		}

		positions = append(positions, map[string]any{
			"partySourceId": partySourceID,
			"partyName":     partyName,
			"position":      politics.PartyPosition(votesFor, votesAgainst),
			"votesFor":      votesFor,
			"votesAgainst":  votesAgainst,
			"totalVotes":    totalVotes,
		})
	}
	if err := rows.Err(); err != nil {
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusOK, map[string]any{
		"motionKey":      motionKey,
		"partyPositions": positions,
	})
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
	VotesSyncedAt     *time.Time
	DecisionCount     int
	VoteCount         int
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
		&row.VotesSyncedAt,
		&row.DecisionCount,
		&row.VoteCount,
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
		"votesSyncedAt":     row.VotesSyncedAt,
		"decisionCount":     row.DecisionCount,
		"voteCount":         row.VoteCount,
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

func dateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func selectedCabinetPeriod(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, periodKey string) (analysis.CabinetPeriod, error) {
	if periodKey != "" {
		return analysis.LoadCabinetPeriod(ctx, pool, jurisdiction, periodKey)
	}
	periods, err := analysis.LoadCabinetPeriods(ctx, pool, jurisdiction)
	if err != nil {
		return analysis.CabinetPeriod{}, err
	}
	if len(periods) == 0 {
		return analysis.CabinetPeriod{}, pgx.ErrNoRows
	}
	return periods[0], nil
}

func clamp(value int, minValue int, maxValue int) int {
	return min(max(value, minValue), maxValue)
}
