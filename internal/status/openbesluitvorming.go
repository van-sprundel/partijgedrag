package status

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/analysis"
)

type OBVReport struct {
	SourceKey       string
	JurisdictionKey string
	Label           string

	Motions             int64
	Moties              int64
	Amendments          int64
	MotionsWithVotes    int64
	MotiesWithVotes     int64
	MotiesWithDocument  int64
	TextsFetched        int64
	MotiesWithBullets   int64
	MotiesWithAsk       int64
	MotiesWithVerzoekt  int64
	MotiesStatementOnly int64
	FirstVote, LastVote *time.Time

	Votes           int64
	Members         int64
	UnattributedBy  []NameCount
	OutcomeMismatch []OutcomeCount

	Parties []OBVParty
}

type NameCount struct {
	Name  string
	Count int64
}

type OutcomeCount struct {
	Derived  string
	Upstream string
	Status   string
	Count    int64
}

type OBVParty struct {
	SourceID   string
	Name       string
	Seats      int
	ActiveFrom *time.Time
	ActiveTo   *time.Time
	Votes      int64
	Spellings  []NameCount
}

func LoadOBVReport(ctx context.Context, pool *pgxpool.Pool, sourceKey string) (OBVReport, error) {
	report := OBVReport{SourceKey: sourceKey}

	err := pool.QueryRow(ctx, `
		SELECT d.jurisdiction_key, j.name
		FROM data_sources d JOIN jurisdictions j ON j.jurisdiction_key = d.jurisdiction_key
		WHERE d.source_key = $1
	`, sourceKey).Scan(&report.JurisdictionKey, &report.Label)
	if err != nil {
		return report, err
	}

	err = pool.QueryRow(ctx, `
		WITH m AS (
			SELECT m.*,
			       EXISTS (SELECT 1 FROM votes v WHERE v.motion_key = m.motion_key AND v.source_deleted = false) AS has_votes,
			       COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(r.payload->'attachment') = 'array' THEN r.payload->'attachment' END), 0) > 0 AS has_document,
			       EXISTS (SELECT 1 FROM raw_records t WHERE t.source_key = m.source_key AND t.collection = 'MotionText' AND t.source_id = m.source_id) AS has_text,
			       EXISTS (SELECT 1 FROM jsonb_array_elements_text(COALESCE(m.bullet_points, '[]'::jsonb)) b
			               WHERE b ~* $2) AS has_ask,
			       EXISTS (SELECT 1 FROM jsonb_array_elements_text(COALESCE(m.bullet_points, '[]'::jsonb)) b
			               WHERE b ~* '^spreekt') AS has_statement,
			       EXISTS (SELECT 1 FROM jsonb_array_elements_text(COALESCE(m.bullet_points, '[]'::jsonb)) b
			               WHERE b ILIKE 'verzoekt%') AS has_verzoekt
			FROM motions m
			LEFT JOIN raw_records r ON r.source_key = m.source_key AND r.collection = 'Motion' AND r.source_id = m.source_id
			WHERE m.source_key = $1 AND m.source_deleted = false
		)
		SELECT count(*),
		       count(*) FILTER (WHERE kind = 'Motie'),
		       count(*) FILTER (WHERE kind = 'Amendement'),
		       count(*) FILTER (WHERE has_votes),
		       count(*) FILTER (WHERE kind = 'Motie' AND has_votes),
		       count(*) FILTER (WHERE kind = 'Motie' AND has_votes AND has_document),
		       count(*) FILTER (WHERE has_text),
		       count(*) FILTER (WHERE kind = 'Motie' AND jsonb_array_length(COALESCE(bullet_points, '[]'::jsonb)) > 0),
		       count(*) FILTER (WHERE kind = 'Motie' AND has_ask),
		       count(*) FILTER (WHERE kind = 'Motie' AND has_verzoekt),
		       count(*) FILTER (WHERE kind = 'Motie' AND has_statement AND NOT has_ask),
		       min(proposed_at) FILTER (WHERE has_votes),
		       max(proposed_at) FILTER (WHERE has_votes)
		FROM m
	`, sourceKey, analysis.AskBulletPattern).Scan(
		&report.Motions, &report.Moties, &report.Amendments,
		&report.MotionsWithVotes, &report.MotiesWithVotes, &report.MotiesWithDocument,
		&report.TextsFetched, &report.MotiesWithBullets, &report.MotiesWithAsk, &report.MotiesWithVerzoekt, &report.MotiesStatementOnly,
		&report.FirstVote, &report.LastVote,
	)
	if err != nil {
		return report, err
	}

	err = pool.QueryRow(ctx, `
		SELECT count(*), count(DISTINCT person_source_id)
		FROM votes WHERE source_key = $1 AND source_deleted = false
	`, sourceKey).Scan(&report.Votes, &report.Members)
	if err != nil {
		return report, err
	}

	if report.UnattributedBy, err = loadNameCounts(ctx, pool, `
		SELECT COALESCE(party_name, '(geen fractie)'), count(*)
		FROM votes
		WHERE source_key = $1 AND source_deleted = false AND party_source_id IS NULL
		GROUP BY 1 ORDER BY 2 DESC
	`, sourceKey); err != nil {
		return report, err
	}

	rows, err := pool.Query(ctx, `
		SELECT COALESCE(d.decision_type, ''), COALESCE(d.status, ''), COALESCE(d.decision_text, ''), count(*)
		FROM decisions d
		JOIN motions m ON m.motion_key = d.motion_key
		WHERE d.source_key = $1 AND d.source_deleted = false AND m.source_deleted = false
		  AND d.voting_type IS NOT NULL
		  AND COALESCE(d.decision_type, '') IS DISTINCT FROM COALESCE(d.status, '')
		GROUP BY 1, 2, 3
		ORDER BY 4 DESC
		LIMIT 15
	`, sourceKey)
	if err != nil {
		return report, err
	}
	for rows.Next() {
		var item OutcomeCount
		if err := rows.Scan(&item.Derived, &item.Upstream, &item.Status, &item.Count); err != nil {
			rows.Close()
			return report, err
		}
		report.OutcomeMismatch = append(report.OutcomeMismatch, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return report, err
	}

	rows, err = pool.Query(ctx, `
		SELECT p.source_id, COALESCE(p.short_name, ''), COALESCE(p.seats, 0), p.active_from, p.active_to,
		       (SELECT count(*) FROM votes v WHERE v.source_key = p.source_key AND v.party_source_id = p.source_id AND v.source_deleted = false)
		FROM parties p
		WHERE p.source_key = $1 AND p.source_deleted = false
		ORDER BY p.active_to IS NOT NULL, p.seats DESC NULLS LAST, p.short_name
	`, sourceKey)
	if err != nil {
		return report, err
	}
	for rows.Next() {
		var party OBVParty
		if err := rows.Scan(&party.SourceID, &party.Name, &party.Seats, &party.ActiveFrom, &party.ActiveTo, &party.Votes); err != nil {
			rows.Close()
			return report, err
		}
		report.Parties = append(report.Parties, party)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return report, err
	}

	for i := range report.Parties {
		spellings, err := loadNameCounts(ctx, pool, `
			SELECT party_name, count(*)
			FROM votes
			WHERE source_key = $1 AND party_source_id = $2 AND source_deleted = false
			GROUP BY 1 ORDER BY 2 DESC
		`, sourceKey, report.Parties[i].SourceID)
		if err != nil {
			return report, err
		}
		report.Parties[i].Spellings = spellings
	}

	return report, nil
}

func loadNameCounts(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) ([]NameCount, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var counts []NameCount
	for rows.Next() {
		var item NameCount
		if err := rows.Scan(&item.Name, &item.Count); err != nil {
			return nil, err
		}
		counts = append(counts, item)
	}
	return counts, rows.Err()
}
