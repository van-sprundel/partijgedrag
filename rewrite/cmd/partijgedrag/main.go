package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"partijgedrag/rewrite/internal/config"
	"partijgedrag/rewrite/internal/db"
	"partijgedrag/rewrite/internal/httpapi"
	"partijgedrag/rewrite/internal/ingest"
	"partijgedrag/rewrite/internal/inspect"
	"partijgedrag/rewrite/internal/migrate"
	"partijgedrag/rewrite/internal/source/tweedekamer"
	"partijgedrag/rewrite/internal/status"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer database.Close()

	args := os.Args[1:]
	if len(args) == 0 {
		return usage()
	}

	switch args[0] {
	case "migrate":
		return migrate.Run(ctx, database.Pool)
	case "ingest":
		return runIngest(ctx, cfg, database, args[1:])
	case "sync":
		return runSync(ctx, cfg, database, args[1:])
	case "status":
		return runStatus(ctx, database, args[1:])
	case "inspect":
		return runInspect(ctx, database, args[1:])
	case "serve":
		address := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
		fmt.Printf("partijgedrag listening on http://%s\n", address)
		server := httpapi.Server{Pool: database.Pool}
		err := httpapi.ListenAndServe(ctx, address, server.Handler())
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	default:
		return usage()
	}
}

func runInspect(ctx context.Context, database *db.DB, args []string) error {
	if len(args) != 2 || args[0] != "motion" {
		return usage()
	}
	return inspect.PrintMotion(ctx, database.Pool, os.Stdout, args[1])
}

func runIngest(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	if len(args) < 2 || args[0] != "tweedekamer" {
		return usage()
	}

	switch args[1] {
	case "parties":
		return runIngestParties(ctx, cfg, database, args[2:])
	case "motions":
		return runIngestMotions(ctx, cfg, database, args[2:])
	case "motion-votes":
		return runIngestMotionVotes(ctx, cfg, database, args[2:])
	default:
		return usage()
	}
}

func runIngestParties(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("ingest tweedekamer parties", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	maxPages := flags.Int("max-pages", cfg.TweedeKamerMaxPages, "maximum OData pages to process, 0 means all")
	batchSize := flags.Int("batch-size", cfg.TweedeKamerBatchSize, "records per OData page")
	resetCursor := flags.Bool("reset-cursor", false, "delete the stored cursor before ingesting")
	sinceValue := flags.String("since", "", "override cursor with an RFC3339 ApiGewijzigdOp timestamp")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *batchSize <= 0 {
		return fmt.Errorf("--batch-size must be greater than 0")
	}
	if *maxPages < 0 {
		return fmt.Errorf("--max-pages must be 0 or greater")
	}

	var sinceOverride *time.Time
	if *sinceValue != "" {
		parsed, err := time.Parse(time.RFC3339, *sinceValue)
		if err != nil {
			return fmt.Errorf("parse --since: %w", err)
		}
		sinceOverride = &parsed
	}

	job := ingest.TweedeKamerPartyIngest{
		Pool:          database.Pool,
		Client:        tweedekamer.NewClient(cfg.TweedeKamerODataBaseURL),
		BatchSize:     *batchSize,
		MaxPages:      *maxPages,
		InitialSince:  cfg.TweedeKamerInitialSince,
		CursorOverlap: cfg.CursorOverlap,
		SinceOverride: sinceOverride,
		ResetCursor:   *resetCursor,
	}
	return job.Run(ctx)
}

func runIngestMotions(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("ingest tweedekamer motions", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	maxPages := flags.Int("max-pages", cfg.TweedeKamerMaxPages, "maximum OData pages to process, 0 means all")
	batchSize := flags.Int("batch-size", cfg.TweedeKamerBatchSize, "records per OData page")
	resetCursor := flags.Bool("reset-cursor", false, "delete the stored cursor before ingesting")
	sinceValue := flags.String("since", "", "override cursor with an RFC3339 ApiGewijzigdOp timestamp")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *batchSize <= 0 {
		return fmt.Errorf("--batch-size must be greater than 0")
	}
	if *maxPages < 0 {
		return fmt.Errorf("--max-pages must be 0 or greater")
	}

	var sinceOverride *time.Time
	if *sinceValue != "" {
		parsed, err := time.Parse(time.RFC3339, *sinceValue)
		if err != nil {
			return fmt.Errorf("parse --since: %w", err)
		}
		sinceOverride = &parsed
	}

	job := ingest.TweedeKamerMotionIngest{
		Pool:          database.Pool,
		Client:        tweedekamer.NewClient(cfg.TweedeKamerODataBaseURL),
		BatchSize:     *batchSize,
		MaxPages:      *maxPages,
		InitialSince:  cfg.TweedeKamerInitialSince,
		CursorOverlap: cfg.CursorOverlap,
		SinceOverride: sinceOverride,
		ResetCursor:   *resetCursor,
	}
	return job.Run(ctx)
}

func runIngestMotionVotes(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("ingest tweedekamer motion-votes", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	limit := flags.Int("limit", 25, "number of motions to sync votes for")
	resyncAfter := flags.Duration("resync-after", 0, "also resync motions whose votes were synced before this duration, e.g. 168h; 0 means only unsynced")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *limit <= 0 {
		return fmt.Errorf("--limit must be greater than 0")
	}
	if *resyncAfter < 0 {
		return fmt.Errorf("--resync-after must be 0 or greater")
	}

	job := ingest.TweedeKamerMotionVotesIngest{
		Pool:        database.Pool,
		Client:      tweedekamer.NewClient(cfg.TweedeKamerODataBaseURL),
		Limit:       *limit,
		ResyncAfter: *resyncAfter,
	}
	return job.Run(ctx)
}

func runSync(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	if len(args) == 0 || args[0] != "tweedekamer" {
		return usage()
	}

	flags := flag.NewFlagSet("sync tweedekamer", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	motionMaxPages := flags.Int("motion-max-pages", cfg.TweedeKamerMaxPages, "maximum motion OData pages to process, 0 means all")
	motionBatchSize := flags.Int("motion-batch-size", cfg.TweedeKamerBatchSize, "motion records per OData page")
	partyMaxPages := flags.Int("party-max-pages", cfg.TweedeKamerMaxPages, "maximum party OData pages to process, 0 means all")
	partyBatchSize := flags.Int("party-batch-size", cfg.TweedeKamerBatchSize, "party records per OData page")
	motionVoteLimit := flags.Int("motion-vote-limit", 100, "number of known motions to sync votes for")
	motionVoteResyncAfter := flags.Duration("motion-vote-resync-after", 0, "also resync motions whose votes were synced before this duration, e.g. 168h; 0 means only unsynced")
	skipParties := flags.Bool("skip-parties", false, "skip party ingestion")
	skipMotions := flags.Bool("skip-motions", false, "skip motion ingestion")
	skipMotionVotes := flags.Bool("skip-motion-votes", false, "skip motion vote ingestion")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return usage()
	}
	if *motionMaxPages < 0 {
		return fmt.Errorf("--motion-max-pages must be 0 or greater")
	}
	if *motionBatchSize <= 0 {
		return fmt.Errorf("--motion-batch-size must be greater than 0")
	}
	if *partyMaxPages < 0 {
		return fmt.Errorf("--party-max-pages must be 0 or greater")
	}
	if *partyBatchSize <= 0 {
		return fmt.Errorf("--party-batch-size must be greater than 0")
	}
	if *motionVoteLimit <= 0 {
		return fmt.Errorf("--motion-vote-limit must be greater than 0")
	}
	if *motionVoteResyncAfter < 0 {
		return fmt.Errorf("--motion-vote-resync-after must be 0 or greater")
	}
	if *skipParties && *skipMotions && *skipMotionVotes {
		return fmt.Errorf("sync has nothing to do when --skip-parties, --skip-motions, and --skip-motion-votes are set")
	}

	client := tweedekamer.NewClient(cfg.TweedeKamerODataBaseURL)
	if !*skipParties {
		fmt.Println("sync step=parties")
		job := ingest.TweedeKamerPartyIngest{
			Pool:          database.Pool,
			Client:        client,
			BatchSize:     *partyBatchSize,
			MaxPages:      *partyMaxPages,
			InitialSince:  cfg.TweedeKamerInitialSince,
			CursorOverlap: cfg.CursorOverlap,
		}
		if err := job.Run(ctx); err != nil {
			return err
		}
	}

	if !*skipMotions {
		fmt.Println("sync step=motions")
		job := ingest.TweedeKamerMotionIngest{
			Pool:          database.Pool,
			Client:        client,
			BatchSize:     *motionBatchSize,
			MaxPages:      *motionMaxPages,
			InitialSince:  cfg.TweedeKamerInitialSince,
			CursorOverlap: cfg.CursorOverlap,
		}
		if err := job.Run(ctx); err != nil {
			return err
		}
	}

	if !*skipMotionVotes {
		fmt.Println("sync step=motion-votes")
		job := ingest.TweedeKamerMotionVotesIngest{
			Pool:        database.Pool,
			Client:      client,
			Limit:       *motionVoteLimit,
			ResyncAfter: *motionVoteResyncAfter,
		}
		if err := job.Run(ctx); err != nil {
			return err
		}
	}

	fmt.Println("sync complete source=tweedekamer")
	return nil
}

func runStatus(ctx context.Context, database *db.DB, args []string) error {
	if len(args) == 0 {
		return usage()
	}

	switch args[0] {
	case "ingestion-runs":
		return runStatusIngestionRuns(ctx, database, args[1:])
	case "summary":
		return runStatusSummary(ctx, database, args[1:])
	case "vote-backfill":
		return runStatusVoteBackfill(ctx, database, args[1:])
	default:
		return usage()
	}
}

func runStatusIngestionRuns(ctx context.Context, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("status ingestion-runs", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	limit := flags.Int("limit", 10, "number of runs to show")
	pipeline := flags.String("pipeline", "", "filter by pipeline")
	failedOnly := flags.Bool("failed", false, "show failed runs only")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *limit <= 0 {
		return fmt.Errorf("--limit must be greater than 0")
	}

	rows, err := database.Pool.Query(ctx, `
		SELECT id,
		       source_key,
		       pipeline,
		       status,
		       cursor_saved,
		       stop_reason,
		       records_seen,
		       records_changed,
		       error_message,
		       started_at,
		       finished_at
		FROM ingestion_runs
		WHERE ($1::text = '' OR pipeline = $1)
		  AND ($2::boolean = false OR status = 'failed')
		ORDER BY started_at DESC
		LIMIT $3
	`, *pipeline, *failedOnly, *limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var sourceKey, pipeline, status string
		var cursorSaved bool
		var stopReason *string
		var recordsSeen, recordsChanged int
		var errorMessage *string
		var startedAt time.Time
		var finishedAt *time.Time
		if err := rows.Scan(&id, &sourceKey, &pipeline, &status, &cursorSaved, &stopReason, &recordsSeen, &recordsChanged, &errorMessage, &startedAt, &finishedAt); err != nil {
			return err
		}

		finished := "running"
		if finishedAt != nil {
			finished = finishedAt.Format(time.RFC3339)
		}
		errorText := ""
		if errorMessage != nil {
			errorText = " error=" + *errorMessage
		}
		stopText := ""
		if stopReason != nil && *stopReason != "" {
			stopText = " stop=" + *stopReason
		}
		fmt.Printf("#%d %s/%s %s seen=%d changed=%d cursor_saved=%t%s started=%s finished=%s%s\n",
			id,
			sourceKey,
			pipeline,
			status,
			recordsSeen,
			recordsChanged,
			cursorSaved,
			stopText,
			startedAt.Format(time.RFC3339),
			finished,
			errorText,
		)
	}
	return rows.Err()
}

func runStatusSummary(ctx context.Context, database *db.DB, args []string) error {
	if len(args) != 0 {
		return usage()
	}

	summary, err := status.LoadSummary(ctx, database.Pool)
	if err != nil {
		return err
	}

	fmt.Printf("parties=%d motions=%d motions_with_votes=%d motions_without_votes=%d motions_with_no_decisions=%d decisions=%d decisions_without_votes=%d votes=%d mistake_votes=%d deleted_votes=%d raw_records=%d\n",
		summary.Parties,
		summary.Motions,
		summary.MotionsWithVotes,
		summary.MotionsWithoutVotes,
		summary.MotionsWithNoDecisions,
		summary.Decisions,
		summary.DecisionsWithoutVotes,
		summary.Votes,
		summary.MistakeVotes,
		summary.DeletedVotes,
		summary.RawRecords,
	)
	return nil
}

func runStatusVoteBackfill(ctx context.Context, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("status vote-backfill", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	resyncAfter := flags.Duration("resync-after", 0, "include motions whose votes were synced before this duration, e.g. 168h")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return usage()
	}
	if *resyncAfter < 0 {
		return fmt.Errorf("--resync-after must be 0 or greater")
	}

	backfill, err := status.LoadVoteBackfill(ctx, database.Pool, *resyncAfter)
	if err != nil {
		return err
	}

	fmt.Printf("total=%d synced=%d unsynced=%d eligible=%d oldest_synced=%s newest_synced=%s resync_before=%s\n",
		backfill.TotalMotions,
		backfill.SyncedMotions,
		backfill.UnsyncedMotions,
		backfill.EligibleMotions,
		formatOptionalTime(backfill.OldestVotesSynced),
		formatOptionalTime(backfill.NewestVotesSynced),
		formatOptionalTime(backfill.ResyncBefore),
	)
	return nil
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func usage() error {
	return fmt.Errorf(`usage:
  partijgedrag migrate
  partijgedrag ingest tweedekamer parties [--max-pages=N] [--batch-size=N] [--since=RFC3339] [--reset-cursor]
  partijgedrag ingest tweedekamer motions [--max-pages=N] [--batch-size=N] [--since=RFC3339] [--reset-cursor]
  partijgedrag ingest tweedekamer motion-votes [--limit=N] [--resync-after=168h]
  partijgedrag sync tweedekamer [--party-max-pages=N] [--party-batch-size=N] [--motion-max-pages=N] [--motion-batch-size=N] [--motion-vote-limit=N] [--motion-vote-resync-after=168h] [--skip-parties] [--skip-motions] [--skip-motion-votes]
  partijgedrag status ingestion-runs [--limit=N] [--pipeline=NAME] [--failed]
  partijgedrag status summary
  partijgedrag status vote-backfill [--resync-after=168h]
  partijgedrag inspect motion MOTION_KEY
  partijgedrag serve`)
}
