package db

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	dbfiles "github.com/guitarpawat/worthly-tracker/db"
)

func TestApplyMigrations_ConvertsBooleanAutoIncrementSchemaToNumericAmount(t *testing.T) {
	t.Parallel()

	database, err := Open(SQLiteConfig{Path: ":memory:"})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	ctx := context.Background()
	if err := ApplyMigrations(ctx, database, mustBuildMigrationFS(
		t,
		"0001_init.up.sql",
		"0002_add_goals.up.sql",
		"0003_add_asset_increment_amount.up.sql",
	)); err != nil {
		t.Fatalf("apply pre-0004 migrations: %v", err)
	}

	statements := []string{
		`INSERT INTO asset_types (id, name, ordering) VALUES (1, 'Investment', 1)`,
		`INSERT INTO assets (id, asset_type_id, name, broker, is_cash, is_active, auto_increment, increment_amount, ordering) VALUES
			(1, 1, 'SET50 ETF', 'KKP', FALSE, TRUE, TRUE, 6500, 1),
			(2, 1, 'Global Bond', 'SCB', FALSE, TRUE, FALSE, 0, 2)`,
		`INSERT INTO record_snapshots (id, record_date) VALUES (1, '2026-04-12')`,
		`INSERT INTO record_items (id, snapshot_id, asset_id, bought_price, current_price, remarks) VALUES
			(1, 1, 1, 120000, 126000, 'Core'),
			(2, 1, 2, 50000, 51000, 'Bond')`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("seed pre-0004 schema: %v", err)
		}
	}

	if err := ApplyMigrations(ctx, database, mustBuildMigrationFS(
		t,
		"0004_change_auto_increment_to_amount.up.sql",
	)); err != nil {
		t.Fatalf("apply 0004 migration: %v", err)
	}

	type assetRow struct {
		ID            int64   `db:"id"`
		AutoIncrement float64 `db:"auto_increment"`
	}

	rows := []assetRow{}
	if err := database.SelectContext(ctx, &rows, `
		SELECT id, auto_increment
		FROM assets
		ORDER BY id
	`); err != nil {
		t.Fatalf("select migrated assets: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 migrated assets, got %d", len(rows))
	}
	if rows[0].AutoIncrement != 6500 {
		t.Fatalf("expected first asset auto_increment 6500, got %f", rows[0].AutoIncrement)
	}
	if rows[1].AutoIncrement != 0 {
		t.Fatalf("expected second asset auto_increment 0, got %f", rows[1].AutoIncrement)
	}

	var joinedRows int
	if err := database.GetContext(ctx, &joinedRows, `
		SELECT COUNT(*)
		FROM record_items ri
		INNER JOIN assets a ON a.id = ri.asset_id
		WHERE ri.deleted_at IS NULL
	`); err != nil {
		t.Fatalf("count joined record items after migration: %v", err)
	}
	if joinedRows != 2 {
		t.Fatalf("expected 2 joined record items after migration, got %d", joinedRows)
	}
}

func mustBuildMigrationFS(t *testing.T, names ...string) fs.FS {
	t.Helper()

	files := fstest.MapFS{}
	for _, name := range names {
		content, err := fs.ReadFile(dbfiles.FS, "migrations/"+name)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		files["migrations/"+name] = &fstest.MapFile{Data: content}
	}

	return files
}
