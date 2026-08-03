package ingest

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/source/tweedekamer"
)

const partyLogosPipeline = "party_logos.raw"

// logoEnumerationSince is the "everything" lower bound for the fractie listing.
// The listing is filtered on ApiGewijzigdOp, and the OData endpoint rejects a
// zero time, so this mirrors the default initial cursor.
var logoEnumerationSince = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

// TweedeKamerPartyLogoIngest fills in the party logos the OData API keeps as a
// separate resource per fractie. The logos let the partijgelijkenis matrix label
// its axes with an icon, since the full name does not fit.
//
// It enumerates the fracties to see which ones have a logo, then downloads only
// the ones missing locally. There is no cursor to advance. A missing logo is
// harmless: the pages fall back to a text monogram.
type TweedeKamerPartyLogoIngest struct {
	Pool        *pgxpool.Pool
	Client      *tweedekamer.Client
	BatchSize   int
	Concurrency int
	ResyncAfter time.Duration
}

func (ingest TweedeKamerPartyLogoIngest) Run(ctx context.Context) error {
	releaseLock, err := acquirePipelineLock(ctx, ingest.Pool, partyLogosPipeline)
	if err != nil {
		return err
	}
	defer releaseLock()

	runID, err := startPipelineRun(ctx, ingest.Pool, partyLogosPipeline)
	if err != nil {
		return err
	}

	fail := func(err error, seen int, changed int) error {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, partyLogosPipeline, "failed", seen, changed, false, "error", err.Error())
		return err
	}

	available, err := ingest.logoAvailability(ctx)
	if err != nil {
		return fail(err, 0, 0)
	}

	candidates, err := ingest.candidates(ctx)
	if err != nil {
		return fail(err, 0, 0)
	}

	seen, changed, failures := ingest.process(ctx, candidates, available)
	if ctx.Err() != nil {
		return fail(ctx.Err(), seen, changed)
	}

	if err := finishPipelineRun(ctx, ingest.Pool, runID, partyLogosPipeline, "succeeded", seen, changed, false, "complete", ""); err != nil {
		return err
	}

	fmt.Printf(
		"party logo backfill complete run_id=%d candidates=%d seen=%d stored=%d failed=%d with_logo_upstream=%d\n",
		runID, len(candidates), seen, changed, failures, len(available),
	)
	return nil
}

// logoAvailability asks the API which fracties expose a logo resource, so the
// backfill never spends a request on the majority that have none.
func (ingest TweedeKamerPartyLogoIngest) logoAvailability(ctx context.Context) (map[string]bool, error) {
	batchSize := ingest.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	available := map[string]bool{}
	nextURL := ""
	skip := 0
	for {
		result, err := ingest.Client.FetchChangedParties(ctx, logoEnumerationSince, batchSize, skip, nextURL)
		if err != nil {
			return nil, fmt.Errorf("enumerate parties for logos: %w", err)
		}

		for _, record := range result.Records {
			if record.HasLogo() {
				available[record.ID] = true
			}
		}

		nextURL = result.NextURL
		skip += len(result.Records)
		if nextURL == "" && len(result.Records) < batchSize {
			return available, nil
		}
	}
}

type partyLogoCandidate struct {
	PartyKey string
	SourceID string
}

func (ingest TweedeKamerPartyLogoIngest) candidates(ctx context.Context) ([]partyLogoCandidate, error) {
	rows, err := ingest.Pool.Query(ctx, `
		SELECT party_key, source_id
		FROM parties
		WHERE source_key = $1
		  AND source_deleted = false
		  AND (logo_synced_at IS NULL OR ($2::timestamptz IS NOT NULL AND logo_synced_at < $2))
		ORDER BY logo_synced_at ASC NULLS FIRST, short_name
	`, tweedeKamerSourceKey, ingest.resyncBefore())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []partyLogoCandidate
	for rows.Next() {
		var candidate partyLogoCandidate
		if err := rows.Scan(&candidate.PartyKey, &candidate.SourceID); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func (ingest TweedeKamerPartyLogoIngest) process(
	ctx context.Context,
	candidates []partyLogoCandidate,
	available map[string]bool,
) (int, int, int) {
	if len(candidates) == 0 {
		return 0, 0, 0
	}

	workerCount := ingest.Concurrency
	if workerCount <= 0 {
		workerCount = 4
	}
	if workerCount > len(candidates) {
		workerCount = len(candidates)
	}

	jobs := make(chan partyLogoCandidate)
	var mutex sync.Mutex
	seen, changed, failures := 0, 0, 0

	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for candidate := range jobs {
				outcome, stored, err := ingest.storeLogo(ctx, candidate, available[candidate.SourceID])

				mutex.Lock()
				seen++
				if stored {
					changed++
				}
				if err != nil {
					failures++
				}
				mutex.Unlock()

				if err != nil {
					fmt.Printf("party=%s outcome=error error=%v\n", candidate.SourceID, err)
					continue
				}
				fmt.Printf("party=%s outcome=%s\n", candidate.SourceID, outcome)
			}
		}()
	}

	for _, candidate := range candidates {
		select {
		case jobs <- candidate:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return seen, changed, failures
		}
	}
	close(jobs)
	wg.Wait()

	return seen, changed, failures
}

func (ingest TweedeKamerPartyLogoIngest) storeLogo(
	ctx context.Context,
	candidate partyLogoCandidate,
	hasLogoUpstream bool,
) (string, bool, error) {
	// Marking parties without a logo as synced keeps them out of the next run's
	// candidate list, so a repeated backfill costs one request per new party.
	if !hasLogoUpstream {
		return "no_logo", false, ingest.markSynced(ctx, candidate)
	}

	logo, err := ingest.Client.FetchPartyLogo(ctx, candidate.SourceID)
	if errors.Is(err, tweedekamer.ErrNoLogo) {
		return "no_logo", false, ingest.markSynced(ctx, candidate)
	}
	if err != nil {
		return "", false, err
	}

	tag, err := ingest.Pool.Exec(ctx, `
		UPDATE parties
		SET logo_data = $2,
		    logo_content_type = $3,
		    logo_synced_at = now(),
		    updated_at = now()
		WHERE party_key = $1
	`, candidate.PartyKey, logo.Data, logo.ContentType)
	if err != nil {
		return "", false, err
	}
	return "stored", tag.RowsAffected() > 0, nil
}

func (ingest TweedeKamerPartyLogoIngest) markSynced(ctx context.Context, candidate partyLogoCandidate) error {
	_, err := ingest.Pool.Exec(ctx, `
		UPDATE parties
		SET logo_synced_at = now()
		WHERE party_key = $1
	`, candidate.PartyKey)
	return err
}

func (ingest TweedeKamerPartyLogoIngest) resyncBefore() *time.Time {
	if ingest.ResyncAfter <= 0 {
		return nil
	}
	value := time.Now().Add(-ingest.ResyncAfter)
	return &value
}
