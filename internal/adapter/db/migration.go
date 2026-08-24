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

		if err := applyMigration(ctx, database, version, string(sqlBytes)); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(
	ctx context.Context,
	database *sqlx.DB,
	version string,
	rawSQL string,
) (returnErr error) {
	migrationSQL, disableForeignKeys, err := prepareMigrationSQL(rawSQL)
	if err != nil {
		return fmt.Errorf("prepare migration %s: %w", version, err)
	}

	connection, err := database.Connx(ctx)
	if err != nil {
		return fmt.Errorf("open migration connection %s: %w", version, err)
	}
	defer func() {
		_ = connection.Close()
	}()

	if disableForeignKeys {
		if _, err := connection.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
			return fmt.Errorf("disable foreign keys for migration %s: %w", version, err)
		}
		defer func() {
			if _, err := connection.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`); err != nil && returnErr == nil {
				returnErr = fmt.Errorf("restore foreign keys after migration %s: %w", version, err)
			}
		}()
	}

	tx, err := connection.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, migrationSQL); err != nil {
		return fmt.Errorf("execute migration %s: %w", version, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations(version) VALUES (?)`,
		version,
	); err != nil {
		return fmt.Errorf("track migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	return nil
}

func prepareMigrationSQL(rawSQL string) (string, bool, error) {
	const (
		beginStatement              = "BEGIN;"
		commitStatement             = "COMMIT;"
		disableForeignKeysStatement = "PRAGMA foreign_keys = OFF;"
		enableForeignKeysStatement  = "PRAGMA foreign_keys = ON;"
	)

	migrationSQL := strings.TrimSpace(rawSQL)
	disableForeignKeys := strings.HasPrefix(migrationSQL, disableForeignKeysStatement)
	if disableForeignKeys {
		migrationSQL = strings.TrimSpace(strings.TrimPrefix(migrationSQL, disableForeignKeysStatement))
		if !strings.HasSuffix(migrationSQL, enableForeignKeysStatement) {
			return "", false, fmt.Errorf("foreign key migration must restore foreign keys")
		}
		migrationSQL = strings.TrimSpace(strings.TrimSuffix(migrationSQL, enableForeignKeysStatement))
	}

	if !strings.HasPrefix(migrationSQL, beginStatement) || !strings.HasSuffix(migrationSQL, commitStatement) {
		return "", false, fmt.Errorf("migration must be wrapped in BEGIN and COMMIT")
	}
	migrationSQL = strings.TrimSpace(strings.TrimPrefix(migrationSQL, beginStatement))
	migrationSQL = strings.TrimSpace(strings.TrimSuffix(migrationSQL, commitStatement))
	if migrationSQL == "" {
		return "", false, fmt.Errorf("migration body must not be empty")
	}

	return migrationSQL, disableForeignKeys, nil
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
