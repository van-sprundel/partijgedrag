package analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"partijgedrag/internal/cache"
)

// PartyComparison answers "where do these two parties actually differ?" — the
// question a visitor is left with after reading a single cell of the likeness
// matrix. It deliberately classifies positions the same way LoadPartyLikeness
// does, so the headline percentage here equals the cell that was clicked.
type PartyComparisonOptions struct {
	Jurisdiction   string
	Party1SourceID string
	Party2SourceID string
	DateFrom       *time.Time
	DateTo         *time.Time
}

type PartyComparison struct {
	CommonMotions  int
	SameVotes      int
	DifferentVotes int
	Similarity     float64
	Categories     []ComparisonCategory
}

type ComparisonCategory struct {
	CategoryKey    string
	Name           string
	Kind           string
	CommonMotions  int
	SameVotes      int
	DifferentVotes int
	Agreement      float64
}

type ComparisonMotionOptions struct {
	PartyComparisonOptions
	// Relation selects which side of the comparison to list: "disagree"
	// (default), "agree", or "all".
	Relation string
	Category string
	Limit    int
	Offset   int
}

type ComparisonMotion struct {
	MotionKey      string
	Number         *string
	Title          *string
	Subject        *string
	ProposedAt     *time.Time
	Party1Position string
	Party2Position string
	VotesFor       int
	VotesAgainst   int
	Categories     []ComparisonMotionCategory
}

type ComparisonMotionCategory struct {
	CategoryKey string
	Name        string
	Kind        string
}

func NormalizeComparisonRelation(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "disagree":
		return "disagree", true
	case "agree":
		return "agree", true
	case "all":
		return "all", true
	default:
		return "", false
	}
}

// LoadPartyComparison returns the headline totals plus the per-category
// breakdown for a pair of parties.
func LoadPartyComparison(ctx context.Context, pool *pgxpool.Pool, options PartyComparisonOptions) (PartyComparison, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}

	cacheKey := fmt.Sprintf("analysis:party_comparison:%s:%s:%s:%s:%s",
		jurisdiction, options.Party1SourceID, options.Party2SourceID,
		formatOptTime(options.DateFrom), formatOptTime(options.DateTo))
	if cached, ok := cache.Global().Get(cacheKey); ok {
		comparison := cached.(PartyComparison)
		comparison.Categories = copyComparisonCategories(comparison.Categories)
		return comparison, nil
	}

	comparison := PartyComparison{}
	err := pool.QueryRow(ctx, comparisonPairsSQL()+`
		SELECT COUNT(*)::int AS common_motions,
		       COUNT(*) FILTER (WHERE agree)::int AS same_votes,
		       COUNT(*) FILTER (WHERE NOT agree)::int AS different_votes
		FROM pairs
	`, jurisdiction, options.Party1SourceID, options.Party2SourceID, options.DateFrom, options.DateTo).Scan(
		&comparison.CommonMotions,
		&comparison.SameVotes,
		&comparison.DifferentVotes,
	)
	if err != nil {
		return PartyComparison{}, err
	}
	if comparison.CommonMotions > 0 {
		comparison.Similarity = (float64(comparison.SameVotes) / float64(comparison.CommonMotions)) * 100
	}

	categories, err := loadComparisonCategories(ctx, pool, jurisdiction, options)
	if err != nil {
		return PartyComparison{}, err
	}
	comparison.Categories = categories

	cache.Global().Set(cacheKey, comparison)
	return comparison, nil
}

// loadComparisonCategories orders by the number of disagreements rather than by
// agreement percentage: a category the parties split 0/3 on would otherwise top
// the list ahead of one they split 61/100 on, which is the opposite of useful.
func loadComparisonCategories(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, options PartyComparisonOptions) ([]ComparisonCategory, error) {
	rows, err := pool.Query(ctx, comparisonPairsSQL()+`
		SELECT c.category_key,
		       c.name,
		       c.kind,
		       COUNT(*)::int AS common_motions,
		       COUNT(*) FILTER (WHERE p.agree)::int AS same_votes,
		       COUNT(*) FILTER (WHERE NOT p.agree)::int AS different_votes
		FROM pairs p
		JOIN motion_categories mc ON mc.motion_key = p.motion_key
		JOIN categories c ON c.category_key = mc.category_key
		GROUP BY c.category_key, c.name, c.kind
		ORDER BY different_votes DESC, common_motions DESC, c.name
	`, jurisdiction, options.Party1SourceID, options.Party2SourceID, options.DateFrom, options.DateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []ComparisonCategory{}
	for rows.Next() {
		var category ComparisonCategory
		if err := rows.Scan(
			&category.CategoryKey,
			&category.Name,
			&category.Kind,
			&category.CommonMotions,
			&category.SameVotes,
			&category.DifferentVotes,
		); err != nil {
			return nil, err
		}
		if category.CommonMotions > 0 {
			category.Agreement = (float64(category.SameVotes) / float64(category.CommonMotions)) * 100
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

// LoadComparisonMotions lists the motions behind the comparison, filtered to the
// requested relation, and returns the total for that filter so the page can
// paginate.
func LoadComparisonMotions(ctx context.Context, pool *pgxpool.Pool, options ComparisonMotionOptions) ([]ComparisonMotion, int, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	relation, ok := NormalizeComparisonRelation(options.Relation)
	if !ok {
		return nil, 0, fmt.Errorf("invalid relation %q", options.Relation)
	}
	limit := options.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}
	offset := options.Offset
	if offset < 0 {
		offset = 0
	}

	cacheKey := fmt.Sprintf("analysis:comparison_motions:%s:%s:%s:%s:%s:%s:%s:%d:%d",
		jurisdiction, options.Party1SourceID, options.Party2SourceID,
		formatOptTime(options.DateFrom), formatOptTime(options.DateTo),
		relation, options.Category, limit, offset)
	if cached, ok := cache.Global().Get(cacheKey); ok {
		page := cached.(comparisonMotionPage)
		return copyComparisonMotions(page.Motions), page.Total, nil
	}

	rows, err := pool.Query(ctx, comparisonPairsSQL()+`
		SELECT m.motion_key,
		       m.number,
		       m.title,
		       m.subject,
		       m.proposed_at,
		       p.party1_position,
		       p.party2_position,
		       (SELECT COALESCE(SUM(CASE WHEN v.person_source_id IS NULL THEN COALESCE(v.party_size, 1) ELSE 1 END), 0)::int
		          FROM votes v
		         WHERE v.motion_key = m.motion_key AND v.source_deleted = false AND v.mistake = false AND v.vote_type = 'Voor') AS votes_for,
		       (SELECT COALESCE(SUM(CASE WHEN v.person_source_id IS NULL THEN COALESCE(v.party_size, 1) ELSE 1 END), 0)::int
		          FROM votes v
		         WHERE v.motion_key = m.motion_key AND v.source_deleted = false AND v.mistake = false AND v.vote_type = 'Tegen') AS votes_against,
		       COUNT(*) OVER ()::int AS total
		FROM pairs p
		JOIN motions m ON m.motion_key = p.motion_key
		WHERE (
		    $6::text = 'all'
		    OR ($6::text = 'agree' AND p.agree)
		    OR ($6::text = 'disagree' AND NOT p.agree)
		  )
		  AND (
		    $7::text = ''
		    OR EXISTS (
		      SELECT 1
		      FROM motion_categories mc
		      WHERE mc.motion_key = p.motion_key
		        AND mc.category_key = $7
		    )
		  )
		ORDER BY m.proposed_at DESC NULLS LAST, m.motion_key
		LIMIT $8 OFFSET $9
	`, jurisdiction, options.Party1SourceID, options.Party2SourceID, options.DateFrom, options.DateTo,
		relation, options.Category, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	motions := []ComparisonMotion{}
	total := 0
	for rows.Next() {
		var motion ComparisonMotion
		if err := rows.Scan(
			&motion.MotionKey,
			&motion.Number,
			&motion.Title,
			&motion.Subject,
			&motion.ProposedAt,
			&motion.Party1Position,
			&motion.Party2Position,
			&motion.VotesFor,
			&motion.VotesAgainst,
			&total,
		); err != nil {
			return nil, 0, err
		}
		motions = append(motions, motion)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if err := attachComparisonCategories(ctx, pool, motions); err != nil {
		return nil, 0, err
	}

	cache.Global().Set(cacheKey, comparisonMotionPage{Motions: copyComparisonMotions(motions), Total: total})
	return motions, total, nil
}

// attachComparisonCategories labels the listed motions in one round trip rather
// than one query per row.
func attachComparisonCategories(ctx context.Context, pool *pgxpool.Pool, motions []ComparisonMotion) error {
	if len(motions) == 0 {
		return nil
	}

	keys := make([]string, 0, len(motions))
	for _, motion := range motions {
		keys = append(keys, motion.MotionKey)
	}

	rows, err := pool.Query(ctx, `
		SELECT mc.motion_key, c.category_key, c.name, c.kind
		FROM motion_categories mc
		JOIN categories c ON c.category_key = mc.category_key
		WHERE mc.motion_key = ANY($1::text[])
		ORDER BY c.kind, c.name
	`, keys)
	if err != nil {
		return err
	}
	defer rows.Close()

	byMotion := map[string][]ComparisonMotionCategory{}
	for rows.Next() {
		var motionKey string
		var category ComparisonMotionCategory
		if err := rows.Scan(&motionKey, &category.CategoryKey, &category.Name, &category.Kind); err != nil {
			return err
		}
		byMotion[motionKey] = append(byMotion[motionKey], category)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for index := range motions {
		motions[index].Categories = byMotion[motions[index].MotionKey]
	}
	return nil
}

// comparisonPairsSQL classifies both parties' positions per motion and pairs
// them up. It mirrors LoadPartyLikeness: only motions where a party cast a
// clear (non-tied) Voor/Tegen majority count, so the totals derived here match
// the likeness matrix cell that links to this page.
func comparisonPairsSQL() string {
	return `
		WITH party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       CASE
			         WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) THEN 'FOR'
			         ELSE 'AGAINST'
			       END AS position
			FROM votes v
			JOIN motions m ON m.motion_key = v.motion_key
			WHERE m.jurisdiction_key = $1
			  AND m.source_deleted = false
			  AND v.source_deleted = false
			  AND v.mistake = false
			  AND v.party_source_id IN ($2, $3)
			  AND v.vote_type IN ('Voor', 'Tegen')
			  AND ($4::timestamptz IS NULL OR m.proposed_at >= $4)
			  AND ($5::timestamptz IS NULL OR m.proposed_at <= $5)
			GROUP BY v.motion_key, v.party_source_id
			HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
		),
		pairs AS (
			SELECT p1.motion_key,
			       p1.position AS party1_position,
			       p2.position AS party2_position,
			       (p1.position = p2.position) AS agree
			FROM party_positions p1
			JOIN party_positions p2 ON p2.motion_key = p1.motion_key
			                       AND p2.party_source_id = $3
			WHERE p1.party_source_id = $2
		)
	`
}

type comparisonMotionPage struct {
	Motions []ComparisonMotion
	Total   int
}

func copyComparisonCategories(src []ComparisonCategory) []ComparisonCategory {
	if src == nil {
		return nil
	}
	dst := make([]ComparisonCategory, len(src))
	copy(dst, src)
	return dst
}

func copyComparisonMotions(src []ComparisonMotion) []ComparisonMotion {
	if src == nil {
		return nil
	}
	dst := make([]ComparisonMotion, len(src))
	copy(dst, src)
	return dst
}
