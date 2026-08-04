package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/cache"
)

type Party struct {
	PartyKey   string
	SourceID   string
	ShortName  string
	Name       *string
	Seats      *int
	ActiveFrom *time.Time
	ActiveTo   *time.Time
}

type PartyLikeness struct {
	Party1SourceID string
	Party1Name     string
	Party2SourceID string
	Party2Name     string
	CommonMotions  int
	SameVotes      int
	Similarity     float64
}

type PartyListOptions struct {
	Jurisdiction string
	ActiveOnly   bool
	// ActiveFrom/ActiveTo keep only parties whose activity overlaps the window.
	ActiveFrom *time.Time
	ActiveTo   *time.Time
}

type PartyLikenessOptions struct {
	Jurisdiction string
	DateFrom     *time.Time
	DateTo       *time.Time
	MinCommon    int
}

func LoadParties(ctx context.Context, pool *pgxpool.Pool, options PartyListOptions) ([]Party, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	cacheKey := fmt.Sprintf("analysis:parties:%s:%t:%s:%s", jurisdiction, options.ActiveOnly, formatOptTime(options.ActiveFrom), formatOptTime(options.ActiveTo))
	if cached, ok := cache.Global().Get(cacheKey); ok {
		return copyParties(cached.([]Party)), nil
	}

	rows, err := pool.Query(ctx, `
		SELECT party_key,
		       source_id,
		       COALESCE(short_name, name, source_id) AS short_name,
		       name,
		       seats,
		       active_from::timestamptz,
		       active_to::timestamptz
		FROM parties
		WHERE jurisdiction_key = $1
		  AND source_deleted = false
		  AND ($2::boolean = false OR active_to IS NULL OR active_to >= CURRENT_DATE)
		  AND ($3::timestamptz IS NULL OR active_to IS NULL OR active_to >= $3)
		  AND ($4::timestamptz IS NULL OR active_from IS NULL OR active_from < $4)
		ORDER BY active_to IS NOT NULL, lower(COALESCE(short_name, name, source_id))
	`, jurisdiction, options.ActiveOnly, options.ActiveFrom, options.ActiveTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parties := []Party{}
	for rows.Next() {
		var party Party
		if err := rows.Scan(&party.PartyKey, &party.SourceID, &party.ShortName, &party.Name, &party.Seats, &party.ActiveFrom, &party.ActiveTo); err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cache.Global().Set(cacheKey, copyParties(parties))
	return parties, nil
}

// LoadPartyLogoAvailability reports which parties have a logo on file, keyed by
// source id. Pages that label a party use it to choose between the icon and a
// text fallback without pulling the image bytes into the page query.
func LoadPartyLogoAvailability(ctx context.Context, pool *pgxpool.Pool, jurisdiction string) (map[string]bool, error) {
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	cacheKey := "analysis:party_logo_availability:" + jurisdiction
	if cached, ok := cache.Global().Get(cacheKey); ok {
		return copyMapBool(cached.(map[string]bool)), nil
	}

	rows, err := pool.Query(ctx, `
		SELECT source_id
		FROM parties
		WHERE jurisdiction_key = $1
		  AND logo_data IS NOT NULL
	`, jurisdiction)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	available := map[string]bool{}
	for rows.Next() {
		var sourceID string
		if err := rows.Scan(&sourceID); err != nil {
			return nil, err
		}
		available[sourceID] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cache.Global().Set(cacheKey, copyMapBool(available))
	return available, nil
}

func LoadPartyLikeness(ctx context.Context, pool *pgxpool.Pool, options PartyLikenessOptions) ([]PartyLikeness, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	minCommon := options.MinCommon
	if minCommon <= 0 {
		minCommon = 10
	}

	cacheKey := fmt.Sprintf("analysis:party_likeness:%s:%s:%s:%d", jurisdiction, formatOptTime(options.DateFrom), formatOptTime(options.DateTo), minCommon)
	if cached, ok := cache.Global().Get(cacheKey); ok {
		return copyPartyLikeness(cached.([]PartyLikeness)), nil
	}

	rows, err := pool.Query(ctx, `
		WITH party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END)::int AS votes_for,
			       SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)::int AS votes_against
			FROM votes v
			JOIN motions m ON m.motion_key = v.motion_key
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND v.source_deleted = false
			  AND v.mistake = false
			  AND v.party_source_id IS NOT NULL
			  AND v.vote_type IN ('Voor', 'Tegen')
			  AND ($2::timestamptz IS NULL OR m.proposed_at >= $2)
			  AND ($3::timestamptz IS NULL OR m.proposed_at <= $3)
			GROUP BY v.motion_key, v.party_source_id
		),
		classified AS (
			SELECT motion_key,
			       party_source_id,
			       CASE
			         WHEN votes_for > votes_against THEN 'FOR'
			         WHEN votes_against > votes_for THEN 'AGAINST'
			         ELSE 'NEUTRAL'
			       END AS position
			FROM party_positions
			WHERE votes_for <> votes_against
		),
		pair_stats AS (
			SELECT p1.party_source_id AS party1_source_id,
			       p2.party_source_id AS party2_source_id,
			       COUNT(*)::int AS common_motions,
			       SUM(CASE WHEN p1.position = p2.position THEN 1 ELSE 0 END)::int AS same_votes
			FROM classified p1
			JOIN classified p2 ON p1.motion_key = p2.motion_key
			                  AND p1.party_source_id < p2.party_source_id
			GROUP BY p1.party_source_id, p2.party_source_id
			HAVING COUNT(*) >= $4
		)
		SELECT ps.party1_source_id,
		       COALESCE(party1.short_name, ps.party1_source_id) AS party1_name,
		       ps.party2_source_id,
		       COALESCE(party2.short_name, ps.party2_source_id) AS party2_name,
		       ps.common_motions,
		       ps.same_votes,
		       ROUND((ps.same_votes::numeric / ps.common_motions::numeric) * 100, 2)::float8 AS similarity
		FROM pair_stats ps
		LEFT JOIN parties party1 ON party1.source_key = 'tweedekamer-odata-v2'
		                         AND party1.source_id = ps.party1_source_id
		LEFT JOIN parties party2 ON party2.source_key = 'tweedekamer-odata-v2'
		                         AND party2.source_id = ps.party2_source_id
		ORDER BY similarity DESC, common_motions DESC, party1_name, party2_name
	`, jurisdiction, options.DateFrom, options.DateTo, minCommon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rowsOut := []PartyLikeness{}
	for rows.Next() {
		var row PartyLikeness
		if err := rows.Scan(&row.Party1SourceID, &row.Party1Name, &row.Party2SourceID, &row.Party2Name, &row.CommonMotions, &row.SameVotes, &row.Similarity); err != nil {
			return nil, err
		}
		rowsOut = append(rowsOut, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cache.Global().Set(cacheKey, copyPartyLikeness(rowsOut))
	return rowsOut, nil
}

func formatOptTime(t *time.Time) string {
	if t == nil {
		return "nil"
	}
	return t.Format(time.RFC3339)
}

func copyParties(src []Party) []Party {
	if src == nil {
		return nil
	}
	dst := make([]Party, len(src))
	copy(dst, src)
	return dst
}

func copyPartyLikeness(src []PartyLikeness) []PartyLikeness {
	if src == nil {
		return nil
	}
	dst := make([]PartyLikeness, len(src))
	copy(dst, src)
	return dst
}

func copyMapBool(src map[string]bool) map[string]bool {
	if src == nil {
		return nil
	}
	dst := make(map[string]bool, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
