package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/source/officielebekendmakingen"
	"partijgedrag/internal/source/tweedekamer"
)

const motionDocumentsPipeline = "motion_documents.raw"

// TweedeKamerMotionDocumentsIngest enriches motions with the bullet points of
// their published motion text. It resolves the kamerstuk publication via the
// OData Zaak -> Kamerstukdossier/Document navigation, fetches the XML from
// zoek.officielebekendmakingen.nl, and stores the constaterende/overwegende/
// verzoekt paragraphs on the motion row.
type TweedeKamerMotionDocumentsIngest struct {
	Pool        *pgxpool.Pool
	Client      *tweedekamer.Client
	Documents   *officielebekendmakingen.Client
	Limit       int
	Concurrency int
	ResyncAfter time.Duration
	// ResyncGrace retries motions that still have no bullet points, until this
	// long after they were proposed. A published document never changes, so a
	// successful extraction is never revisited. This covers motions ingested
	// before their text appeared on officielebekendmakingen. Zero disables.
	ResyncGrace time.Duration
}

func (ingest TweedeKamerMotionDocumentsIngest) Run(ctx context.Context) error {
	releaseLock, err := acquirePipelineLock(ctx, ingest.Pool, motionDocumentsPipeline)
	if err != nil {
		return err
	}
	defer releaseLock()

	runID, err := startPipelineRun(ctx, ingest.Pool, motionDocumentsPipeline)
	if err != nil {
		return err
	}

	resyncBefore := ingest.resyncBefore()
	proposedAfter := ingest.proposedAfter()
	pendingBefore, err := ingest.pendingCount(ctx, resyncBefore, proposedAfter)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionDocumentsPipeline, "failed", 0, 0, false, "error", err.Error())
		return err
	}

	motions, err := ingest.motionCandidates(ctx, resyncBefore, proposedAfter)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionDocumentsPipeline, "failed", 0, 0, false, "error", err.Error())
		return err
	}

	recordsSeen, recordsChanged, err := ingest.processMotionCandidates(ctx, motions)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionDocumentsPipeline, "failed", recordsSeen, recordsChanged, false, "error", err.Error())
		return err
	}

	pendingAfter, err := ingest.pendingCount(ctx, resyncBefore, proposedAfter)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionDocumentsPipeline, "failed", recordsSeen, recordsChanged, false, "error", err.Error())
		return err
	}

	// Blank retries keep motions eligible across runs, so pendingAfter no
	// longer drains to zero. Only a filled batch means we were cut short.
	stopReason := "complete"
	if ingest.Limit > 0 && len(motions) >= ingest.Limit {
		stopReason = "batch_limit"
	}

	if err := finishPipelineRun(ctx, ingest.Pool, runID, motionDocumentsPipeline, "succeeded", recordsSeen, recordsChanged, false, stopReason, ""); err != nil {
		return err
	}

	fmt.Printf("motion document batch complete run_id=%d motions=%d seen=%d changed=%d pending_before=%d pending_after=%d stop=%s\n", runID, len(motions), recordsSeen, recordsChanged, pendingBefore, pendingAfter, stopReason)
	return nil
}

type motionDocumentResult struct {
	motion  motionDocumentCandidate
	outcome string
	bullets int
	changed bool
	err     error
}

func (ingest TweedeKamerMotionDocumentsIngest) processMotionCandidates(ctx context.Context, motions []motionDocumentCandidate) (int, int, error) {
	if len(motions) == 0 {
		return 0, 0, nil
	}

	workerCount := ingest.Concurrency
	if workerCount <= 0 {
		workerCount = 4
	}
	if workerCount > 16 {
		workerCount = 16
	}
	if workerCount > len(motions) {
		workerCount = len(motions)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan motionDocumentCandidate)
	results := make(chan motionDocumentResult)
	var wg sync.WaitGroup

	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for motion := range jobs {
				result := ingest.processMotionCandidate(ctx, motion)
				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
				if result.err != nil {
					cancel()
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, motion := range motions {
			select {
			case jobs <- motion:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	recordsSeen := 0
	recordsChanged := 0
	var firstErr error
	for result := range results {
		recordsSeen++
		if result.changed {
			recordsChanged++
		}
		if result.err != nil && firstErr == nil {
			firstErr = result.err
			continue
		}
		fmt.Printf("motion=%s outcome=%s bullets=%d\n", result.motion.MotionKey, result.outcome, result.bullets)
	}

	return recordsSeen, recordsChanged, firstErr
}

func (ingest TweedeKamerMotionDocumentsIngest) processMotionCandidate(ctx context.Context, motion motionDocumentCandidate) motionDocumentResult {
	result := motionDocumentResult{motion: motion}

	info, err := ingest.Client.FetchMotionDocumentInfo(ctx, motion.SourceID)
	if err != nil {
		result.err = err
		return result
	}

	documentURL := motionDocumentURL(info)
	if documentURL == "" {
		result.outcome = "no_document"
		result.changed, result.err = ingest.storeResult(ctx, motion, nil, nil)
		return result
	}

	xmlData, err := ingest.Documents.FetchDocument(ctx, documentURL)
	if errors.Is(err, officielebekendmakingen.ErrNotFound) {
		result.outcome = "not_published"
		result.changed, result.err = ingest.storeResult(ctx, motion, nil, &documentURL)
		return result
	}
	if err != nil {
		result.err = err
		return result
	}

	parsed, err := officielebekendmakingen.ExtractBulletPoints(xmlData)
	if err != nil {
		result.err = fmt.Errorf("parse document %s: %w", documentURL, err)
		return result
	}
	if parsed == nil {
		result.outcome = "not_a_motion"
		result.changed, result.err = ingest.storeResult(ctx, motion, nil, &documentURL)
		return result
	}

	result.outcome = "ok"
	result.bullets = len(parsed.BulletPoints)
	result.changed, result.err = ingest.storeResult(ctx, motion, parsed.BulletPoints, &documentURL)
	return result
}

// motionDocumentURL resolves the officielebekendmakingen URL for a motion.
// The zaak's own Document navigation holds the motion document; prefer the one
// whose volgnummer matches the zaak, mirroring the matching the old importer did
// against the dossier's documents.
func motionDocumentURL(info tweedekamer.MotionDocumentInfo) string {
	if len(info.Dossiers) == 0 {
		return ""
	}
	dossier := info.Dossiers[0]
	dossierNummer := dossier.Nummer.String()
	if dossierNummer == "" {
		return ""
	}

	document := pickMotionDocument(info)
	if document == nil || document.Volgnummer == nil || *document.Volgnummer <= 0 {
		return ""
	}

	toevoeging := ""
	if dossier.Toevoeging != nil {
		toevoeging = strings.TrimSpace(*dossier.Toevoeging)
	}
	return officielebekendmakingen.DocumentURL(dossierNummer, toevoeging, *document.Volgnummer)
}

func pickMotionDocument(info tweedekamer.MotionDocumentInfo) *tweedekamer.ZaakDocumentRecord {
	var fallback *tweedekamer.ZaakDocumentRecord
	for i := range info.Documents {
		document := &info.Documents[i]
		if document.Verwijderd != nil && *document.Verwijderd {
			continue
		}
		if info.Volgnummer != nil && document.Volgnummer != nil && *document.Volgnummer == *info.Volgnummer {
			return document
		}
		if fallback == nil && document.Soort != nil && strings.EqualFold(*document.Soort, "Motie") {
			fallback = document
		}
	}
	if fallback != nil {
		return fallback
	}
	for i := range info.Documents {
		document := &info.Documents[i]
		if document.Verwijderd == nil || !*document.Verwijderd {
			return document
		}
	}
	return nil
}

func (ingest TweedeKamerMotionDocumentsIngest) storeResult(ctx context.Context, motion motionDocumentCandidate, bulletPoints []string, documentURL *string) (bool, error) {
	var bulletJSON *string
	if len(bulletPoints) > 0 {
		encoded, err := json.Marshal(bulletPoints)
		if err != nil {
			return false, err
		}
		value := string(encoded)
		bulletJSON = &value
	}

	tag, err := ingest.Pool.Exec(ctx, `
		UPDATE motions
		SET bullet_points = $2::jsonb,
		    document_url = $3,
		    document_synced_at = now(),
		    updated_at = now()
		WHERE motion_key = $1
		  AND (bullet_points IS DISTINCT FROM $2::jsonb OR document_url IS DISTINCT FROM $3)
	`, motion.MotionKey, bulletJSON, documentURL)
	if err != nil {
		return false, err
	}
	changed := tag.RowsAffected() > 0
	if !changed {
		_, err = ingest.Pool.Exec(ctx, `
			UPDATE motions
			SET document_synced_at = now()
			WHERE motion_key = $1
		`, motion.MotionKey)
		if err != nil {
			return false, err
		}
	}
	return changed, nil
}

type motionDocumentCandidate struct {
	MotionKey string
	SourceID  string
}

func (ingest TweedeKamerMotionDocumentsIngest) motionCandidates(ctx context.Context, resyncBefore, proposedAfter *time.Time) ([]motionDocumentCandidate, error) {
	rows, err := ingest.Pool.Query(ctx, `
		SELECT motion_key, source_id
		FROM motions
		WHERE source_key = $1
		  AND source_deleted = false
		  AND (document_synced_at IS NULL
		       OR ($3::timestamptz IS NOT NULL AND document_synced_at < $3)
		       OR ($4::timestamptz IS NOT NULL AND bullet_points IS NULL AND proposed_at > $4))
		ORDER BY document_synced_at ASC NULLS FIRST, proposed_at DESC NULLS LAST
		LIMIT $2
	`, tweedeKamerSourceKey, ingest.Limit, resyncBefore, proposedAfter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var motions []motionDocumentCandidate
	for rows.Next() {
		var motion motionDocumentCandidate
		if err := rows.Scan(&motion.MotionKey, &motion.SourceID); err != nil {
			return nil, err
		}
		motions = append(motions, motion)
	}
	return motions, rows.Err()
}

func (ingest TweedeKamerMotionDocumentsIngest) pendingCount(ctx context.Context, resyncBefore, proposedAfter *time.Time) (int, error) {
	var count int
	err := ingest.Pool.QueryRow(ctx, `
		SELECT count(*)::int
		FROM motions
		WHERE source_key = $1
		  AND source_deleted = false
		  AND (document_synced_at IS NULL
		       OR ($2::timestamptz IS NOT NULL AND document_synced_at < $2)
		       OR ($3::timestamptz IS NOT NULL AND bullet_points IS NULL AND proposed_at > $3))
	`, tweedeKamerSourceKey, resyncBefore, proposedAfter).Scan(&count)
	return count, err
}

func (ingest TweedeKamerMotionDocumentsIngest) resyncBefore() *time.Time {
	if ingest.ResyncAfter <= 0 {
		return nil
	}
	value := time.Now().Add(-ingest.ResyncAfter)
	return &value
}

// proposedAfter is the earliest proposal date for which a still-blank motion is
// worth retrying. Past it the document counts as permanently absent, which stops
// the long tail of non-motion documents from being refetched forever.
func (ingest TweedeKamerMotionDocumentsIngest) proposedAfter() *time.Time {
	if ingest.ResyncGrace <= 0 {
		return nil
	}
	value := time.Now().Add(-ingest.ResyncGrace)
	return &value
}
