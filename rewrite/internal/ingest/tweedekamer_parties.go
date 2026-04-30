package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/rewrite/internal/source/tweedekamer"
)

const (
	partiesPipeline   = "parties.raw"
	fractieCollection = "Fractie"
)

type TweedeKamerPartyIngest struct {
	Pool          *pgxpool.Pool
	Client        *tweedekamer.Client
	BatchSize     int
	MaxPages      int
	InitialSince  time.Time
	CursorOverlap time.Duration
	SinceOverride *time.Time
	ResetCursor   bool
}

func (ingest TweedeKamerPartyIngest) Run(ctx context.Context) error {
	releaseLock, err := acquirePipelineLock(ctx, ingest.Pool, partiesPipeline)
	if err != nil {
		return err
	}
	defer releaseLock()

	source, err := ingest.getSource(ctx)
	if err != nil {
		return err
	}

	if ingest.ResetCursor {
		if err := ingest.resetCursor(ctx); err != nil {
			return err
		}
	}

	cursorBefore, err := ingest.getCursor(ctx)
	if err != nil {
		return err
	}
	if ingest.SinceOverride != nil {
		cursorBefore = Cursor{ApiUpdatedAt: ingest.SinceOverride}
	}

	since := ingest.cursorSince(cursorBefore)
	if ingest.SinceOverride != nil {
		since = *ingest.SinceOverride
	}

	runID, err := startPipelineRunWithCursor(ctx, ingest.Pool, partiesPipeline, cursorBefore)
	if err != nil {
		return err
	}

	recordsSeen := 0
	recordsChanged := 0
	maxUpdatedAt := cursorBefore.ApiUpdatedAt
	nextURL := ""
	stopReason := ""
	skip := 0
	pagesProcessed := 0

	for page := 1; ; page++ {
		result, err := ingest.Client.FetchChangedParties(ctx, since, ingest.BatchSize, skip, nextURL)
		if err != nil {
			_ = finishPipelineRunWithCursor(ctx, ingest.Pool, runID, partiesPipeline, "failed", cursorBefore, recordsSeen, recordsChanged, false, "error", err.Error())
			return err
		}

		nextURL = result.NextURL
		recordsSeen += len(result.Records)
		skip += len(result.Records)
		pagesProcessed = page

		for _, record := range result.Records {
			changed, err := ingest.storePartyRecord(ctx, source.JurisdictionKey, record)
			if err != nil {
				_ = finishPipelineRunWithCursor(ctx, ingest.Pool, runID, partiesPipeline, "failed", cursorBefore, recordsSeen, recordsChanged, false, "error", err.Error())
				return err
			}
			if changed {
				recordsChanged++
			}

			apiUpdatedAt := timePtr(record.ApiGewijzigdOp)
			if apiUpdatedAt != nil && (maxUpdatedAt == nil || apiUpdatedAt.After(*maxUpdatedAt)) {
				value := *apiUpdatedAt
				maxUpdatedAt = &value
			}
		}

		hasMore := nextURL != "" || len(result.Records) == ingest.BatchSize
		fmt.Printf("parties page=%d seen=%d changed=%d next=%t\n", page, recordsSeen, recordsChanged, hasMore)

		if !hasMore {
			stopReason = "complete"
			break
		}
		if ingest.MaxPages > 0 && page >= ingest.MaxPages {
			stopReason = "max_pages"
			break
		}
	}

	cursorAfter := Cursor{ApiUpdatedAt: maxUpdatedAt}
	cursorSaved := stopReason == "complete"
	if cursorSaved {
		if err := ingest.saveCursor(ctx, cursorAfter); err != nil {
			_ = finishPipelineRunWithCursor(ctx, ingest.Pool, runID, partiesPipeline, "failed", cursorAfter, recordsSeen, recordsChanged, false, stopReason, err.Error())
			return err
		}
	} else {
		cursorAfter = cursorBefore
	}

	if err := finishPipelineRunWithCursor(ctx, ingest.Pool, runID, partiesPipeline, "succeeded", cursorAfter, recordsSeen, recordsChanged, cursorSaved, stopReason, ""); err != nil {
		return err
	}

	fmt.Printf(
		"party ingestion complete run_id=%d pages=%d seen=%d changed=%d cursor_before=%s cursor_after=%s cursor_saved=%t stop_reason=%s\n",
		runID,
		pagesProcessed,
		recordsSeen,
		recordsChanged,
		formatCursor(cursorBefore),
		formatCursor(cursorAfter),
		cursorSaved,
		stopReason,
	)
	return nil
}

func (ingest TweedeKamerPartyIngest) getSource(ctx context.Context) (source, error) {
	var result source
	err := ingest.Pool.QueryRow(ctx, `
		SELECT source_key, jurisdiction_key, base_url
		FROM data_sources
		WHERE source_key = $1 AND enabled = true
	`, tweedeKamerSourceKey).Scan(&result.SourceKey, &result.JurisdictionKey, &result.BaseURL)
	if err != nil {
		return source{}, fmt.Errorf("get data source %s: %w", tweedeKamerSourceKey, err)
	}
	return result, nil
}

func (ingest TweedeKamerPartyIngest) getCursor(ctx context.Context) (Cursor, error) {
	var raw []byte
	err := ingest.Pool.QueryRow(ctx, `
		SELECT cursor
		FROM source_cursors
		WHERE source_key = $1 AND pipeline = $2
	`, tweedeKamerSourceKey, partiesPipeline).Scan(&raw)
	if err == pgx.ErrNoRows {
		initial := ingest.InitialSince
		return Cursor{ApiUpdatedAt: &initial}, nil
	}
	if err != nil {
		return Cursor{}, err
	}

	var cursor Cursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return Cursor{}, err
	}
	return cursor, nil
}

func (ingest TweedeKamerPartyIngest) resetCursor(ctx context.Context) error {
	_, err := ingest.Pool.Exec(ctx, `
		DELETE FROM source_cursors
		WHERE source_key = $1 AND pipeline = $2
	`, tweedeKamerSourceKey, partiesPipeline)
	return err
}

func (ingest TweedeKamerPartyIngest) cursorSince(cursor Cursor) time.Time {
	if cursor.ApiUpdatedAt == nil {
		return ingest.InitialSince
	}
	if cursor.ApiUpdatedAt.Equal(ingest.InitialSince) {
		return ingest.InitialSince
	}
	return cursor.ApiUpdatedAt.Add(-ingest.CursorOverlap)
}

func (ingest TweedeKamerPartyIngest) saveCursor(ctx context.Context, cursorAfter Cursor) error {
	raw, err := json.Marshal(cursorAfter)
	if err != nil {
		return err
	}

	_, err = ingest.Pool.Exec(ctx, `
		INSERT INTO source_cursors (source_key, pipeline, cursor, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (source_key, pipeline)
		DO UPDATE SET cursor = EXCLUDED.cursor,
		              updated_at = now()
	`, tweedeKamerSourceKey, partiesPipeline, string(raw))
	return err
}

func (ingest TweedeKamerPartyIngest) storePartyRecord(
	ctx context.Context,
	jurisdictionKey string,
	record tweedekamer.PartyRecord,
) (bool, error) {
	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	raw := projectPartyRaw(record)
	party := projectParty(jurisdictionKey, record)

	rawChanged, err := storeRawRecord(ctx, tx, raw)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO parties (
			party_key,
			source_key,
			jurisdiction_key,
			source_id,
			number,
			short_name,
			name,
			name_en,
			seats,
			electoral_votes,
			active_from,
			active_to,
			source_updated_at,
			source_deleted,
			raw_collection,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::timestamptz::date, $12::timestamptz::date, $13, $14, $15, now())
		ON CONFLICT (source_key, source_id)
		DO UPDATE SET number = EXCLUDED.number,
		              short_name = EXCLUDED.short_name,
		              name = EXCLUDED.name,
		              name_en = EXCLUDED.name_en,
		              seats = EXCLUDED.seats,
		              electoral_votes = EXCLUDED.electoral_votes,
		              active_from = EXCLUDED.active_from,
		              active_to = EXCLUDED.active_to,
		              source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              updated_at = now()
	`, party.PartyKey, party.SourceKey, party.JurisdictionKey, party.SourceID, party.Number, party.ShortName, party.Name, party.NameEN, party.Seats, party.ElectoralVotes, party.ActiveFrom, party.ActiveTo, party.SourceUpdatedAt, party.SourceDeleted, party.RawCollection)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return rawChanged, nil
}
