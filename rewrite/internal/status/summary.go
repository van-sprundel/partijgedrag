package status

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Summary struct {
	Motions          int64 `json:"motions"`
	MotionsWithVotes int64 `json:"motionsWithVotes"`
	Decisions        int64 `json:"decisions"`
	Votes            int64 `json:"votes"`
	RawRecords       int64 `json:"rawRecords"`
}

func LoadSummary(ctx context.Context, pool *pgxpool.Pool) (Summary, error) {
	var summary Summary
	err := pool.QueryRow(ctx, `
		SELECT
		  (SELECT count(*) FROM motions WHERE source_deleted = false),
		  (SELECT count(*) FROM motions WHERE source_deleted = false AND votes_synced_at IS NOT NULL),
		  (SELECT count(*) FROM decisions WHERE source_deleted = false),
		  (SELECT count(*) FROM votes WHERE source_deleted = false),
		  (SELECT count(*) FROM raw_records)
	`).Scan(
		&summary.Motions,
		&summary.MotionsWithVotes,
		&summary.Decisions,
		&summary.Votes,
		&summary.RawRecords,
	)
	return summary, err
}
