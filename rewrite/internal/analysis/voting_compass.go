package analysis

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VotingCompassPosition struct {
	PartySourceID *string
	PartyName     string
	Position      string
}

type VotingCompassMotion struct {
	MotionKey  string
	Number     *string
	Title      *string
	Subject    *string
	ProposedAt *time.Time
	Positions  []VotingCompassPosition
}

type VotingCompassOptions struct {
	Jurisdiction string
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	MinParties   int
}

func LoadVotingCompassMotions(ctx context.Context, pool *pgxpool.Pool, options VotingCompassOptions) ([]VotingCompassMotion, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	limit := options.Limit
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	minParties := options.MinParties
	if minParties <= 0 {
		minParties = 8
	}

	rows, err := pool.Query(ctx, `
		WITH party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown') AS party_name,
			       CASE
			         WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) THEN 'FOR'
			         WHEN SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) THEN 'AGAINST'
			         ELSE 'NEUTRAL'
			       END AS position
			FROM votes v
			JOIN motions m ON m.motion_key = v.motion_key
			LEFT JOIN parties p ON p.source_key = v.source_key
			                   AND p.source_id = v.party_source_id
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND ($2::timestamptz IS NULL OR m.proposed_at >= $2)
			  AND ($3::timestamptz IS NULL OR m.proposed_at < $3)
			  AND v.source_deleted = false
			  AND v.mistake = false
			  AND v.vote_type IN ('Voor', 'Tegen')
			GROUP BY v.motion_key,
			         v.party_source_id,
			         COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown')
			HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
		),
		eligible_motions AS (
			SELECT m.motion_key,
			       m.number,
			       m.title,
			       m.subject,
			       m.proposed_at
			FROM motions m
			JOIN party_positions pp ON pp.motion_key = m.motion_key
			GROUP BY m.motion_key
			HAVING COUNT(*) >= $4
			ORDER BY m.proposed_at DESC NULLS LAST, m.motion_key
			LIMIT $5
		)
		SELECT em.motion_key,
		       em.number,
		       em.title,
		       em.subject,
		       em.proposed_at,
		       pp.party_source_id,
		       pp.party_name,
		       pp.position
		FROM eligible_motions em
		JOIN party_positions pp ON pp.motion_key = em.motion_key
		ORDER BY em.proposed_at DESC NULLS LAST, em.motion_key, pp.party_name
	`, jurisdiction, options.DateFrom, options.DateTo, minParties, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	motionIndex := map[string]int{}
	motions := []VotingCompassMotion{}
	for rows.Next() {
		var motion VotingCompassMotion
		var position VotingCompassPosition
		if err := rows.Scan(
			&motion.MotionKey,
			&motion.Number,
			&motion.Title,
			&motion.Subject,
			&motion.ProposedAt,
			&position.PartySourceID,
			&position.PartyName,
			&position.Position,
		); err != nil {
			return nil, err
		}

		index, ok := motionIndex[motion.MotionKey]
		if !ok {
			index = len(motions)
			motionIndex[motion.MotionKey] = index
			motions = append(motions, motion)
		}
		motions[index].Positions = append(motions[index].Positions, position)
	}
	return motions, rows.Err()
}
