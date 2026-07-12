package status

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Summary struct {
	Parties                int64 `json:"parties"`
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

// SiteStats holds the handful of counts shown on public pages. It stays
// cheap by scanning each table at most once, unlike the full Summary.
type SiteStats struct {
	Parties          int64 `json:"parties"`
	Motions          int64 `json:"motions"`
	MotionsWithVotes int64 `json:"motionsWithVotes"`
	Votes            int64 `json:"votes"`
}

type VoteBackfill struct {
	TotalMotions       int64      `json:"totalMotions"`
	SyncedMotions      int64      `json:"syncedMotions"`
	UnsyncedMotions    int64      `json:"unsyncedMotions"`
	EligibleMotions    int64      `json:"eligibleMotions"`
	OldestVotesSynced  *time.Time `json:"oldestVotesSynced"`
	NewestVotesSynced  *time.Time `json:"newestVotesSynced"`
	ResyncBefore       *time.Time `json:"resyncBefore"`
	ResyncAfterSeconds int64      `json:"resyncAfterSeconds"`
}

type IngestionRunHealth struct {
	RunningRuns             int64 `json:"runningRuns"`
	StaleRunningRuns        int64 `json:"staleRunningRuns"`
	StaleAfterSeconds       int64 `json:"staleAfterSeconds"`
	FailedRunsLastDay       int64 `json:"failedRunsLastDay"`
	FinishedWithoutStopRuns int64 `json:"finishedWithoutStopRuns"`
}

func LoadSiteStats(ctx context.Context, pool *pgxpool.Pool) (SiteStats, error) {
	var stats SiteStats
	err := pool.QueryRow(ctx, `
		WITH motion_stats AS (
			SELECT count(*) AS motions,
			       count(*) FILTER (WHERE votes_synced_at IS NOT NULL) AS motions_with_votes
			FROM motions
			WHERE source_deleted = false
		)
		SELECT
		  (SELECT count(*) FROM parties WHERE source_deleted = false),
		  motions,
		  motions_with_votes,
		  (SELECT count(*) FROM votes WHERE source_deleted = false)
		FROM motion_stats
	`).Scan(
		&stats.Parties,
		&stats.Motions,
		&stats.MotionsWithVotes,
		&stats.Votes,
	)
	return stats, err
}

func LoadSummary(ctx context.Context, pool *pgxpool.Pool) (Summary, error) {
	var summary Summary
	err := pool.QueryRow(ctx, `
		WITH motion_stats AS (
			SELECT count(*) FILTER (WHERE source_deleted = false) AS motions,
			       count(*) FILTER (WHERE source_deleted = false AND votes_synced_at IS NOT NULL) AS motions_with_votes,
			       count(*) FILTER (WHERE source_deleted = false AND votes_synced_at IS NULL) AS motions_without_votes
			FROM motions
		),
		vote_stats AS (
			SELECT count(*) FILTER (WHERE source_deleted = false) AS votes,
			       count(*) FILTER (WHERE source_deleted = false AND mistake = true) AS mistake_votes,
			       count(*) FILTER (WHERE source_deleted = true) AS deleted_votes
			FROM votes
		)
		SELECT
		  (SELECT count(*) FROM parties WHERE source_deleted = false),
		  motion_stats.motions,
		  motion_stats.motions_with_votes,
		  motion_stats.motions_without_votes,
		  (SELECT count(*) FROM motions m WHERE m.source_deleted = false AND m.votes_synced_at IS NOT NULL AND NOT EXISTS (SELECT 1 FROM decisions d WHERE d.motion_key = m.motion_key AND d.source_deleted = false)),
		  (SELECT count(*) FROM decisions WHERE source_deleted = false),
		  (SELECT count(*) FROM decisions d WHERE d.source_deleted = false AND NOT EXISTS (SELECT 1 FROM votes v WHERE v.decision_key = d.decision_key AND v.source_deleted = false)),
		  vote_stats.votes,
		  vote_stats.mistake_votes,
		  vote_stats.deleted_votes,
		  (SELECT count(*) FROM raw_records)
		FROM motion_stats, vote_stats
	`).Scan(
		&summary.Parties,
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

func LoadVoteBackfill(ctx context.Context, pool *pgxpool.Pool, resyncAfter time.Duration) (VoteBackfill, error) {
	var result VoteBackfill
	var resyncBefore *time.Time
	if resyncAfter > 0 {
		value := time.Now().Add(-resyncAfter)
		resyncBefore = &value
		result.ResyncBefore = resyncBefore
		result.ResyncAfterSeconds = int64(resyncAfter.Seconds())
	}

	err := pool.QueryRow(ctx, `
		SELECT
		  count(*)::bigint AS total_motions,
		  count(*) FILTER (WHERE votes_synced_at IS NOT NULL)::bigint AS synced_motions,
		  count(*) FILTER (WHERE votes_synced_at IS NULL)::bigint AS unsynced_motions,
		  count(*) FILTER (WHERE votes_synced_at IS NULL OR ($1::timestamptz IS NOT NULL AND votes_synced_at < $1))::bigint AS eligible_motions,
		  min(votes_synced_at) FILTER (WHERE votes_synced_at IS NOT NULL) AS oldest_votes_synced,
		  max(votes_synced_at) FILTER (WHERE votes_synced_at IS NOT NULL) AS newest_votes_synced
		FROM motions
		WHERE source_key = 'tweedekamer-odata-v2'
		  AND source_deleted = false
	`, resyncBefore).Scan(
		&result.TotalMotions,
		&result.SyncedMotions,
		&result.UnsyncedMotions,
		&result.EligibleMotions,
		&result.OldestVotesSynced,
		&result.NewestVotesSynced,
	)
	return result, err
}

func LoadIngestionRunHealth(ctx context.Context, pool *pgxpool.Pool, staleAfter time.Duration) (IngestionRunHealth, error) {
	if staleAfter <= 0 {
		staleAfter = time.Hour
	}

	staleBefore := time.Now().Add(-staleAfter)
	var health IngestionRunHealth
	health.StaleAfterSeconds = int64(staleAfter.Seconds())

	err := pool.QueryRow(ctx, `
		SELECT
		  count(*) FILTER (WHERE status = 'running')::bigint AS running_runs,
		  count(*) FILTER (WHERE status = 'running' AND started_at < $1)::bigint AS stale_running_runs,
		  count(*) FILTER (WHERE status = 'failed' AND started_at >= now() - interval '24 hours')::bigint AS failed_runs_last_day,
		  count(*) FILTER (WHERE status <> 'running' AND COALESCE(stop_reason, '') = '')::bigint AS finished_without_stop_runs
		FROM ingestion_runs
	`, staleBefore).Scan(
		&health.RunningRuns,
		&health.StaleRunningRuns,
		&health.FailedRunsLastDay,
		&health.FinishedWithoutStopRuns,
	)
	return health, err
}
