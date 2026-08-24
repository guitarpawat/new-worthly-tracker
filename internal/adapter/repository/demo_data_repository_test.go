package repository

import (
	"context"
	"testing"

	dbfiles "github.com/guitarpawat/worthly-tracker/db"
)

func TestDemoDataRepository_SeedDevDataInsertsSeedRows(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	repo := NewDemoDataRepository(database, dbfiles.FS)

	hasData, err := repo.HasAnyUserData(context.Background())
	if err != nil {
		t.Fatalf("HasAnyUserData returned error: %v", err)
	}
	if hasData {
		t.Fatal("expected empty test database before seeding")
	}

	if err := repo.SeedDevData(context.Background()); err != nil {
		t.Fatalf("SeedDevData returned error: %v", err)
	}

	hasData, err = repo.HasAnyUserData(context.Background())
	if err != nil {
		t.Fatalf("HasAnyUserData returned error after seed: %v", err)
	}
	if !hasData {
		t.Fatal("expected seeded database to contain user data")
	}

	countTests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "asset types", query: `SELECT COUNT(*) FROM asset_types`, want: 8},
		{name: "assets", query: `SELECT COUNT(*) FROM assets`, want: 16},
		{name: "snapshots", query: `SELECT COUNT(*) FROM record_snapshots`, want: 28},
		{name: "record items", query: `SELECT COUNT(*) FROM record_items`, want: 328},
		{name: "goals", query: `SELECT COUNT(*) FROM goals`, want: 3},
		{
			name: "latest snapshot items",
			query: `SELECT COUNT(*)
				FROM record_items ri
				INNER JOIN record_snapshots rs ON rs.id = ri.snapshot_id
				WHERE rs.record_date = '2026-08-12'`,
			want: 13,
		},
		{
			name: "same named latest assets",
			query: `SELECT COUNT(*)
				FROM record_items ri
				INNER JOIN record_snapshots rs ON rs.id = ri.snapshot_id
				INNER JOIN assets a ON a.id = ri.asset_id
				WHERE rs.record_date = '2026-08-12'
				  AND a.name = 'Reserve'`,
			want: 2,
		},
	}

	for _, testCase := range countTests {
		t.Run(testCase.name, func(t *testing.T) {
			var count int
			if err := database.GetContext(context.Background(), &count, testCase.query); err != nil {
				t.Fatalf("count seeded %s: %v", testCase.name, err)
			}
			if count != testCase.want {
				t.Fatalf("expected %d seeded %s, got %d", testCase.want, testCase.name, count)
			}
		})
	}
}
