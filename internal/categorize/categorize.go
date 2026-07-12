package categorize

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	CategoryKey string
	Name        string
	Kind        string
	Keywords    []string
}

type Matcher struct {
	categories []compiledCategory
}

type compiledCategory struct {
	categoryKey string
	patterns    []*regexp.Regexp
}

// NewMatcher compiles keyword patterns that match whole words only, so a
// keyword like "eu" does not match inside "euro". Boundaries are based on
// Unicode letters/digits because Go's \b is ASCII-only and mishandles Dutch
// diacritics such as "patiënt".
func NewMatcher(categories []Category) (*Matcher, error) {
	compiled := make([]compiledCategory, 0, len(categories))
	for _, category := range categories {
		patterns := make([]*regexp.Regexp, 0, len(category.Keywords))
		for _, keyword := range category.Keywords {
			keyword = strings.ToLower(strings.TrimSpace(keyword))
			if keyword == "" {
				continue
			}
			pattern, err := regexp.Compile(`(^|[^\p{L}\p{N}])` + regexp.QuoteMeta(keyword) + `([^\p{L}\p{N}]|$)`)
			if err != nil {
				return nil, fmt.Errorf("compile keyword %q for category %s: %w", keyword, category.CategoryKey, err)
			}
			patterns = append(patterns, pattern)
		}
		compiled = append(compiled, compiledCategory{
			categoryKey: category.CategoryKey,
			patterns:    patterns,
		})
	}
	return &Matcher{categories: compiled}, nil
}

// Match returns the category keys whose keywords appear in the motion's
// title or subject. A motion can match multiple categories.
func (matcher *Matcher) Match(title *string, subject *string) []string {
	searchText := strings.ToLower(strings.TrimSpace(stringValue(title) + " " + stringValue(subject)))
	if searchText == "" {
		return nil
	}

	matches := []string{}
	for _, category := range matcher.categories {
		for _, pattern := range category.patterns {
			if pattern.MatchString(searchText) {
				matches = append(matches, category.categoryKey)
				break
			}
		}
	}
	return matches
}

func LoadCategories(ctx context.Context, pool *pgxpool.Pool, jurisdiction string) ([]Category, error) {
	rows, err := pool.Query(ctx, `
		SELECT category_key, name, kind, keywords
		FROM categories
		WHERE jurisdiction_key = $1
		ORDER BY kind, name
	`, jurisdiction)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.CategoryKey, &category.Name, &category.Kind, &category.Keywords); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

type Options struct {
	Jurisdiction string
	BatchSize    int
	MaxMotions   int
	Recategorize bool
}

type Stats struct {
	MotionsSeen    int
	MotionsMatched int
	Assignments    int
}

// Run assigns categories to motions that have not been categorized yet.
// Motions are marked with categorized_at even when nothing matches, so each
// motion is evaluated once; use Recategorize after changing keywords.
func Run(ctx context.Context, pool *pgxpool.Pool, options Options) (Stats, error) {
	jurisdiction := options.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 500
	}

	if options.Recategorize {
		if err := clearAssignments(ctx, pool, jurisdiction); err != nil {
			return Stats{}, err
		}
	}

	categories, err := LoadCategories(ctx, pool, jurisdiction)
	if err != nil {
		return Stats{}, err
	}
	if len(categories) == 0 {
		return Stats{}, fmt.Errorf("no categories found for jurisdiction %s, run migrate first", jurisdiction)
	}
	matcher, err := NewMatcher(categories)
	if err != nil {
		return Stats{}, err
	}

	stats := Stats{}
	for page := 1; ; page++ {
		limit := batchSize
		if options.MaxMotions > 0 && options.MaxMotions-stats.MotionsSeen < limit {
			limit = options.MaxMotions - stats.MotionsSeen
		}
		if limit <= 0 {
			break
		}

		motions, err := loadUncategorizedMotions(ctx, pool, jurisdiction, limit)
		if err != nil {
			return stats, err
		}
		if len(motions) == 0 {
			break
		}

		matched, assignments, err := storeAssignments(ctx, pool, matcher, motions)
		if err != nil {
			return stats, err
		}

		stats.MotionsSeen += len(motions)
		stats.MotionsMatched += matched
		stats.Assignments += assignments
		fmt.Printf("categorize page=%d seen=%d matched=%d assignments=%d\n", page, stats.MotionsSeen, stats.MotionsMatched, stats.Assignments)
	}

	return stats, nil
}

type motionText struct {
	MotionKey string
	Title     *string
	Subject   *string
}

func loadUncategorizedMotions(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, limit int) ([]motionText, error) {
	rows, err := pool.Query(ctx, `
		SELECT motion_key, title, subject
		FROM motions
		WHERE jurisdiction_key = $1
		  AND source_deleted = false
		  AND categorized_at IS NULL
		ORDER BY motion_key
		LIMIT $2
	`, jurisdiction, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	motions := []motionText{}
	for rows.Next() {
		var motion motionText
		if err := rows.Scan(&motion.MotionKey, &motion.Title, &motion.Subject); err != nil {
			return nil, err
		}
		motions = append(motions, motion)
	}
	return motions, rows.Err()
}

func storeAssignments(ctx context.Context, pool *pgxpool.Pool, matcher *Matcher, motions []motionText) (int, int, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	matched := 0
	assignments := 0
	batch := &pgx.Batch{}
	for _, motion := range motions {
		categoryKeys := matcher.Match(motion.Title, motion.Subject)
		if len(categoryKeys) > 0 {
			matched++
		}
		for _, categoryKey := range categoryKeys {
			assignments++
			batch.Queue(`
				INSERT INTO motion_categories (motion_key, category_key)
				VALUES ($1, $2)
				ON CONFLICT (motion_key, category_key) DO NOTHING
			`, motion.MotionKey, categoryKey)
		}
		batch.Queue(`
			UPDATE motions
			SET categorized_at = now()
			WHERE motion_key = $1
		`, motion.MotionKey)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		return 0, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return matched, assignments, nil
}

func clearAssignments(ctx context.Context, pool *pgxpool.Pool, jurisdiction string) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM motion_categories
		USING motions
		WHERE motion_categories.motion_key = motions.motion_key
		  AND motions.jurisdiction_key = $1
	`, jurisdiction)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE motions
		SET categorized_at = NULL
		WHERE jurisdiction_key = $1
	`, jurisdiction)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
