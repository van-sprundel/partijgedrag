package inspect

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func PrintMotion(ctx context.Context, pool *pgxpool.Pool, writer io.Writer, motionKey string) error {
	var motion struct {
		MotionKey     string
		SourceID      string
		Title         *string
		Subject       *string
		Status        *string
		ProposedAt    *time.Time
		VotesSyncedAt *time.Time
	}

	err := pool.QueryRow(ctx, `
		SELECT motion_key,
		       source_id,
		       title,
		       subject,
		       status,
		       proposed_at,
		       votes_synced_at
		FROM motions
		WHERE motion_key = $1
	`, motionKey).Scan(
		&motion.MotionKey,
		&motion.SourceID,
		&motion.Title,
		&motion.Subject,
		&motion.Status,
		&motion.ProposedAt,
		&motion.VotesSyncedAt,
	)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("motion not found: %s", motionKey)
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "motion_key: %s\n", motion.MotionKey)
	fmt.Fprintf(writer, "source_id: %s\n", motion.SourceID)
	fmt.Fprintf(writer, "title: %s\n", value(motion.Title))
	fmt.Fprintf(writer, "subject: %s\n", value(motion.Subject))
	fmt.Fprintf(writer, "status: %s\n", value(motion.Status))
	fmt.Fprintf(writer, "proposed_at: %s\n", timeValue(motion.ProposedAt))
	fmt.Fprintf(writer, "votes_synced_at: %s\n", timeValue(motion.VotesSyncedAt))

	if err := printDecisions(ctx, pool, writer, motionKey); err != nil {
		return err
	}

	return printPartyPositions(ctx, pool, writer, motionKey)
}

func printDecisions(ctx context.Context, pool *pgxpool.Pool, writer io.Writer, motionKey string) error {
	rows, err := pool.Query(ctx, `
		SELECT d.decision_key,
		       d.source_id,
		       d.decision_type,
		       d.decision_text,
		       d.status,
		       d.decision_order,
		       count(v.vote_key) FILTER (WHERE v.source_deleted = false)::int AS vote_count,
		       count(v.vote_key) FILTER (WHERE v.source_deleted = false AND v.mistake = true)::int AS mistake_count
		FROM decisions d
		LEFT JOIN votes v ON v.decision_key = d.decision_key
		WHERE d.motion_key = $1
		  AND d.source_deleted = false
		GROUP BY d.decision_key
		ORDER BY d.decision_order NULLS LAST, d.source_updated_at NULLS LAST
	`, motionKey)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Fprintln(writer, "\ndecisions:")
	count := 0
	for rows.Next() {
		count++
		var decisionKey, sourceID string
		var decisionType, decisionText, status *string
		var decisionOrder *int
		var voteCount, mistakeCount int
		if err := rows.Scan(&decisionKey, &sourceID, &decisionType, &decisionText, &status, &decisionOrder, &voteCount, &mistakeCount); err != nil {
			return err
		}

		fmt.Fprintf(writer, "- %s source_id=%s type=%s status=%s order=%s votes=%d mistakes=%d text=%s\n",
			decisionKey,
			sourceID,
			value(decisionType),
			value(status),
			intValue(decisionOrder),
			voteCount,
			mistakeCount,
			value(decisionText),
		)
	}
	if count == 0 {
		fmt.Fprintln(writer, "- none")
	}
	return rows.Err()
}

func printPartyPositions(ctx context.Context, pool *pgxpool.Pool, writer io.Writer, motionKey string) error {
	rows, err := pool.Query(ctx, `
		SELECT COALESCE(party_name, actor_name, party_source_id, 'unknown') AS party_name,
		       party_source_id,
		       SUM(CASE WHEN vote_type = 'Voor' THEN 1 ELSE 0 END)::int AS votes_for,
		       SUM(CASE WHEN vote_type = 'Tegen' THEN 1 ELSE 0 END)::int AS votes_against,
		       COUNT(*)::int AS total_votes
		FROM votes
		WHERE motion_key = $1
		  AND source_deleted = false
		  AND mistake = false
		  AND vote_type IN ('Voor', 'Tegen')
		GROUP BY COALESCE(party_name, actor_name, party_source_id, 'unknown'), party_source_id
		ORDER BY party_name
	`, motionKey)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Fprintln(writer, "\nparty_positions:")
	count := 0
	for rows.Next() {
		count++
		var partyName string
		var partySourceID *string
		var votesFor, votesAgainst, totalVotes int
		if err := rows.Scan(&partyName, &partySourceID, &votesFor, &votesAgainst, &totalVotes); err != nil {
			return err
		}

		position := "NEUTRAL"
		if votesFor > votesAgainst {
			position = "FOR"
		}
		if votesAgainst > votesFor {
			position = "AGAINST"
		}

		fmt.Fprintf(writer, "- %s party_id=%s position=%s voor=%d tegen=%d total=%d\n",
			partyName,
			value(partySourceID),
			position,
			votesFor,
			votesAgainst,
			totalVotes,
		)
	}
	if count == 0 {
		fmt.Fprintln(writer, "- none")
	}
	return rows.Err()
}

func value(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func timeValue(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.RFC3339)
}

func intValue(value *int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}
