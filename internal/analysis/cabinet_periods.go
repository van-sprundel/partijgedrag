package analysis

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CabinetPeriod struct {
	PeriodKey    string
	Jurisdiction string
	Name         string
	StartedOn    time.Time
	EndedOn      *time.Time
	Parties      []string
}

func LoadCabinetPeriods(ctx context.Context, pool *pgxpool.Pool, jurisdiction string) ([]CabinetPeriod, error) {
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	rows, err := pool.Query(ctx, `
		SELECT period_key,
		       jurisdiction_key,
		       name,
		       started_on::timestamptz,
		       ended_on::timestamptz,
		       parties
		FROM cabinet_periods
		WHERE jurisdiction_key = $1
		ORDER BY started_on DESC
	`, jurisdiction)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	periods := []CabinetPeriod{}
	for rows.Next() {
		var period CabinetPeriod
		if err := rows.Scan(&period.PeriodKey, &period.Jurisdiction, &period.Name, &period.StartedOn, &period.EndedOn, &period.Parties); err != nil {
			return nil, err
		}
		periods = append(periods, period)
	}
	return periods, rows.Err()
}

func LoadCabinetPeriod(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, periodKey string) (CabinetPeriod, error) {
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	var period CabinetPeriod
	err := pool.QueryRow(ctx, `
		SELECT period_key,
		       jurisdiction_key,
		       name,
		       started_on::timestamptz,
		       ended_on::timestamptz,
		       parties
		FROM cabinet_periods
		WHERE jurisdiction_key = $1
		  AND period_key = $2
	`, jurisdiction, periodKey).Scan(&period.PeriodKey, &period.Jurisdiction, &period.Name, &period.StartedOn, &period.EndedOn, &period.Parties)
	return period, err
}

func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
