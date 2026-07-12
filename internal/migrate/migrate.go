package migrate

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext('partijgedrag:rewrite:migrations'))"); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`); err != nil {
		return err
	}

	applied, err := appliedVersions(ctx, tx)
	if err != nil {
		return err
	}

	migrations, err := migrations()
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if applied[migration.version] {
			continue
		}

		fmt.Printf("Applying %s\n", migration.version)
		if _, err := tx.Exec(ctx, migration.sql); err != nil {
			return fmt.Errorf("apply %s: %w", migration.version, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.version); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func appliedVersions(ctx context.Context, tx pgx.Tx) (map[string]bool, error) {
	rows, err := tx.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

type migration struct {
	version string
	sql     string
}

func migrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	result := make([]migration, 0, len(files))
	for _, file := range files {
		path := filepath.Join("migrations", file)
		content, err := migrationFiles.ReadFile(path)
		if err != nil {
			return nil, err
		}

		result = append(result, migration{
			version: strings.TrimSuffix(file, ".sql"),
			sql:     string(content),
		})
	}

	return result, nil
}
