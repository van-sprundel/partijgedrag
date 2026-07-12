package status

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StaleRunningRun struct {
	ID        int64
	SourceKey string
	Pipeline  string
	StartedAt time.Time
}

func LoadStaleRunningRuns(ctx context.Context, pool *pgxpool.Pool, olderThan time.Duration, limit int) ([]StaleRunningRun, error) {
	if olderThan <= 0 {
		olderThan = time.Hour
	}
	if limit <= 0 {
		limit = 50
	}

	rows, err := pool.Query(ctx, `
		SELECT id,
		       source_key,
		       pipeline,
		       started_at
		FROM ingestion_runs
		WHERE status = 'running'
		  AND started_at < $1
		ORDER BY started_at ASC
		LIMIT $2
	`, time.Now().Add(-olderThan), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []StaleRunningRun{}
	for rows.Next() {
		var run StaleRunningRun
		if err := rows.Scan(&run.ID, &run.SourceKey, &run.Pipeline, &run.StartedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func FailStaleRunningRuns(ctx context.Context, pool *pgxpool.Pool, olderThan time.Duration) ([]StaleRunningRun, error) {
	if olderThan <= 0 {
		olderThan = time.Hour
	}

	rows, err := pool.Query(ctx, `
		UPDATE ingestion_runs
		SET status = 'failed',
		    stop_reason = 'interrupted',
		    error_message = 'marked failed by maintenance after stale running state',
		    finished_at = now()
		WHERE status = 'running'
		  AND started_at < $1
		RETURNING id,
		          source_key,
		          pipeline,
		          started_at
	`, time.Now().Add(-olderThan))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []StaleRunningRun{}
	for rows.Next() {
		var run StaleRunningRun
		if err := rows.Scan(&run.ID, &run.SourceKey, &run.Pipeline, &run.StartedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}
