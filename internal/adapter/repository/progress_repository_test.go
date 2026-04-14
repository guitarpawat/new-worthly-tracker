package repository

import (
	"context"
	"testing"
)

func TestProgressRepository_ListSnapshotDatesReturnsAscendingDates(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	repo := NewProgressRepository(database)

	dates, err := repo.ListSnapshotDates(context.Background())
	if err != nil {
		t.Fatalf("ListSnapshotDates returned error: %v", err)
	}

	if len(dates) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(dates))
	}
	if dates[0] != "2026-03-12" || dates[1] != "2026-04-12" {
		t.Fatalf("unexpected dates: %+v", dates)
	}
}

func TestProgressRepository_ListSnapshotItemsInRangeReturnsOrderedRows(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	repo := NewProgressRepository(database)

	rows, err := repo.ListSnapshotItemsInRange(context.Background(), "2026-03-01", "2026-04-30")
	if err != nil {
		t.Fatalf("ListSnapshotItemsInRange returned error: %v", err)
	}

	if len(rows) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(rows))
	}
	if rows[0].SnapshotDate != "2026-03-12" || rows[0].AssetName != "Old Asset" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}
	if rows[3].SnapshotDate != "2026-04-12" || rows[3].AssetName != "SET50 ETF" {
		t.Fatalf("unexpected last row: %+v", rows[3])
	}
}
