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
	MotionKey    string
	Number       *string
	Title        *string
	Subject      *string
	ProposedAt   *time.Time
	BulletPoints []string
	DocumentURL  *string
	Positions    []VotingCompassPosition
}

type VotingCompassOptions struct {
	Jurisdiction string
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	MinParties   int
	ExcludeKeys  []string
	// CategoryKeys keeps only motions tagged with at least one of these categories.
	CategoryKeys []string
	// PartySourceIDs keeps only motions where the selected parties (two or more)
	// did not all vote the same way.
	PartySourceIDs []string
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
	excludeKeys := options.ExcludeKeys
	if excludeKeys == nil {
		excludeKeys = []string{}
	}
	categoryKeys := options.CategoryKeys
	if categoryKeys == nil {
		categoryKeys = []string{}
	}
	partySourceIDs := options.PartySourceIDs
	if partySourceIDs == nil {
		partySourceIDs = []string{}
	}

	rows, err := pool.Query(ctx, `
		WITH candidates AS (
			SELECT m.motion_key,
			       m.number,
			       m.title,
			       m.subject,
			       m.proposed_at,
			       m.bullet_points,
			       m.document_url
			FROM motions m
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND m.votes_synced_at IS NOT NULL
			  AND m.bullet_points IS NOT NULL
			  AND jsonb_array_length(m.bullet_points) > 0
			  AND EXISTS (
			      SELECT 1 FROM jsonb_array_elements_text(m.bullet_points) AS bullet
			      WHERE bullet ILIKE 'verzoekt%'
			  )
			  AND ($2::timestamptz IS NULL OR m.proposed_at >= $2)
			  AND ($3::timestamptz IS NULL OR m.proposed_at < $3)
			  AND (cardinality($6::text[]) = 0 OR m.motion_key <> ALL($6))
			  AND (cardinality($7::text[]) = 0 OR EXISTS (
			      SELECT 1 FROM motion_categories mc
			      WHERE mc.motion_key = m.motion_key
			        AND mc.category_key = ANY($7)
			  ))
			  AND (cardinality($8::text[]) < 2 OR (
			      SELECT COUNT(DISTINCT v.vote_type)
			      FROM votes v
			      WHERE v.motion_key = m.motion_key
			        AND v.party_source_id = ANY($8)
			        AND v.vote_type IN ('Voor', 'Tegen')
			        AND v.source_deleted = false
			        AND v.mistake = false
			  ) > 1)
			ORDER BY m.proposed_at DESC NULLS LAST, m.motion_key
			LIMIT $5 * 5
		),
		party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown') AS party_name,
			       CASE
			         WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) THEN 'FOR'
			         WHEN SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) THEN 'AGAINST'
			         ELSE 'NEUTRAL'
			       END AS position
			FROM votes v
			JOIN candidates c ON c.motion_key = v.motion_key
			LEFT JOIN parties p ON p.source_key = v.source_key
			                   AND p.source_id = v.party_source_id
			WHERE v.source_deleted = false
			  AND v.mistake = false
			  AND v.vote_type IN ('Voor', 'Tegen')
			GROUP BY v.motion_key,
			         v.party_source_id,
			         COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown')
			HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
		),
		eligible_motions AS (
			SELECT c.motion_key,
			       c.number,
			       c.title,
			       c.subject,
			       c.proposed_at,
			       c.bullet_points,
			       c.document_url
			FROM candidates c
			JOIN party_positions pp ON pp.motion_key = c.motion_key
			GROUP BY c.motion_key, c.number, c.title, c.subject, c.proposed_at, c.bullet_points, c.document_url
			HAVING COUNT(*) >= $4
			ORDER BY c.proposed_at DESC NULLS LAST, c.motion_key
			LIMIT $5
		)
		SELECT em.motion_key,
		       em.number,
		       em.title,
		       em.subject,
		       em.proposed_at,
		       em.bullet_points,
		       em.document_url,
		       pp.party_source_id,
		       pp.party_name,
		       pp.position
		FROM eligible_motions em
		JOIN party_positions pp ON pp.motion_key = em.motion_key
		ORDER BY em.proposed_at DESC NULLS LAST, em.motion_key, pp.party_name
	`, jurisdiction, options.DateFrom, options.DateTo, minParties, limit, excludeKeys, categoryKeys, partySourceIDs)
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
			&motion.BulletPoints,
			&motion.DocumentURL,
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
