package analysis

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CoalitionSummary struct {
	MotionsWithCoalitionVotes int
	ClearBlocPosition         int
	UnanimousFor              int
	UnanimousAgainst          int
	Split                     int
}

type CoalitionPartyAlignment struct {
	PartySourceID    *string
	PartyName        string
	CoalitionParty   bool
	CommonMotions    int
	WithCoalition    int
	AgainstCoalition int
	Alignment        float64
}

type CoalitionAnalysis struct {
	Summary CoalitionSummary
	Parties []CoalitionPartyAlignment
}

type CoalitionAnalysisOptions struct {
	Period    CabinetPeriod
	MinCommon int
}

func LoadCoalitionAnalysis(ctx context.Context, pool *pgxpool.Pool, options CoalitionAnalysisOptions) (CoalitionAnalysis, error) {
	minCommon := options.MinCommon
	if minCommon <= 0 {
		minCommon = 5
	}

	coalitionParties := normalizedPartyNames(options.Period.Parties)
	analysis := CoalitionAnalysis{}

	summary, err := loadCoalitionSummary(ctx, pool, options.Period, coalitionParties)
	if err != nil {
		return CoalitionAnalysis{}, err
	}
	analysis.Summary = summary

	parties, err := loadCoalitionPartyAlignment(ctx, pool, options.Period, coalitionParties, minCommon)
	if err != nil {
		return CoalitionAnalysis{}, err
	}
	analysis.Parties = parties

	return analysis, nil
}

func loadCoalitionSummary(ctx context.Context, pool *pgxpool.Pool, period CabinetPeriod, coalitionParties []string) (CoalitionSummary, error) {
	var summary CoalitionSummary
	err := pool.QueryRow(ctx, coalitionPositionSQL()+`
		SELECT COUNT(*)::int AS motions_with_coalition_votes,
		       COUNT(*) FILTER (WHERE coalition_for <> coalition_against)::int AS clear_bloc_position,
		       COUNT(*) FILTER (WHERE coalition_for > 0 AND coalition_against = 0)::int AS unanimous_for,
		       COUNT(*) FILTER (WHERE coalition_against > 0 AND coalition_for = 0)::int AS unanimous_against,
		       COUNT(*) FILTER (WHERE coalition_for > 0 AND coalition_against > 0)::int AS split
		FROM coalition_by_motion
	`, period.Jurisdiction, period.StartedOn, period.EndedOn, coalitionParties).Scan(
		&summary.MotionsWithCoalitionVotes,
		&summary.ClearBlocPosition,
		&summary.UnanimousFor,
		&summary.UnanimousAgainst,
		&summary.Split,
	)
	return summary, err
}

func loadCoalitionPartyAlignment(ctx context.Context, pool *pgxpool.Pool, period CabinetPeriod, coalitionParties []string, minCommon int) ([]CoalitionPartyAlignment, error) {
	rows, err := pool.Query(ctx, coalitionPositionSQL()+`
		SELECT pp.party_source_id,
		       pp.party_name,
		       (upper(pp.party_name) = ANY($4::text[])) AS coalition_party,
		       COUNT(*)::int AS common_motions,
		       COUNT(*) FILTER (
		         WHERE (cbm.coalition_for > cbm.coalition_against AND pp.position = 'FOR')
		            OR (cbm.coalition_against > cbm.coalition_for AND pp.position = 'AGAINST')
		       )::int AS with_coalition,
		       COUNT(*) FILTER (
		         WHERE (cbm.coalition_for > cbm.coalition_against AND pp.position = 'AGAINST')
		            OR (cbm.coalition_against > cbm.coalition_for AND pp.position = 'FOR')
		       )::int AS against_coalition,
		       ROUND((
		         COUNT(*) FILTER (
		           WHERE (cbm.coalition_for > cbm.coalition_against AND pp.position = 'FOR')
		              OR (cbm.coalition_against > cbm.coalition_for AND pp.position = 'AGAINST')
		         )::numeric / COUNT(*)::numeric
		       ) * 100, 2)::float8 AS alignment
		FROM party_positions pp
		JOIN coalition_by_motion cbm ON cbm.motion_key = pp.motion_key
		WHERE cbm.coalition_for <> cbm.coalition_against
		GROUP BY pp.party_source_id, pp.party_name
		HAVING COUNT(*) >= $5
		ORDER BY coalition_party DESC, alignment DESC, common_motions DESC, pp.party_name
	`, period.Jurisdiction, period.StartedOn, period.EndedOn, coalitionParties, minCommon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parties := []CoalitionPartyAlignment{}
	for rows.Next() {
		var party CoalitionPartyAlignment
		if err := rows.Scan(
			&party.PartySourceID,
			&party.PartyName,
			&party.CoalitionParty,
			&party.CommonMotions,
			&party.WithCoalition,
			&party.AgainstCoalition,
			&party.Alignment,
		); err != nil {
			return nil, err
		}
		parties = append(parties, party)
	}
	return parties, rows.Err()
}

func coalitionPositionSQL() string {
	return `
		WITH party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown') AS party_name,
			       CASE
			         WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) THEN 'FOR'
			         WHEN SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) THEN 'AGAINST'
			         ELSE 'NEUTRAL'
			       END AS position
			FROM votes v
			JOIN motions m ON m.motion_key = v.motion_key
			LEFT JOIN parties p ON p.source_key = v.source_key
			                   AND p.source_id = v.party_source_id
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND m.proposed_at >= $2
			  AND ($3::timestamptz IS NULL OR m.proposed_at < $3)
			  AND v.source_deleted = false
			  AND v.mistake = false
			  AND v.vote_type IN ('Voor', 'Tegen')
			GROUP BY v.motion_key,
			         v.party_source_id,
			         COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown')
			HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
		),
		coalition_by_motion AS (
			SELECT motion_key,
			       COUNT(*)::int AS coalition_parties_seen,
			       COUNT(*) FILTER (WHERE position = 'FOR')::int AS coalition_for,
			       COUNT(*) FILTER (WHERE position = 'AGAINST')::int AS coalition_against
			FROM party_positions
			WHERE upper(party_name) = ANY($4::text[])
			GROUP BY motion_key
		)
	`
}

func normalizedPartyNames(names []string) []string {
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.ToUpper(strings.TrimSpace(name))
		if name != "" {
			normalized = append(normalized, name)
		}
	}
	return normalized
}
