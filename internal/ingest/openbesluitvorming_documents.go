package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/source/openbesluitvorming"
)

type OpenBesluitvormingMotionDocumentsIngest struct {
	Pool    *pgxpool.Pool
	Client  *openbesluitvorming.Client
	Source  string
	Limit   int
	Reparse bool
}

type obvMotionText struct {
	MarkdownText string `json:"markdownText"`
	PDFURL       string `json:"pdfUrl,omitempty"`
	DownloadURL  string `json:"downloadUrl,omitempty"`
	PDFEntityID  string `json:"pdfEntityId,omitempty"`
}

func (text obvMotionText) documentURL() string {
	if text.PDFURL != "" {
		return text.PDFURL
	}
	return text.DownloadURL
}

func (ingest OpenBesluitvormingMotionDocumentsIngest) Run(ctx context.Context) error {
	source, err := ensureOBVSource(ctx, ingest.Pool, ingest.Client, ingest.Source)
	if err != nil {
		return err
	}

	release, err := obvLock(ctx, ingest.Pool, source.SourceKey, obvDocumentsPipeline)
	if err != nil {
		return err
	}
	defer release()

	if ingest.Reparse {
		return ingest.reparse(ctx, source)
	}

	candidates, err := ingest.loadCandidates(ctx, source)
	if err != nil {
		return err
	}

	runID, err := obvStartRun(ctx, ingest.Pool, source.SourceKey, obvDocumentsPipeline, map[string]any{})
	if err != nil {
		return err
	}

	seen, changed, withBullets, missing, consecutiveErrors := 0, 0, 0, 0, 0
	started := time.Now()
	for _, motion := range candidates {
		seen++
		entity, err := ingest.Client.FetchEntity(ctx, motion.SourceID)
		switch {
		case errors.Is(err, openbesluitvorming.ErrEntityNotFound), errors.Is(err, openbesluitvorming.ErrEntityRemoved):
			missing++
			if _, err := ingest.Pool.Exec(ctx, `UPDATE motions SET document_synced_at = now() WHERE motion_key = $1`, motion.MotionKey); err != nil {
				return obvFailDocuments(ctx, ingest.Pool, runID, seen, changed, err)
			}
			continue
		case err != nil:
			if ctx.Err() != nil {
				return obvFailDocuments(ctx, ingest.Pool, runID, seen, changed, ctx.Err())
			}
			consecutiveErrors++
			fmt.Printf("motion=%s fetch failed: %v\n", motion.SourceID, err)
			if consecutiveErrors >= 5 {
				return obvFailDocuments(ctx, ingest.Pool, runID, seen, changed, fmt.Errorf("5 consecutive fetch failures, last: %w", err))
			}
			continue
		}
		consecutiveErrors = 0

		text := obvMotionText{
			MarkdownText: entity.MarkdownText,
			PDFURL:       entity.PDFURL,
			DownloadURL:  entity.DownloadURL,
			PDFEntityID:  entity.PDFEntityID,
		}
		bullets, recordChanged, err := ingest.storeText(ctx, source, motion, text)
		if err != nil {
			return obvFailDocuments(ctx, ingest.Pool, runID, seen, changed, err)
		}
		if recordChanged {
			changed++
		}
		if bullets > 0 {
			withBullets++
		}

		if seen%50 == 0 {
			fmt.Printf("source=%s fetched=%d/%d with_bullets=%d elapsed=%s\n", source.OBVKey, seen, len(candidates), withBullets, time.Since(started).Round(time.Second))
		}
	}

	if err := obvFinishRun(ctx, ingest.Pool, runID, "succeeded", map[string]any{}, seen, changed, false, "complete", ""); err != nil {
		return err
	}
	fmt.Printf("motion documents complete source=%s run_id=%d fetched=%d with_bullets=%d missing=%d changed=%d elapsed=%s\n",
		source.SourceKey, runID, seen, withBullets, missing, changed, time.Since(started).Round(time.Second))
	return nil
}

type obvDocumentCandidate struct {
	MotionKey string
	SourceID  string
}

func (ingest OpenBesluitvormingMotionDocumentsIngest) loadCandidates(ctx context.Context, source obvSource) ([]obvDocumentCandidate, error) {
	rows, err := ingest.Pool.Query(ctx, `
		SELECT m.motion_key, m.source_id
		FROM motions m
		JOIN raw_records r ON r.source_key = m.source_key
		                  AND r.collection = $2
		                  AND r.source_id = m.source_id
		WHERE m.source_key = $1
		  AND m.source_deleted = false
		  AND m.document_synced_at IS NULL
		  AND m.kind = 'Motie'
		  AND jsonb_typeof(r.payload->'attachment') = 'array'
		  AND jsonb_array_length(r.payload->'attachment') > 0
		  AND EXISTS (
		    SELECT 1 FROM votes v WHERE v.motion_key = m.motion_key AND v.source_deleted = false
		  )
		ORDER BY m.proposed_at DESC NULLS LAST, m.motion_key
		LIMIT $3
	`, source.SourceKey, obvMotionCollection, ingest.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []obvDocumentCandidate
	for rows.Next() {
		var candidate obvDocumentCandidate
		if err := rows.Scan(&candidate.MotionKey, &candidate.SourceID); err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func (ingest OpenBesluitvormingMotionDocumentsIngest) storeText(ctx context.Context, source obvSource, motion obvDocumentCandidate, text obvMotionText) (int, bool, error) {
	payload, err := json.Marshal(text)
	if err != nil {
		return 0, false, err
	}

	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	if _, err := obvStoreRaw(ctx, tx, source.SourceKey, obvMotionTextCollection, motion.SourceID, &now, false, payload, hashBytes(payload)); err != nil {
		return 0, false, err
	}
	bullets, changed, err := obvStoreBullets(ctx, tx, motion.MotionKey, text)
	if err != nil {
		return 0, false, err
	}
	return bullets, changed, tx.Commit(ctx)
}

func obvStoreBullets(ctx context.Context, tx pgx.Tx, motionKey string, text obvMotionText) (int, bool, error) {
	bullets := openbesluitvorming.ExtractBulletPoints(text.MarkdownText)
	var bulletJSON *string
	if len(bullets) > 0 {
		encoded, err := json.Marshal(bullets)
		if err != nil {
			return 0, false, err
		}
		value := string(encoded)
		bulletJSON = &value
	}

	documentURL := nullIfEmpty(text.documentURL())
	tag, err := tx.Exec(ctx, `
		UPDATE motions
		SET bullet_points = $2::jsonb,
		    document_url = $3,
		    document_synced_at = now(),
		    updated_at = now()
		WHERE motion_key = $1
		  AND (bullet_points IS DISTINCT FROM $2::jsonb OR document_url IS DISTINCT FROM $3)
	`, motionKey, bulletJSON, documentURL)
	if err != nil {
		return 0, false, err
	}
	changed := tag.RowsAffected() > 0
	if !changed {
		if _, err := tx.Exec(ctx, `UPDATE motions SET document_synced_at = now() WHERE motion_key = $1`, motionKey); err != nil {
			return 0, false, err
		}
	}
	return len(bullets), changed, nil
}

func (ingest OpenBesluitvormingMotionDocumentsIngest) reparse(ctx context.Context, source obvSource) error {
	rows, err := ingest.Pool.Query(ctx, `
		SELECT m.motion_key, r.payload
		FROM raw_records r
		JOIN motions m ON m.source_key = r.source_key AND m.source_id = r.source_id
		WHERE r.source_key = $1 AND r.collection = $2
	`, source.SourceKey, obvMotionTextCollection)
	if err != nil {
		return err
	}
	type stored struct {
		motionKey string
		text      obvMotionText
	}
	var texts []stored
	for rows.Next() {
		var item stored
		var payload []byte
		if err := rows.Scan(&item.motionKey, &payload); err != nil {
			rows.Close()
			return err
		}
		if err := json.Unmarshal(payload, &item.text); err != nil {
			rows.Close()
			return fmt.Errorf("decode stored text of %s: %w", item.motionKey, err)
		}
		texts = append(texts, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	withBullets := 0
	for _, item := range texts {
		bullets, _, err := obvStoreBullets(ctx, tx, item.motionKey, item.text)
		if err != nil {
			return err
		}
		if bullets > 0 {
			withBullets++
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	fmt.Printf("reparse complete source=%s texts=%d with_bullets=%d\n", source.SourceKey, len(texts), withBullets)
	return nil
}

func obvFailDocuments(ctx context.Context, pool *pgxpool.Pool, runID int64, seen, changed int, err error) error {
	_ = obvFinishRun(ctx, pool, runID, "failed", map[string]any{}, seen, changed, false, "error", err.Error())
	return err
}
