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

const (
	obvProvider             = "openbesluitvorming"
	obvMotionsPipeline      = "obv.motions"
	obvDocumentsPipeline    = "obv.motion-documents"
	obvMotionCollection     = "Motion"
	obvMotionTextCollection = "MotionText"
	obvDerivedPartyRaw      = "DerivedFromVotes"
)

func OpenBesluitvormingSourceKey(obvKey string) string {
	return obvProvider + "-" + obvKey
}

func OpenBesluitvormingJurisdictionKey(obvKey string) string {
	return "nl-gemeente-" + obvKey
}

type obvSource struct {
	OBVKey          string
	SourceKey       string
	JurisdictionKey string
	Label           string
}

func ensureOBVSource(ctx context.Context, pool *pgxpool.Pool, client *openbesluitvorming.Client, obvKey string) (obvSource, error) {
	result := obvSource{
		OBVKey:          obvKey,
		SourceKey:       OpenBesluitvormingSourceKey(obvKey),
		JurisdictionKey: OpenBesluitvormingJurisdictionKey(obvKey),
	}

	err := pool.QueryRow(ctx, `
		SELECT j.name
		FROM data_sources d
		JOIN jurisdictions j ON j.jurisdiction_key = d.jurisdiction_key
		WHERE d.source_key = $1 AND d.enabled = true
	`, result.SourceKey).Scan(&result.Label)
	if err == nil {
		return result, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return obvSource{}, err
	}

	sources, err := client.FetchSources(ctx)
	if err != nil {
		return obvSource{}, fmt.Errorf("list openbesluitvorming sources: %w", err)
	}
	var source *openbesluitvorming.Source
	for i := range sources {
		if sources[i].Key == obvKey {
			source = &sources[i]
			break
		}
	}
	switch {
	case source == nil:
		return obvSource{}, fmt.Errorf("openbesluitvorming has no source %q; see https://openbesluitvorming.nl/api/sources", obvKey)
	case source.IsAggregate:
		return obvSource{}, fmt.Errorf("openbesluitvorming source %q is an aggregate of a whole supplier, not one council", obvKey)
	case source.OrganizationType != "gemeente":
		return obvSource{}, fmt.Errorf("openbesluitvorming source %q is a %s; only gemeenten are supported", obvKey, source.OrganizationType)
	}
	result.Label = source.Label

	var cbsID *string
	if statuses, err := client.FetchStatus(ctx); err == nil {
		for _, status := range statuses {
			if status.SourceKey == obvKey && status.CBSID != "" {
				value := status.CBSID
				cbsID = &value
				break
			}
		}
	}

	config, err := json.Marshal(map[string]string{
		"obvSource": source.Key,
		"sourceRef": source.SourceRef,
		"supplier":  source.Supplier,
	})
	if err != nil {
		return obvSource{}, err
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return obvSource{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO jurisdictions (jurisdiction_key, kind, name, country_code, municipality_code)
		VALUES ($1, 'municipality', $2, 'NL', $3)
		ON CONFLICT (jurisdiction_key) DO UPDATE
		SET name = EXCLUDED.name,
		    municipality_code = COALESCE(EXCLUDED.municipality_code, jurisdictions.municipality_code),
		    updated_at = now()
	`, result.JurisdictionKey, result.Label, cbsID); err != nil {
		return obvSource{}, fmt.Errorf("register jurisdiction %s: %w", result.JurisdictionKey, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO data_sources (source_key, jurisdiction_key, provider, name, base_url, config)
		VALUES ($1, $2, $3, $4, 'https://openbesluitvorming.nl', $5)
		ON CONFLICT (source_key) DO UPDATE
		SET config = EXCLUDED.config,
		    enabled = true,
		    updated_at = now()
	`, result.SourceKey, result.JurisdictionKey, obvProvider, "Open Besluitvorming "+result.Label, string(config)); err != nil {
		return obvSource{}, fmt.Errorf("register data source %s: %w", result.SourceKey, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return obvSource{}, err
	}

	return result, nil
}

type obvMotionCursor struct {
	Phase          string `json:"phase"`
	SnapshotCursor string `json:"snapshotCursor,omitempty"`
	ChangesCursor  string `json:"changesCursor,omitempty"`
}

const (
	obvPhaseSnapshot       = "snapshot"
	obvPhaseChanges        = "changes"
	obvMotionSnapshotStart = "motion"
)

type OpenBesluitvormingMotionIngest struct {
	Pool        *pgxpool.Pool
	Client      *openbesluitvorming.Client
	Source      string
	PageSize    int
	MaxPages    int
	ResetCursor bool
}

func (ingest OpenBesluitvormingMotionIngest) Run(ctx context.Context) error {
	source, err := ensureOBVSource(ctx, ingest.Pool, ingest.Client, ingest.Source)
	if err != nil {
		return err
	}

	release, err := obvLock(ctx, ingest.Pool, source.SourceKey, obvMotionsPipeline)
	if err != nil {
		return err
	}
	defer release()

	if ingest.ResetCursor {
		if _, err := ingest.Pool.Exec(ctx, `DELETE FROM source_cursors WHERE source_key = $1 AND pipeline = $2`, source.SourceKey, obvMotionsPipeline); err != nil {
			return err
		}
	}

	cursor := obvMotionCursor{Phase: obvPhaseSnapshot, SnapshotCursor: obvMotionSnapshotStart}
	if found, err := obvLoadCursor(ctx, ingest.Pool, source.SourceKey, obvMotionsPipeline, &cursor); err != nil {
		return err
	} else if !found {
		cursor = obvMotionCursor{Phase: obvPhaseSnapshot, SnapshotCursor: obvMotionSnapshotStart}
	}
	cursorBefore := cursor

	runID, err := obvStartRun(ctx, ingest.Pool, source.SourceKey, obvMotionsPipeline, cursorBefore)
	if err != nil {
		return err
	}

	fail := func(err error, seen, changed int) error {
		_ = obvFinishRun(ctx, ingest.Pool, runID, "failed", cursor, seen, changed, false, "error", err.Error())
		return err
	}

	seen, changed, deleted, pages := 0, 0, 0, 0
	stopReason := ""
	for {
		if ingest.MaxPages > 0 && pages >= ingest.MaxPages {
			stopReason = "max_pages"
			break
		}

		switch cursor.Phase {
		case obvPhaseSnapshot:
			page, err := ingest.Client.FetchSnapshot(ctx, source.OBVKey, cursor.SnapshotCursor, ingest.PageSize)
			if err != nil {
				return fail(err, seen, changed)
			}
			if cursor.ChangesCursor == "" {
				cursor.ChangesCursor = page.ChangesCursor
			}

			pastMotions := false
			for _, record := range page.Records {
				if record.EntityType != obvMotionCollection {
					pastMotions = true
					continue
				}
				seen++
				recordChanged, err := ingest.storeMotion(ctx, source, record)
				if err != nil {
					return fail(fmt.Errorf("store %s: %w", record.EntityID, err), seen, changed)
				}
				if recordChanged {
					changed++
				}
			}
			pages++

			if pastMotions || !page.HasMore || page.NextCursor == "" {
				cursor.Phase = obvPhaseChanges
				cursor.SnapshotCursor = ""
			} else {
				cursor.SnapshotCursor = page.NextCursor
			}
			fmt.Printf("source=%s phase=snapshot page=%d seen=%d changed=%d\n", source.OBVKey, pages, seen, changed)

		case obvPhaseChanges:
			page, err := ingest.Client.FetchChanges(ctx, source.OBVKey, cursor.ChangesCursor, ingest.PageSize)
			if err != nil {
				return fail(err, seen, changed)
			}
			for _, record := range page.Records {
				if record.EntityType != obvMotionCollection {
					continue
				}
				seen++
				if record.Op == "delete" {
					if err := ingest.deleteMotion(ctx, source, record); err != nil {
						return fail(fmt.Errorf("delete %s: %w", record.EntityID, err), seen, changed)
					}
					deleted++
					changed++
					continue
				}
				recordChanged, err := ingest.storeMotion(ctx, source, record)
				if err != nil {
					return fail(fmt.Errorf("store %s: %w", record.EntityID, err), seen, changed)
				}
				if recordChanged {
					changed++
				}
			}
			pages++
			if page.NextCursor != "" {
				cursor.ChangesCursor = page.NextCursor
			}
			fmt.Printf("source=%s phase=changes page=%d records=%d seen=%d changed=%d cursor=%s\n", source.OBVKey, pages, len(page.Records), seen, changed, cursor.ChangesCursor)
			if !page.HasMore {
				stopReason = "complete"
			}

		default:
			return fail(fmt.Errorf("unknown cursor phase %q", cursor.Phase), seen, changed)
		}

		if err := obvSaveCursor(ctx, ingest.Pool, source.SourceKey, obvMotionsPipeline, cursor); err != nil {
			return fail(err, seen, changed)
		}
		if stopReason != "" {
			break
		}
	}

	parties, err := deriveOBVParties(ctx, ingest.Pool, source)
	if err != nil {
		return fail(fmt.Errorf("derive parties: %w", err), seen, changed)
	}

	if err := obvFinishRun(ctx, ingest.Pool, runID, "succeeded", cursor, seen, changed, true, stopReason, ""); err != nil {
		return err
	}
	fmt.Printf("ingestion complete source=%s run_id=%d pages=%d seen=%d changed=%d deleted=%d parties=%d phase=%s stop_reason=%s\n",
		source.SourceKey, runID, pages, seen, changed, deleted, parties, cursor.Phase, stopReason)
	return nil
}

func (ingest OpenBesluitvormingMotionIngest) storeMotion(ctx context.Context, source obvSource, record openbesluitvorming.ExportRecord) (bool, error) {
	var payload openbesluitvorming.MotionPayload
	if err := json.Unmarshal(record.Payload, &payload); err != nil {
		return false, fmt.Errorf("decode payload: %w", err)
	}

	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	payloadHash := record.ContentHash
	if payloadHash == "" {
		payloadHash = hashBytes(record.Payload)
	}
	rawChanged, err := obvStoreRaw(ctx, tx, source.SourceKey, obvMotionCollection, record.EntityID, record.Time, false, record.Payload, payloadHash)
	if err != nil {
		return false, err
	}

	motionKey := obvMotionKey(source, record.EntityID)
	number, title := payload.NumberAndTitle()
	if _, err := tx.Exec(ctx, `
		INSERT INTO motions (
			motion_key, source_key, jurisdiction_key, source_id,
			number, title, status, kind, proposed_at,
			source_updated_at, source_deleted, raw_collection, votes_synced_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, false, $11, now(), now())
		ON CONFLICT (source_key, source_id)
		DO UPDATE SET number = EXCLUDED.number,
		              title = EXCLUDED.title,
		              status = EXCLUDED.status,
		              kind = EXCLUDED.kind,
		              proposed_at = EXCLUDED.proposed_at,
		              source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = false,
		              votes_synced_at = now(),
		              updated_at = now()
	`, motionKey, source.SourceKey, source.JurisdictionKey, record.EntityID,
		nullIfEmpty(number), nullIfEmpty(title), nullIfEmpty(payload.Status), nullIfEmpty(payload.Kind()), payload.ProposedAt(),
		record.Time, obvMotionCollection); err != nil {
		return false, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM votes WHERE motion_key = $1`, motionKey); err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM decisions WHERE motion_key = $1`, motionKey); err != nil {
		return false, err
	}

	votes := payload.UniqueVotes()
	decisionKey := source.SourceKey + ":decision:" + record.EntityID
	var votingType *string
	if len(votes) > 0 {
		value := "Hoofdelijk"
		votingType = &value
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO decisions (
			decision_key, source_key, motion_key, source_id,
			voting_type, decision_type, decision_text, comment, status, decision_order,
			source_updated_at, source_deleted, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 1, $10, false, now())
	`, decisionKey, source.SourceKey, motionKey, record.EntityID+"#decision",
		votingType, nullIfEmpty(payload.Outcome()), nullIfEmpty(payload.Status), nullIfEmpty(payload.VoteSummary), nullIfEmpty(payload.Result),
		record.Time); err != nil {
		return false, err
	}

	if len(votes) > 0 {
		batch := &pgx.Batch{}
		for _, vote := range votes {
			voteType := "Voor"
			if vote.Option == "tegen" {
				voteType = "Tegen"
			}
			var partySourceID *string
			if key := openbesluitvorming.PartyKey(vote.GroupName, source.Label); key != "" {
				value := "party:" + key
				partySourceID = &value
			}
			voteSourceID := record.EntityID + "#" + vote.VoterKey()
			batch.Queue(`
				INSERT INTO votes (
					vote_key, source_key, motion_key, decision_key, source_id,
					vote_type, party_source_id, party_name, actor_name, party_size,
					mistake, person_source_id, source_updated_at, source_deleted, updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL, false, $10, $11, false, now())
			`, source.SourceKey+":vote:"+voteSourceID, source.SourceKey, motionKey, decisionKey, voteSourceID,
				voteType, partySourceID, nullIfEmpty(vote.GroupName), nullIfEmpty(vote.VoterName),
				vote.VoterKey(), record.Time)
		}
		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return rawChanged, nil
}

func (ingest OpenBesluitvormingMotionIngest) deleteMotion(ctx context.Context, source obvSource, record openbesluitvorming.ExportRecord) error {
	motionKey := obvMotionKey(source, record.EntityID)
	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, statement := range []string{
		`UPDATE raw_records SET source_deleted = true, source_updated_at = $3, last_seen_at = now() WHERE source_key = $1 AND collection = 'Motion' AND source_id = $2`,
		`UPDATE motions SET source_deleted = true, source_updated_at = $3, updated_at = now() WHERE source_key = $1 AND source_id = $2`,
	} {
		if _, err := tx.Exec(ctx, statement, source.SourceKey, record.EntityID, record.Time); err != nil {
			return err
		}
	}
	for _, statement := range []string{
		`UPDATE decisions SET source_deleted = true, updated_at = now() WHERE motion_key = $1`,
		`UPDATE votes SET source_deleted = true, updated_at = now() WHERE motion_key = $1`,
	} {
		if _, err := tx.Exec(ctx, statement, motionKey); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func deriveOBVParties(ctx context.Context, pool *pgxpool.Pool, source obvSource) (int, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		WITH v AS (
			SELECT v.party_source_id, v.party_name, v.person_source_id, v.motion_key, m.proposed_at
			FROM votes v
			JOIN motions m ON m.motion_key = v.motion_key
			WHERE v.source_key = $1
			  AND v.source_deleted = false
			  AND m.source_deleted = false
			  AND v.party_source_id IS NOT NULL
			  AND m.proposed_at IS NOT NULL
		),
		names AS (
			SELECT party_source_id,
			       party_name,
			       row_number() OVER (PARTITION BY party_source_id ORDER BY count(*) DESC, max(proposed_at) DESC, party_name) AS rn
			FROM v
			GROUP BY party_source_id, party_name
		),
		per_motion AS (
			SELECT party_source_id, motion_key, max(proposed_at) AS voted_at, count(DISTINCT person_source_id) AS members
			FROM v
			GROUP BY party_source_id, motion_key
		),
		spans AS (
			SELECT party_source_id, min(voted_at) AS first_at, max(voted_at) AS last_at
			FROM per_motion
			GROUP BY party_source_id
		),
		seats AS (
			SELECT pm.party_source_id, max(pm.members)::int AS seats
			FROM per_motion pm
			JOIN spans s ON s.party_source_id = pm.party_source_id
			WHERE pm.voted_at >= s.last_at - interval '90 days'
			GROUP BY pm.party_source_id
		),
		latest AS (
			SELECT max(voted_at) AS voted_at FROM per_motion
		)
		INSERT INTO parties (
			party_key, source_key, jurisdiction_key, source_id,
			short_name, name, seats, active_from, active_to,
			source_updated_at, source_deleted, raw_collection, updated_at
		)
		SELECT $1 || ':' || s.party_source_id, $1, $2, s.party_source_id,
		       n.party_name, n.party_name, se.seats, s.first_at::date,
		       CASE WHEN s.last_at < l.voted_at - interval '60 days' THEN s.last_at::date END,
		       s.last_at, false, $3, now()
		FROM spans s
		JOIN names n ON n.party_source_id = s.party_source_id AND n.rn = 1
		JOIN seats se ON se.party_source_id = s.party_source_id
		CROSS JOIN latest l
		ON CONFLICT (source_key, source_id)
		DO UPDATE SET short_name = EXCLUDED.short_name,
		              name = EXCLUDED.name,
		              seats = EXCLUDED.seats,
		              active_from = EXCLUDED.active_from,
		              active_to = EXCLUDED.active_to,
		              source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = false,
		              updated_at = now()
		RETURNING source_id
	`, source.SourceKey, source.JurisdictionKey, obvDerivedPartyRaw)
	if err != nil {
		return 0, err
	}
	var current []string
	for rows.Next() {
		var sourceID string
		if err := rows.Scan(&sourceID); err != nil {
			rows.Close()
			return 0, err
		}
		current = append(current, sourceID)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE parties
		SET source_deleted = true, updated_at = now()
		WHERE source_key = $1
		  AND source_deleted = false
		  AND NOT (source_id = ANY($2::text[]))
	`, source.SourceKey, current); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(current), nil
}

func obvMotionKey(source obvSource, entityID string) string {
	return source.SourceKey + ":" + entityID
}

func obvStoreRaw(ctx context.Context, tx pgx.Tx, sourceKey, collection, sourceID string, updatedAt *time.Time, deleted bool, payload []byte, payloadHash string) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO raw_records (source_key, collection, source_id, source_updated_at, source_deleted, payload, payload_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_key, collection, source_id)
		DO UPDATE SET source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              payload = EXCLUDED.payload,
		              payload_hash = EXCLUDED.payload_hash,
		              last_seen_at = now()
		WHERE raw_records.payload_hash IS DISTINCT FROM EXCLUDED.payload_hash
		   OR raw_records.source_deleted IS DISTINCT FROM EXCLUDED.source_deleted
	`, sourceKey, collection, sourceID, updatedAt, deleted, string(payload), payloadHash)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func obvLock(ctx context.Context, pool *pgxpool.Pool, sourceKey, pipeline string) (func(), error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	lockKey := sourceKey + ":" + pipeline
	var acquired bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", lockKey).Scan(&acquired); err != nil {
		conn.Release()
		return nil, err
	}
	if !acquired {
		conn.Release()
		return nil, fmt.Errorf("ingestion pipeline already running: %s/%s", sourceKey, pipeline)
	}
	return func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", lockKey)
		conn.Release()
	}, nil
}

func obvLoadCursor(ctx context.Context, pool *pgxpool.Pool, sourceKey, pipeline string, target any) (bool, error) {
	var raw []byte
	err := pool.QueryRow(ctx, `SELECT cursor FROM source_cursors WHERE source_key = $1 AND pipeline = $2`, sourceKey, pipeline).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(raw, target)
}

func obvSaveCursor(ctx context.Context, pool *pgxpool.Pool, sourceKey, pipeline string, cursor any) error {
	raw, err := json.Marshal(cursor)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO source_cursors (source_key, pipeline, cursor, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (source_key, pipeline)
		DO UPDATE SET cursor = EXCLUDED.cursor, updated_at = now()
	`, sourceKey, pipeline, string(raw))
	return err
}

func obvStartRun(ctx context.Context, pool *pgxpool.Pool, sourceKey, pipeline string, cursorBefore any) (int64, error) {
	raw, err := json.Marshal(cursorBefore)
	if err != nil {
		return 0, err
	}
	var runID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO ingestion_runs (source_key, pipeline, status, cursor_before)
		VALUES ($1, $2, 'running', $3)
		RETURNING id
	`, sourceKey, pipeline, string(raw)).Scan(&runID)
	return runID, err
}

func obvFinishRun(ctx context.Context, pool *pgxpool.Pool, runID int64, status string, cursorAfter any, seen, changed int, cursorSaved bool, stopReason, errorMessage string) error {
	raw, err := json.Marshal(cursorAfter)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		UPDATE ingestion_runs
		SET status = $2, cursor_after = $3, records_seen = $4, records_changed = $5,
		    error_message = $6, cursor_saved = $7, stop_reason = $8, finished_at = now()
		WHERE id = $1
	`, runID, status, string(raw), seen, changed, nullIfEmpty(errorMessage), cursorSaved, stopReason)
	return err
}

func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
