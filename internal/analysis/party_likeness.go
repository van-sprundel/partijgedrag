package analysis

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
		ORDER BY active_to IS NOT NULL, lower(COALESCE(short_name, name, source_id))
	`, jurisdiction, options.ActiveOnly)
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
	return parties, rows.Err()
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
		)
		SELECT p1.party_source_id,
		       COALESCE(party1.short_name, p1.party_source_id) AS party1_name,
		       p2.party_source_id,
		       COALESCE(party2.short_name, p2.party_source_id) AS party2_name,
		       COUNT(*)::int AS common_motions,
		       SUM(CASE WHEN p1.position = p2.position THEN 1 ELSE 0 END)::int AS same_votes,
		       ROUND((SUM(CASE WHEN p1.position = p2.position THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric) * 100, 2)::float8 AS similarity
		FROM classified p1
		JOIN classified p2 ON p1.motion_key = p2.motion_key
		                  AND p1.party_source_id < p2.party_source_id
		LEFT JOIN parties party1 ON party1.source_key = 'tweedekamer-odata-v2'
		                        AND party1.source_id = p1.party_source_id
		LEFT JOIN parties party2 ON party2.source_key = 'tweedekamer-odata-v2'
		                        AND party2.source_id = p2.party_source_id
		GROUP BY p1.party_source_id,
		         COALESCE(party1.short_name, p1.party_source_id),
		         p2.party_source_id,
		         COALESCE(party2.short_name, p2.party_source_id)
		HAVING COUNT(*) >= $4
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
	return rowsOut, rows.Err()
}
