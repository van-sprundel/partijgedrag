package analysis

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PartyFocusOptions struct {
	Jurisdiction  string
	PartySourceID string
	DateFrom      *time.Time
	DateTo        *time.Time
	MinCommon     int
}

type PartyFocus struct {
	Party      Party
	Totals     PartyVoteTotals
	Categories []PartyCategoryStats
	Likeness   []PartyLikeness
}

type PartyVoteTotals struct {
	MotionsVoted int
	VotedFor     int
	VotedAgainst int
}

type PartyCategoryStats struct {
	CategoryKey  string
	Name         string
	Kind         string
	MotionsVoted int
	VotedFor     int
	VotedAgainst int
	ForShare     float64
}

func LoadPartyFocus(ctx context.Context, pool *pgxpool.Pool, options PartyFocusOptions) (PartyFocus, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	focus := PartyFocus{}

	err := pool.QueryRow(ctx, `
		SELECT party_key,
		       source_id,
		       COALESCE(short_name, name, source_id) AS short_name,
		       name,
		       seats,
		       active_from::timestamptz,
		       active_to::timestamptz
		FROM parties
		WHERE jurisdiction_key = $1
		  AND source_id = $2
		  AND source_deleted = false
	`, jurisdiction, options.PartySourceID).Scan(
		&focus.Party.PartyKey,
		&focus.Party.SourceID,
		&focus.Party.ShortName,
		&focus.Party.Name,
		&focus.Party.Seats,
		&focus.Party.ActiveFrom,
		&focus.Party.ActiveTo,
	)
	if err != nil {
		return PartyFocus{}, err
	}

	totals, categories, err := loadPartyCategoryStats(ctx, pool, jurisdiction, options)
	if err != nil {
		return PartyFocus{}, err
	}
	focus.Totals = totals
	focus.Categories = categories

	likeness, err := loadPartyLikenessForParty(ctx, pool, jurisdiction, options)
	if err != nil {
		return PartyFocus{}, err
	}
	focus.Likeness = likeness

	return focus, nil
}

// loadPartyLikenessForParty computes likeness only for pairs that include the
// target party. LoadPartyLikeness builds the full NxN matrix of party pairs
// (which is fine for the dedicated likeness page but O(parties²) per motion);
// here we only need a single row out of that matrix, so we anchor the join on
// the target party and avoid materialising every other pair.
func loadPartyLikenessForParty(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, options PartyFocusOptions) ([]PartyLikeness, error) {
	minCommon := options.MinCommon
	if minCommon <= 0 {
		minCommon = 10
	}

	rows, err := pool.Query(ctx, `
		WITH party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       CASE WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END)
			                 > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
			            THEN 'FOR' ELSE 'AGAINST' END AS position
			FROM votes v
			JOIN motions m ON m.motion_key = v.motion_key
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND v.source_deleted = false
			  AND v.mistake = false
			  AND v.party_source_id IS NOT NULL
			  AND v.vote_type IN ('Voor', 'Tegen')
			  AND ($3::timestamptz IS NULL OR m.proposed_at >= $3)
			  AND ($4::timestamptz IS NULL OR m.proposed_at <= $4)
			GROUP BY v.motion_key, v.party_source_id
			HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END)
			    <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
		)
		SELECT $2::text AS party1_source_id,
		       COALESCE(party1.short_name, $2::text) AS party1_name,
		       pp2.party_source_id AS party2_source_id,
		       COALESCE(party2.short_name, pp2.party_source_id) AS party2_name,
		       COUNT(*)::int AS common_motions,
		       SUM(CASE WHEN pp1.position = pp2.position THEN 1 ELSE 0 END)::int AS same_votes,
		       ROUND((SUM(CASE WHEN pp1.position = pp2.position THEN 1 ELSE 0 END)::numeric / COUNT(*)::numeric) * 100, 2)::float8 AS similarity
		FROM party_positions pp1
		JOIN party_positions pp2 ON pp2.motion_key = pp1.motion_key
		                        AND pp2.party_source_id <> $2
		LEFT JOIN parties party1 ON party1.source_key = 'tweedekamer-odata-v2'
		                         AND party1.source_id = $2
		LEFT JOIN parties party2 ON party2.source_key = 'tweedekamer-odata-v2'
		                         AND party2.source_id = pp2.party_source_id
		WHERE pp1.party_source_id = $2
		GROUP BY pp2.party_source_id,
		         COALESCE(party1.short_name, $2::text),
		         COALESCE(party2.short_name, pp2.party_source_id)
		HAVING COUNT(*) >= $5
		ORDER BY similarity DESC, common_motions DESC, party2_name
	`, jurisdiction, options.PartySourceID, options.DateFrom, options.DateTo, minCommon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []PartyLikeness{}
	for rows.Next() {
		var row PartyLikeness
		if err := rows.Scan(&row.Party1SourceID, &row.Party1Name, &row.Party2SourceID, &row.Party2Name, &row.CommonMotions, &row.SameVotes, &row.Similarity); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// partyPositionsCTE classifies each motion the party cast a clear (non-tied)
// Voor/Tegen majority on, within the jurisdiction and optional date range.
const partyPositionsCTE = `
	SELECT v.motion_key,
	       CASE
	         WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) THEN 'FOR'
	         ELSE 'AGAINST'
	       END AS position
	FROM votes v
	JOIN motions m ON m.motion_key = v.motion_key
	WHERE m.jurisdiction_key = $1
	  AND m.source_deleted = false
	  AND v.source_deleted = false
	  AND v.mistake = false
	  AND v.party_source_id = $2
	  AND v.vote_type IN ('Voor', 'Tegen')
	  AND ($3::timestamptz IS NULL OR m.proposed_at >= $3)
	  AND ($4::timestamptz IS NULL OR m.proposed_at <= $4)
	GROUP BY v.motion_key
	HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
`

func loadPartyCategoryStats(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, options PartyFocusOptions) (PartyVoteTotals, []PartyCategoryStats, error) {
	totals := PartyVoteTotals{}
	err := pool.QueryRow(ctx, `
		WITH party_positions AS (`+partyPositionsCTE+`)
		SELECT COUNT(*)::int AS motions_voted,
		       COALESCE(SUM(CASE WHEN position = 'FOR' THEN 1 ELSE 0 END), 0)::int AS voted_for,
		       COALESCE(SUM(CASE WHEN position = 'AGAINST' THEN 1 ELSE 0 END), 0)::int AS voted_against
		FROM party_positions
	`, jurisdiction, options.PartySourceID, options.DateFrom, options.DateTo).Scan(
		&totals.MotionsVoted,
		&totals.VotedFor,
		&totals.VotedAgainst,
	)
	if err != nil {
		return PartyVoteTotals{}, nil, err
	}

	rows, err := pool.Query(ctx, `
		WITH party_positions AS (`+partyPositionsCTE+`)
		SELECT c.category_key,
		       c.name,
		       c.kind,
		       COUNT(*)::int AS motions_voted,
		       SUM(CASE WHEN pp.position = 'FOR' THEN 1 ELSE 0 END)::int AS voted_for,
		       SUM(CASE WHEN pp.position = 'AGAINST' THEN 1 ELSE 0 END)::int AS voted_against
		FROM party_positions pp
		JOIN motion_categories mc ON mc.motion_key = pp.motion_key
		JOIN categories c ON c.category_key = mc.category_key
		GROUP BY c.category_key, c.name, c.kind
		ORDER BY COUNT(*) DESC, c.name
	`, jurisdiction, options.PartySourceID, options.DateFrom, options.DateTo)
	if err != nil {
		return PartyVoteTotals{}, nil, err
	}
	defer rows.Close()

	categories := []PartyCategoryStats{}
	for rows.Next() {
		var stats PartyCategoryStats
		if err := rows.Scan(&stats.CategoryKey, &stats.Name, &stats.Kind, &stats.MotionsVoted, &stats.VotedFor, &stats.VotedAgainst); err != nil {
			return PartyVoteTotals{}, nil, err
		}
		if stats.MotionsVoted > 0 {
			stats.ForShare = (float64(stats.VotedFor) / float64(stats.MotionsVoted)) * 100
		}
		categories = append(categories, stats)
	}
	return totals, categories, rows.Err()
}
