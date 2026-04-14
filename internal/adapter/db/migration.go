package db

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

func ApplyMigrations(
	ctx context.Context,
	database *sqlx.DB,
	migrationFS fs.FS,
) error {
	if _, err := database.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY
		)`,
	); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.Glob(migrationFS, "migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)

	for _, entry := range entries {
		version := strings.TrimSuffix(strings.TrimPrefix(entry, "migrations/"), ".up.sql")
		applied, err := migrationApplied(ctx, database, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := fs.ReadFile(migrationFS, entry)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		if _, err := database.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("execute migration %s: %w", version, err)
		}

		if _, err := database.ExecContext(
			ctx,
			`INSERT INTO schema_migrations(version) VALUES (?)`,
			version,
		); err != nil {
			return fmt.Errorf("track migration %s: %w", version, err)
		}
	}

	return nil
}

func migrationApplied(
	ctx context.Context,
	database *sqlx.DB,
	version string,
) (bool, error) {
	var count int
	if err := database.GetContext(
		ctx,
		&count,
		`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`,
		version,
	); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return count > 0, nil
}
