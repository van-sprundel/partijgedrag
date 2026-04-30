package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/rewrite/internal/source/tweedekamer"
)

const (
	motionVotesPipeline = "motion_votes.raw"
	besluitCollection   = "Besluit"
	stemmingCollection  = "Stemming"
)

type TweedeKamerMotionVotesIngest struct {
	Pool        *pgxpool.Pool
	Client      *tweedekamer.Client
	Limit       int
	Concurrency int
	ResyncAfter time.Duration
}

func (ingest TweedeKamerMotionVotesIngest) Run(ctx context.Context) error {
	releaseLock, err := acquirePipelineLock(ctx, ingest.Pool, motionVotesPipeline)
	if err != nil {
		return err
	}
	defer releaseLock()

	runID, err := startPipelineRun(ctx, ingest.Pool, motionVotesPipeline)
	if err != nil {
		return err
	}

	resyncBefore := ingest.resyncBefore()
	pendingBefore, err := ingest.pendingCount(ctx, resyncBefore)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionVotesPipeline, "failed", 0, 0, false, "error", err.Error())
		return err
	}

	motions, err := ingest.motionCandidates(ctx, resyncBefore)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionVotesPipeline, "failed", 0, 0, false, "error", err.Error())
		return err
	}

	recordsSeen, recordsChanged, err := ingest.processMotionCandidates(ctx, motions)
	if err != nil {
		_ = finishPipelineRun(ctx, ingest.Pool, runID, motionVotesPipeline, "failed", recordsSeen, recordsChanged, false, "error", err.Error())
		return err
	}

	if err := finishPipelineRun(ctx, ingest.Pool, runID, motionVotesPipeline, "succeeded", recordsSeen, recordsChanged, false, "complete", ""); err != nil {
		return err
	}

	pendingAfter, err := ingest.pendingCount(ctx, resyncBefore)
	if err != nil {
		return err
	}

	fmt.Printf("motion vote ingestion complete run_id=%d motions=%d seen=%d changed=%d pending_before=%d pending_after=%d\n", runID, len(motions), recordsSeen, recordsChanged, pendingBefore, pendingAfter)
	return nil
}

type motionVoteResult struct {
	motion          motionCandidate
	decisionsSeen   int
	votesSeen       int
	decisionChanges int
	voteChanges     int
	err             error
}

func (ingest TweedeKamerMotionVotesIngest) processMotionCandidates(ctx context.Context, motions []motionCandidate) (int, int, error) {
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

	jobs := make(chan motionCandidate)
	results := make(chan motionVoteResult)
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
		recordsSeen += result.decisionsSeen + result.votesSeen
		recordsChanged += result.decisionChanges + result.voteChanges
		if result.err != nil && firstErr == nil {
			firstErr = result.err
			continue
		}
		fmt.Printf("motion=%s decisions=%d votes=%d changed=%d\n", result.motion.MotionKey, result.decisionsSeen, result.votesSeen, result.decisionChanges+result.voteChanges)
	}

	return recordsSeen, recordsChanged, firstErr
}

func (ingest TweedeKamerMotionVotesIngest) processMotionCandidate(ctx context.Context, motion motionCandidate) motionVoteResult {
	result := motionVoteResult{motion: motion}

	decisions, err := ingest.Client.FetchMotionDecisions(ctx, motion.SourceID)
	if err != nil {
		result.err = err
		return result
	}
	result.decisionsSeen = len(decisions)

	decisionChanges, voteSeen, voteChanges, err := ingest.storeMotionDecisionsAndVotes(ctx, motion, decisions)
	if err != nil {
		result.err = err
		return result
	}
	result.votesSeen = voteSeen
	result.decisionChanges = decisionChanges
	result.voteChanges = voteChanges
	return result
}

type motionCandidate struct {
	MotionKey string
	SourceID  string
}

func (ingest TweedeKamerMotionVotesIngest) motionCandidates(ctx context.Context, resyncBefore *time.Time) ([]motionCandidate, error) {
	rows, err := ingest.Pool.Query(ctx, `
		SELECT motion_key, source_id
		FROM motions
		WHERE source_key = $1
		  AND source_deleted = false
		  AND (votes_synced_at IS NULL OR ($3::timestamptz IS NOT NULL AND votes_synced_at < $3))
		ORDER BY votes_synced_at ASC NULLS FIRST, proposed_at DESC NULLS LAST
		LIMIT $2
	`, tweedeKamerSourceKey, ingest.Limit, resyncBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var motions []motionCandidate
	for rows.Next() {
		var motion motionCandidate
		if err := rows.Scan(&motion.MotionKey, &motion.SourceID); err != nil {
			return nil, err
		}
		motions = append(motions, motion)
	}
	return motions, rows.Err()
}

func (ingest TweedeKamerMotionVotesIngest) pendingCount(ctx context.Context, resyncBefore *time.Time) (int, error) {
	var count int
	err := ingest.Pool.QueryRow(ctx, `
		SELECT count(*)::int
		FROM motions
		WHERE source_key = $1
		  AND source_deleted = false
		  AND (votes_synced_at IS NULL OR ($2::timestamptz IS NOT NULL AND votes_synced_at < $2))
	`, tweedeKamerSourceKey, resyncBefore).Scan(&count)
	return count, err
}

func (ingest TweedeKamerMotionVotesIngest) resyncBefore() *time.Time {
	if ingest.ResyncAfter <= 0 {
		return nil
	}
	value := time.Now().Add(-ingest.ResyncAfter)
	return &value
}

func (ingest TweedeKamerMotionVotesIngest) storeMotionDecisionsAndVotes(
	ctx context.Context,
	motion motionCandidate,
	decisions []tweedekamer.DecisionRecord,
) (int, int, int, error) {
	tx, err := ingest.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, 0, err
	}
	defer tx.Rollback(ctx)

	decisionChanges := 0
	votesSeen := 0
	voteChanges := 0

	for _, decision := range decisions {
		changed, err := storeDecision(ctx, tx, motion, decision)
		if err != nil {
			return 0, 0, 0, err
		}
		if changed {
			decisionChanges++
		}

		votes, err := ingest.Client.FetchDecisionVotes(ctx, decision.ID)
		if err != nil {
			return 0, 0, 0, err
		}
		votesSeen += len(votes)

		for _, vote := range votes {
			changed, err := storeVote(ctx, tx, motion, decision, vote)
			if err != nil {
				return 0, 0, 0, err
			}
			if changed {
				voteChanges++
			}
		}
	}

	if _, err := tx.Exec(ctx, "UPDATE motions SET votes_synced_at = now(), updated_at = now() WHERE motion_key = $1", motion.MotionKey); err != nil {
		return 0, 0, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, 0, err
	}

	return decisionChanges, votesSeen, voteChanges, nil
}

func storeDecision(ctx context.Context, tx pgx.Tx, motion motionCandidate, decision tweedekamer.DecisionRecord) (bool, error) {
	raw := projectDecisionRaw(decision)
	projected := projectDecision(motion, decision)

	rawChanged, err := storeRawRecord(ctx, tx, raw)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO decisions (
			decision_key,
			source_key,
			motion_key,
			source_id,
			agenda_point_source_id,
			voting_type,
			decision_type,
			decision_text,
			comment,
			status,
			decision_order,
			source_updated_at,
			source_deleted,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now())
		ON CONFLICT (source_key, source_id)
		DO UPDATE SET motion_key = EXCLUDED.motion_key,
		              agenda_point_source_id = EXCLUDED.agenda_point_source_id,
		              voting_type = EXCLUDED.voting_type,
		              decision_type = EXCLUDED.decision_type,
		              decision_text = EXCLUDED.decision_text,
		              comment = EXCLUDED.comment,
		              status = EXCLUDED.status,
		              decision_order = EXCLUDED.decision_order,
		              source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              updated_at = now()
	`, projected.DecisionKey, projected.SourceKey, projected.MotionKey, projected.SourceID, projected.AgendaPointSourceID, projected.VotingType, projected.DecisionType, projected.DecisionText, projected.Comment, projected.Status, projected.DecisionOrder, projected.SourceUpdatedAt, projected.SourceDeleted)
	if err != nil {
		return false, err
	}

	return rawChanged, nil
}

func storeVote(ctx context.Context, tx pgx.Tx, motion motionCandidate, decision tweedekamer.DecisionRecord, vote tweedekamer.VoteRecord) (bool, error) {
	raw := projectVoteRaw(vote)
	projected := projectVote(motion, decision, vote)

	rawChanged, err := storeRawRecord(ctx, tx, raw)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO votes (
			vote_key,
			source_key,
			motion_key,
			decision_key,
			source_id,
			vote_type,
			party_source_id,
			party_name,
			actor_name,
			party_size,
			mistake,
			person_source_id,
			source_updated_at,
			source_deleted,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now())
		ON CONFLICT (source_key, source_id)
		DO UPDATE SET motion_key = EXCLUDED.motion_key,
		              decision_key = EXCLUDED.decision_key,
		              vote_type = EXCLUDED.vote_type,
		              party_source_id = EXCLUDED.party_source_id,
		              party_name = EXCLUDED.party_name,
		              actor_name = EXCLUDED.actor_name,
		              party_size = EXCLUDED.party_size,
		              mistake = EXCLUDED.mistake,
		              person_source_id = EXCLUDED.person_source_id,
		              source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              updated_at = now()
	`, projected.VoteKey, projected.SourceKey, projected.MotionKey, projected.DecisionKey, projected.SourceID, projected.VoteType, projected.PartySourceID, projected.PartyName, projected.ActorName, projected.PartySize, projected.Mistake, projected.PersonSourceID, projected.SourceUpdatedAt, projected.SourceDeleted)
	if err != nil {
		return false, err
	}

	return rawChanged, nil
}

func storeRawRecord(
	ctx context.Context,
	tx pgx.Tx,
	record rawRecordProjection,
) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO raw_records (
			source_key,
			collection,
			source_id,
			source_updated_at,
			source_deleted,
			payload,
			payload_hash
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_key, collection, source_id)
		DO UPDATE SET source_updated_at = EXCLUDED.source_updated_at,
		              source_deleted = EXCLUDED.source_deleted,
		              payload = EXCLUDED.payload,
		              payload_hash = EXCLUDED.payload_hash,
		              last_seen_at = now()
		WHERE raw_records.payload_hash IS DISTINCT FROM EXCLUDED.payload_hash
		   OR raw_records.source_updated_at IS DISTINCT FROM EXCLUDED.source_updated_at
		   OR raw_records.source_deleted IS DISTINCT FROM EXCLUDED.source_deleted
	`, tweedeKamerSourceKey, record.Collection, record.SourceID, record.SourceUpdatedAt, record.SourceDeleted, string(record.Payload), record.PayloadHash)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func acquirePipelineLock(ctx context.Context, pool *pgxpool.Pool, pipeline string) (func(), error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	lockKey := tweedeKamerSourceKey + ":" + pipeline
	var acquired bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", lockKey).Scan(&acquired); err != nil {
		conn.Release()
		return nil, err
	}
	if !acquired {
		conn.Release()
		return nil, fmt.Errorf("ingestion pipeline already running: %s/%s", tweedeKamerSourceKey, pipeline)
	}

	return func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock(hashtext($1))", lockKey)
		conn.Release()
	}, nil
}

func startPipelineRun(ctx context.Context, pool *pgxpool.Pool, pipeline string) (int64, error) {
	return startPipelineRunWithCursor(ctx, pool, pipeline, Cursor{})
}

func startPipelineRunWithCursor(ctx context.Context, pool *pgxpool.Pool, pipeline string, cursorBefore Cursor) (int64, error) {
	raw, err := json.Marshal(cursorBefore)
	if err != nil {
		return 0, err
	}

	var runID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO ingestion_runs (source_key, pipeline, status, cursor_before)
		VALUES ($1, $2, 'running', $3)
		RETURNING id
	`, tweedeKamerSourceKey, pipeline, string(raw)).Scan(&runID)
	return runID, err
}

func finishPipelineRun(
	ctx context.Context,
	pool *pgxpool.Pool,
	runID int64,
	pipeline string,
	status string,
	recordsSeen int,
	recordsChanged int,
	cursorSaved bool,
	stopReason string,
	errorMessage string,
) error {
	return finishPipelineRunWithCursor(ctx, pool, runID, pipeline, status, Cursor{}, recordsSeen, recordsChanged, cursorSaved, stopReason, errorMessage)
}

func finishPipelineRunWithCursor(
	ctx context.Context,
	pool *pgxpool.Pool,
	runID int64,
	pipeline string,
	status string,
	cursorAfter Cursor,
	recordsSeen int,
	recordsChanged int,
	cursorSaved bool,
	stopReason string,
	errorMessage string,
) error {
	raw, err := json.Marshal(cursorAfter)
	if err != nil {
		return err
	}
	var nullableError *string
	if errorMessage != "" {
		nullableError = &errorMessage
	}

	_, err = pool.Exec(ctx, `
		UPDATE ingestion_runs
		SET status = $2,
		    cursor_after = $3,
		    records_seen = $4,
		    records_changed = $5,
		    error_message = $6,
		    cursor_saved = $7,
		    stop_reason = $8,
		    finished_at = now()
		WHERE id = $1 AND pipeline = $9
	`, runID, status, string(raw), recordsSeen, recordsChanged, nullableError, cursorSaved, stopReason, pipeline)
	return err
}

func decisionKey(sourceID string) string {
	return tweedeKamerSourceKey + ":decision:" + sourceID
}

func voteKey(sourceID string) string {
	return tweedeKamerSourceKey + ":vote:" + sourceID
}
