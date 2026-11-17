package analysis

import (
	"context"
	"sort"
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
