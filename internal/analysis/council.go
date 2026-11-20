package analysis

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Council struct {
	JurisdictionKey string
	Name            string
	Since           *time.Time
}

const councilSeatShare = 0.9

// councilKeyPrefix prefixes every municipality's jurisdiction key; the rest is
// the slug used in /gemeente/{council} URLs.
const councilKeyPrefix = "nl-gemeente-"

// CouncilSummary describes a council on the /gemeente index.
type CouncilSummary struct {
	Slug             string
	Name             string
	MotionsWithVotes int64
	Fracties         int64
	LastVote         *time.Time
}

// CouncilSlug turns a jurisdiction key into its /gemeente/{council} slug.
func CouncilSlug(jurisdiction string) string {
	return strings.TrimPrefix(jurisdiction, councilKeyPrefix)
}

// ListCouncils returns the councils with at least one voted motion, by name.
// Amendementen are left out of the count: the compass only asks moties.
func ListCouncils(ctx context.Context, pool *pgxpool.Pool) ([]CouncilSummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT j.jurisdiction_key, j.name, m.voted, m.last_vote, p.fracties
		FROM jurisdictions j
		CROSS JOIN LATERAL (
			SELECT count(*) AS voted, max(m.proposed_at) AS last_vote
			FROM motions m
			WHERE m.jurisdiction_key = j.jurisdiction_key
			  AND m.source_deleted = false
			  AND m.kind = 'Motie'
			  AND EXISTS (
			      SELECT 1 FROM votes v
			      WHERE v.motion_key = m.motion_key
			        AND v.source_deleted = false
			        AND v.mistake = false
			  )
		) m
		CROSS JOIN LATERAL (
			SELECT count(*) AS fracties
			FROM parties p
			WHERE p.jurisdiction_key = j.jurisdiction_key
			  AND p.source_deleted = false
			  AND p.active_to IS NULL
			  AND p.active_from IS NOT NULL
		) p
		WHERE j.kind = 'municipality'
		  AND m.voted > 0
		ORDER BY j.name
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (CouncilSummary, error) {
		var council CouncilSummary
		var key string
		err := row.Scan(&key, &council.Name, &council.MotionsWithVotes, &council.LastVote, &council.Fracties)
		council.Slug = CouncilSlug(key)
		return council, err
	})
}

type councilParty struct {
	from  time.Time
	seats int
}

func LoadCouncil(ctx context.Context, pool *pgxpool.Pool, jurisdiction string) (Council, error) {
	council := Council{JurisdictionKey: jurisdiction}
	err := pool.QueryRow(ctx, `
		SELECT name FROM jurisdictions WHERE jurisdiction_key = $1 AND kind = 'municipality'
	`, jurisdiction).Scan(&council.Name)
	if err != nil {
		return council, err
	}

	rows, err := pool.Query(ctx, `
		SELECT active_from::timestamptz, COALESCE(seats, 0)
		FROM parties
		WHERE jurisdiction_key = $1 AND source_deleted = false
		  AND active_to IS NULL AND active_from IS NOT NULL
	`, jurisdiction)
	if err != nil {
		return council, err
	}
	parties, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (councilParty, error) {
		var party councilParty
		return party, row.Scan(&party.from, &party.seats)
	})
	if err != nil {
		return council, err
	}
	council.Since = councilSince(parties)
	return council, nil
}

func councilSince(parties []councilParty) *time.Time {
	if len(parties) == 0 {
		return nil
	}
	sort.Slice(parties, func(i, j int) bool { return parties[i].from.Before(parties[j].from) })
	total := 0
	for _, party := range parties {
		total += party.seats
	}
	covered := 0
	for _, party := range parties {
		covered += party.seats
		if float64(covered) >= councilSeatShare*float64(total) {
			return &party.from
		}
	}
	return &parties[len(parties)-1].from
}
