package status

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Summary struct {
	Motions                int64 `json:"motions"`
	MotionsWithVotes       int64 `json:"motionsWithVotes"`
	MotionsWithoutVotes    int64 `json:"motionsWithoutVotes"`
	MotionsWithNoDecisions int64 `json:"motionsWithNoDecisions"`
	Decisions              int64 `json:"decisions"`
	DecisionsWithoutVotes  int64 `json:"decisionsWithoutVotes"`
	Votes                  int64 `json:"votes"`
	MistakeVotes           int64 `json:"mistakeVotes"`
	DeletedVotes           int64 `json:"deletedVotes"`
	RawRecords             int64 `json:"rawRecords"`
}

func LoadSummary(ctx context.Context, pool *pgxpool.Pool) (Summary, error) {
	var summary Summary
	err := pool.QueryRow(ctx, `
		SELECT
		  (SELECT count(*) FROM motions WHERE source_deleted = false),
		  (SELECT count(*) FROM motions WHERE source_deleted = false AND votes_synced_at IS NOT NULL),
		  (SELECT count(*) FROM motions WHERE source_deleted = false AND votes_synced_at IS NULL),
		  (SELECT count(*) FROM motions m WHERE m.source_deleted = false AND m.votes_synced_at IS NOT NULL AND NOT EXISTS (SELECT 1 FROM decisions d WHERE d.motion_key = m.motion_key AND d.source_deleted = false)),
		  (SELECT count(*) FROM decisions WHERE source_deleted = false),
		  (SELECT count(*) FROM decisions d WHERE d.source_deleted = false AND NOT EXISTS (SELECT 1 FROM votes v WHERE v.decision_key = d.decision_key AND v.source_deleted = false)),
		  (SELECT count(*) FROM votes WHERE source_deleted = false),
		  (SELECT count(*) FROM votes WHERE source_deleted = false AND mistake = true),
		  (SELECT count(*) FROM votes WHERE source_deleted = true),
		  (SELECT count(*) FROM raw_records)
	`).Scan(
		&summary.Motions,
		&summary.MotionsWithVotes,
		&summary.MotionsWithoutVotes,
		&summary.MotionsWithNoDecisions,
		&summary.Decisions,
		&summary.DecisionsWithoutVotes,
		&summary.Votes,
		&summary.MistakeVotes,
		&summary.DeletedVotes,
		&summary.RawRecords,
	)
	return summary, err
}
