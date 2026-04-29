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
	"partijgedrag/rewrite/internal/migrate"
	"partijgedrag/rewrite/internal/source/tweedekamer"
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
	case "status":
		return runStatus(ctx, database, args[1:])
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

func runIngest(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	if len(args) < 2 || args[0] != "tweedekamer" || args[1] != "motions" {
		return usage()
	}

	flags := flag.NewFlagSet("ingest tweedekamer motions", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	maxPages := flags.Int("max-pages", cfg.TweedeKamerMaxPages, "maximum OData pages to process, 0 means all")
	batchSize := flags.Int("batch-size", cfg.TweedeKamerBatchSize, "records per OData page")
	resetCursor := flags.Bool("reset-cursor", false, "delete the stored cursor before ingesting")
	sinceValue := flags.String("since", "", "override cursor with an RFC3339 ApiGewijzigdOp timestamp")
	if err := flags.Parse(args[2:]); err != nil {
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

func runStatus(ctx context.Context, database *db.DB, args []string) error {
	if len(args) != 1 || args[0] != "ingestion-runs" {
		return usage()
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
		ORDER BY started_at DESC
		LIMIT 10
	`)
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

func usage() error {
	return fmt.Errorf(`usage:
  partijgedrag migrate
  partijgedrag ingest tweedekamer motions [--max-pages=N] [--batch-size=N] [--since=RFC3339] [--reset-cursor]
  partijgedrag status ingestion-runs
  partijgedrag serve`)
}
