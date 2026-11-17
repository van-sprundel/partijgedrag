package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"partijgedrag/internal/config"
	"partijgedrag/internal/db"
	"partijgedrag/internal/ingest"
	"partijgedrag/internal/source/openbesluitvorming"
	"partijgedrag/internal/status"
)

func runIngestOpenBesluitvorming(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	switch args[0] {
	case "motions":
		flags := flag.NewFlagSet("ingest openbesluitvorming motions", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		source := flags.String("source", "", "source key, e.g. utrecht")
		maxPages := flags.Int("max-pages", 0, "max pages to process (0 for all)")
		pageSize := flags.Int("page-size", 1000, "page size")
		resetCursor := flags.Bool("reset-cursor", false, "reset cursor")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if *source == "" || flags.NArg() != 0 {
			return usage()
		}
		return ingest.OpenBesluitvormingMotionIngest{
			Pool:        database.Pool,
			Client:      openbesluitvorming.NewClient(cfg.OpenBesluitvormingBaseURL),
			Source:      *source,
			PageSize:    *pageSize,
			MaxPages:    *maxPages,
			ResetCursor: *resetCursor,
		}.Run(ctx)

	case "motion-documents":
		flags := flag.NewFlagSet("ingest openbesluitvorming motion-documents", flag.ContinueOnError)
		flags.SetOutput(os.Stderr)
		source := flags.String("source", "", "source key, e.g. utrecht")
		limit := flags.Int("limit", 500, "maximum motion texts to fetch")
		reparse := flags.Bool("reparse", false, "reparse stored texts without fetching")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if *source == "" || flags.NArg() != 0 {
			return usage()
		}
		return ingest.OpenBesluitvormingMotionDocumentsIngest{
			Pool:    database.Pool,
			Client:  openbesluitvorming.NewClient(cfg.OpenBesluitvormingBaseURL),
			Source:  *source,
			Limit:   *limit,
			Reparse: *reparse,
		}.Run(ctx)
	}
	return usage()
}

func runSyncOpenBesluitvorming(ctx context.Context, cfg config.Config, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("sync openbesluitvorming", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	sources := flags.String("source", "", "source keys, comma-separated")
	documentLimit := flags.Int("document-limit", 500, "maximum motion texts to fetch per source (0 to skip)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *sources == "" || flags.NArg() != 0 {
		return usage()
	}

	client := openbesluitvorming.NewClient(cfg.OpenBesluitvormingBaseURL)
	var failures []error
	for _, source := range strings.Split(*sources, ",") {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		if err := (ingest.OpenBesluitvormingMotionIngest{
			Pool: database.Pool, Client: client, Source: source, PageSize: 1000,
		}).Run(ctx); err != nil {
			failures = append(failures, fmt.Errorf("%s motions: %w", source, err))
			continue
		}
		if *documentLimit <= 0 {
			continue
		}
		if err := (ingest.OpenBesluitvormingMotionDocumentsIngest{
			Pool: database.Pool, Client: client, Source: source, Limit: *documentLimit,
		}).Run(ctx); err != nil {
			failures = append(failures, fmt.Errorf("%s motion documents: %w", source, err))
		}
	}
	return errors.Join(failures...)
}

func runStatusOpenBesluitvorming(ctx context.Context, database *db.DB, args []string) error {
	flags := flag.NewFlagSet("status openbesluitvorming", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	source := flags.String("source", "", "source key, e.g. utrecht")
	spellings := flags.Bool("spellings", false, "show raw group names")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *source == "" || flags.NArg() != 0 {
		return usage()
	}

	report, err := status.LoadOBVReport(ctx, database.Pool, ingest.OpenBesluitvormingSourceKey(*source))
	if err != nil {
		return fmt.Errorf("load report for %s: %w", *source, err)
	}

	pct := func(part, whole int64) string {
		if whole == 0 {
			return "-"
		}
		return fmt.Sprintf("%.0f%%", 100*float64(part)/float64(whole))
	}

	fmt.Printf("%s (%s, %s)\n\n", report.Label, report.SourceKey, report.JurisdictionKey)
	fmt.Printf("motions             %6d  (moties %d, amendementen %d)\n", report.Motions, report.Moties, report.Amendments)
	fmt.Printf("with votes          %6d  (moties %d)\n", report.MotionsWithVotes, report.MotiesWithVotes)
	fmt.Printf("voted, from-to      %s .. %s\n", formatOptionalDate(report.FirstVote), formatOptionalDate(report.LastVote))
	fmt.Printf("voted moties w/ doc %6d  %s\n", report.MotiesWithDocument, pct(report.MotiesWithDocument, report.MotiesWithVotes))
	fmt.Printf("texts fetched       %6d\n", report.TextsFetched)
	fmt.Printf("moties w/ bullets   %6d  %s\n", report.MotiesWithBullets, pct(report.MotiesWithBullets, report.TextsFetched))
	fmt.Printf("  with ask          %6d  %s\n", report.MotiesWithAsk, pct(report.MotiesWithAsk, report.TextsFetched))
	fmt.Printf("  with verzoekt     %6d  %s\n", report.MotiesWithVerzoekt, pct(report.MotiesWithVerzoekt, report.TextsFetched))
	fmt.Printf("  statement only    %6d  %s\n", report.MotiesStatementOnly, pct(report.MotiesStatementOnly, report.TextsFetched))
	fmt.Printf("votes               %6d  by %d members\n", report.Votes, report.Members)

	var unattributed int64
	for _, item := range report.UnattributedBy {
		unattributed += item.Count
	}
	fmt.Printf("votes w/o fractie   %6d  %s\n", unattributed, pct(unattributed, report.Votes))
	for _, item := range report.UnattributedBy {
		fmt.Printf("    %-40q %d\n", item.Name, item.Count)
	}

	fmt.Printf("\noutcome mismatches:\n")
	if len(report.OutcomeMismatch) == 0 {
		fmt.Println("    none")
	}
	for _, item := range report.OutcomeMismatch {
		fmt.Printf("    %-11s vs %-11s status=%-40q %d\n", item.Derived, item.Upstream, item.Status, item.Count)
	}

	fmt.Printf("\nparties:\n")
	fmt.Printf("    %-28s %-30s %5s %8s  %-10s %-10s %s\n", "source_id", "name", "seats", "votes", "from", "to", "spellings")
	for _, party := range report.Parties {
		to := formatOptionalDate(party.ActiveTo)
		if to == "" {
			to = "active"
		}
		fmt.Printf("    %-28s %-30s %5d %8d  %-10s %-10s %d\n",
			party.SourceID, truncate(party.Name, 30), party.Seats, party.Votes, formatOptionalDate(party.ActiveFrom), to, len(party.Spellings))
		if *spellings || len(party.Spellings) > 1 {
			for _, spelling := range party.Spellings {
				fmt.Printf("        %-40q %d\n", spelling.Name, spelling.Count)
			}
		}
	}
	return nil
}

func formatOptionalDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-1]) + "…"
}
