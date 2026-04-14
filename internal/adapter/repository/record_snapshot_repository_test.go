package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"

	dbfiles "github.com/guitarpawat/worthly-tracker/db"
	adapterdb "github.com/guitarpawat/worthly-tracker/internal/adapter/db"
	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

func TestRecordSnapshotRepository_GetSnapshotByOffsetReturnsLatestSnapshotWithOrderedItems(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	snapshot, err := repo.GetSnapshotByOffset(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetSnapshotByOffset returned error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if got := snapshot.RecordDate.Format("2006-01-02"); got != "2026-04-12" {
		t.Fatalf("expected latest snapshot date 2026-04-12, got %s", got)
	}
	if len(snapshot.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(snapshot.Items))
	}
	if snapshot.Items[0].AssetTypeName != "Cash" || snapshot.Items[0].AssetName != "Wallet" {
		t.Fatalf("expected cash item first, got %s/%s", snapshot.Items[0].AssetTypeName, snapshot.Items[0].AssetName)
	}
	if snapshot.Items[1].AssetTypeName != "Investment" || snapshot.Items[1].AssetName != "SET50 ETF" {
		t.Fatalf("expected investment item second, got %s/%s", snapshot.Items[1].AssetTypeName, snapshot.Items[1].AssetName)
	}

	olderSnapshot, err := repo.GetSnapshotByOffset(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetSnapshotByOffset returned error for offset 1: %v", err)
	}
	if olderSnapshot == nil {
		t.Fatal("expected older snapshot")
	}
	if got := olderSnapshot.RecordDate.Format("2006-01-02"); got != "2026-03-12" {
		t.Fatalf("expected older snapshot date 2026-03-12, got %s", got)
	}
}

func TestRecordSnapshotRepository_GetSnapshotByOffsetReturnsNilWhenOffsetIsOutOfRange(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	snapshot, err := repo.GetSnapshotByOffset(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetSnapshotByOffset returned error: %v", err)
	}
	if snapshot != nil {
		t.Fatal("expected nil snapshot for out-of-range offset")
	}
}

func TestRecordSnapshotRepository_ListSnapshotOptionsReturnsDatesInDescendingOrder(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	options, err := repo.ListSnapshotOptions(context.Background())
	if err != nil {
		t.Fatalf("ListSnapshotOptions returned error: %v", err)
	}
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
	if options[0].Offset != 0 || options[0].Label != "12 Apr 2026" {
		t.Fatalf("unexpected first option: %+v", options[0])
	}
	if options[1].Offset != 1 || options[1].Label != "12 Mar 2026" {
		t.Fatalf("unexpected second option: %+v", options[1])
	}
}

func TestRecordSnapshotRepository_GetEditableSnapshotByOffsetReturnsSnapshotAndAvailableAssets(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	snapshot, err := repo.GetEditableSnapshotByOffset(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetEditableSnapshotByOffset returned error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected editable snapshot")
	}
	if snapshot.ID != 2 {
		t.Fatalf("expected snapshot id 2, got %d", snapshot.ID)
	}
	if len(snapshot.Items) != 2 {
		t.Fatalf("expected 2 snapshot items, got %d", len(snapshot.Items))
	}
	if len(snapshot.AvailableAssets) != 0 {
		t.Fatalf("expected inactive assets to be excluded from available assets, got %+v", snapshot.AvailableAssets)
	}
}

func TestRecordSnapshotRepository_GetEditableSnapshotByOffsetExcludesAssetsWhoseTypeIsInactiveFromAvailableAssets(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	if _, err := database.Exec(`
		UPDATE asset_types
		SET is_active = FALSE
		WHERE id = 1
	`); err != nil {
		t.Fatalf("deactivate asset type: %v", err)
	}

	repo := NewRecordSnapshotRepository(database)

	snapshot, err := repo.GetEditableSnapshotByOffset(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetEditableSnapshotByOffset returned error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected editable snapshot")
	}
	if len(snapshot.AvailableAssets) != 1 {
		t.Fatalf("expected only active-type asset to remain available, got %+v", snapshot.AvailableAssets)
	}
	if snapshot.AvailableAssets[0].AssetID != 2 {
		t.Fatalf("expected cash asset to remain available, got %+v", snapshot.AvailableAssets)
	}
}

func TestRecordSnapshotRepository_GetNewSnapshotDraftUsesLatestSnapshotAutofillAndFiltersInactiveAssets(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	snapshot, err := repo.GetNewSnapshotDraft(
		context.Background(),
		parseSnapshotDate("2026-05-12"),
	)
	if err != nil {
		t.Fatalf("GetNewSnapshotDraft returned error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected editable snapshot draft")
	}
	if snapshot.ID != 0 {
		t.Fatalf("expected draft snapshot id 0, got %d", snapshot.ID)
	}
	if got := snapshot.RecordDate.Format("2006-01-02"); got != "2026-05-12" {
		t.Fatalf("expected draft date 2026-05-12, got %s", got)
	}
	if len(snapshot.Items) != 2 {
		t.Fatalf("expected 2 active autofill items, got %d", len(snapshot.Items))
	}
	if snapshot.Items[0].AssetID != 2 || snapshot.Items[1].AssetID != 1 {
		t.Fatalf("expected cash then investment ordering, got %+v", snapshot.Items)
	}
	if snapshot.Items[0].BoughtPrice != 0 {
		t.Fatalf("expected cash bought price 0, got %f", snapshot.Items[0].BoughtPrice)
	}
	if snapshot.Items[1].BoughtPrice != 12500 {
		t.Fatalf("expected auto incremented bought price 12500, got %f", snapshot.Items[1].BoughtPrice)
	}
	if snapshot.Items[1].CurrentPrice != 15000 {
		t.Fatalf("expected autofilled current price 15000, got %f", snapshot.Items[1].CurrentPrice)
	}
	if snapshot.Items[1].Remarks != "Current" {
		t.Fatalf("expected autofilled remarks Current, got %q", snapshot.Items[1].Remarks)
	}
	if len(snapshot.AvailableAssets) != 0 {
		t.Fatalf("expected inactive assets to be excluded from addable assets, got %+v", snapshot.AvailableAssets)
	}
}

func TestRecordSnapshotRepository_GetNewSnapshotDraftExcludesAssetsWhoseTypeIsInactive(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	if _, err := database.Exec(`
		UPDATE asset_types
		SET is_active = FALSE
		WHERE id = 1
	`); err != nil {
		t.Fatalf("deactivate asset type: %v", err)
	}

	repo := NewRecordSnapshotRepository(database)

	snapshot, err := repo.GetNewSnapshotDraft(context.Background(), parseSnapshotDate("2026-05-12"))
	if err != nil {
		t.Fatalf("GetNewSnapshotDraft returned error: %v", err)
	}

	if len(snapshot.Items) != 1 {
		t.Fatalf("expected only cash item from active asset type, got %d", len(snapshot.Items))
	}
	if snapshot.Items[0].AssetID != 2 {
		t.Fatalf("expected cash asset only, got %+v", snapshot.Items)
	}
	if len(snapshot.AvailableAssets) != 0 {
		t.Fatalf("expected inactive-type assets to be excluded from addable assets, got %+v", snapshot.AvailableAssets)
	}
}

func TestRecordSnapshotRepository_SaveSnapshotUpdatesDateSoftDeletesRemovedRowsAndAddsAssets(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	if _, err := database.Exec(`
		UPDATE assets
		SET is_active = TRUE
		WHERE id = 3
	`); err != nil {
		t.Fatalf("reactivate asset: %v", err)
	}

	repo := NewRecordSnapshotRepository(database)

	result, err := repo.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   2,
		SnapshotDate: "2026-04-15",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12500, CurrentPrice: 15800, Remarks: "Updated"},
			{AssetID: 3, BoughtPrice: 5100, CurrentPrice: 5200, Remarks: "Added"},
		},
	})
	if err != nil {
		t.Fatalf("SaveSnapshot returned error: %v", err)
	}
	if result.Offset != 0 {
		t.Fatalf("expected offset 0 after save, got %d", result.Offset)
	}

	snapshot, err := repo.GetEditableSnapshotByOffset(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetEditableSnapshotByOffset returned error after save: %v", err)
	}
	if got := snapshot.RecordDate.Format("2006-01-02"); got != "2026-04-15" {
		t.Fatalf("expected snapshot date 2026-04-15, got %s", got)
	}
	if len(snapshot.Items) != 2 {
		t.Fatalf("expected 2 active items, got %d", len(snapshot.Items))
	}
	if snapshot.Items[0].AssetID != 3 || snapshot.Items[1].AssetID != 1 {
		t.Fatalf("expected assets 3 and 1 after save, got %+v", snapshot.Items)
	}

	var deletedCount int
	if err := database.Get(&deletedCount, `
		SELECT COUNT(*)
		FROM record_items
		WHERE snapshot_id = 2
		  AND asset_id = 2
		  AND deleted_at IS NOT NULL
	`); err != nil {
		t.Fatalf("count deleted rows: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("expected 1 soft-deleted removed row, got %d", deletedCount)
	}
}

func TestRecordSnapshotRepository_SaveSnapshotForcesCashBoughtPriceToZero(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   2,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12000, CurrentPrice: 15000, Remarks: "ETF"},
			{AssetID: 2, BoughtPrice: 99999, CurrentPrice: 7000, Remarks: "Cash"},
		},
	})
	if err != nil {
		t.Fatalf("SaveSnapshot returned error: %v", err)
	}

	var boughtPrice float64
	if err := database.Get(&boughtPrice, `
		SELECT bought_price
		FROM record_items
		WHERE snapshot_id = 2
		  AND asset_id = 2
		  AND deleted_at IS NULL
	`); err != nil {
		t.Fatalf("get cash bought price: %v", err)
	}
	if boughtPrice != 0 {
		t.Fatalf("expected cash bought price to be 0, got %f", boughtPrice)
	}
}

func TestRecordSnapshotRepository_SaveSnapshotDeactivatesRemovedAssetsWhenEditingLatestSnapshot(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   2,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12000, CurrentPrice: 15000, Remarks: "ETF only"},
		},
	})
	if err != nil {
		t.Fatalf("SaveSnapshot returned error: %v", err)
	}

	var isActive bool
	if err := database.Get(&isActive, `
		SELECT is_active
		FROM assets
		WHERE id = 2
	`); err != nil {
		t.Fatalf("get removed asset active flag: %v", err)
	}
	if isActive {
		t.Fatal("expected removed asset from latest snapshot to be inactive")
	}

	reloadedSnapshot, err := repo.GetEditableSnapshotByOffset(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetEditableSnapshotByOffset returned error after save: %v", err)
	}
	if len(reloadedSnapshot.AvailableAssets) != 0 {
		t.Fatalf("expected inactive removed asset to stay hidden from available assets, got %+v", reloadedSnapshot.AvailableAssets)
	}
}

func TestRecordSnapshotRepository_CreateSnapshotDeactivatesAssetsRemovedFromPreviousLatestWhenNewSnapshotBecomesLatest(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-05-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12500, CurrentPrice: 16000, Remarks: "ETF only"},
		},
	})
	if err != nil {
		t.Fatalf("CreateSnapshot returned error: %v", err)
	}

	var isActive bool
	if err := database.Get(&isActive, `
		SELECT is_active
		FROM assets
		WHERE id = 2
	`); err != nil {
		t.Fatalf("get removed cash asset active flag: %v", err)
	}
	if isActive {
		t.Fatal("expected omitted asset from new latest snapshot to be inactive")
	}

	newDraft, err := repo.GetNewSnapshotDraft(context.Background(), parseSnapshotDate("2026-06-12"))
	if err != nil {
		t.Fatalf("GetNewSnapshotDraft returned error after create: %v", err)
	}
	if len(newDraft.Items) != 1 || newDraft.Items[0].AssetID != 1 {
		t.Fatalf("expected only remaining active asset in next draft, got %+v", newDraft.Items)
	}
	if len(newDraft.AvailableAssets) != 0 {
		t.Fatalf("expected deactivated asset to stay hidden from addable assets, got %+v", newDraft.AvailableAssets)
	}
}

func TestRecordSnapshotRepository_CreateSnapshotDoesNotDeactivateAssetsWhenNewSnapshotIsNotLatest(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-02-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 9000, CurrentPrice: 9100, Remarks: "Older snapshot"},
		},
	})
	if err != nil {
		t.Fatalf("CreateSnapshot returned error: %v", err)
	}

	var isActive bool
	if err := database.Get(&isActive, `
		SELECT is_active
		FROM assets
		WHERE id = 2
	`); err != nil {
		t.Fatalf("get current latest asset active flag: %v", err)
	}
	if !isActive {
		t.Fatal("expected asset to stay active when created snapshot is not latest")
	}
}

func TestRecordSnapshotRepository_CreateSnapshotRejectsInactiveAssets(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-05-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 3, BoughtPrice: 5000, CurrentPrice: 5100, Remarks: "Inactive"},
		},
	})
	if !errors.Is(err, recorderr.ErrAssetUnavailable) {
		t.Fatalf("expected inactive asset to be rejected, got %v", err)
	}
}

func TestRecordSnapshotRepository_SaveSnapshotRejectsAddingInactiveAssets(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   2,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12000, CurrentPrice: 15000, Remarks: "Current"},
			{AssetID: 2, BoughtPrice: 0, CurrentPrice: 7000, Remarks: "Cash"},
			{AssetID: 3, BoughtPrice: 5000, CurrentPrice: 5100, Remarks: "Inactive"},
		},
	})
	if !errors.Is(err, recorderr.ErrAssetUnavailable) {
		t.Fatalf("expected inactive asset addition to be rejected, got %v", err)
	}
}

func TestRecordSnapshotRepository_SaveSnapshotReturnsConflictForDuplicateDate(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   2,
		SnapshotDate: "2026-03-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12000, CurrentPrice: 15000},
		},
	})
	if !errors.Is(err, recorderr.ErrSnapshotDateAlreadyExists) {
		t.Fatalf("expected duplicate date error, got %v", err)
	}
}

func TestRecordSnapshotRepository_CreateSnapshotInsertsItemsAndReturnsLatestOffset(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	if _, err := database.Exec(`
		UPDATE assets
		SET is_active = TRUE
		WHERE id = 3
	`); err != nil {
		t.Fatalf("reactivate asset: %v", err)
	}

	repo := NewRecordSnapshotRepository(database)

	result, err := repo.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-05-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 12500, CurrentPrice: 16000, Remarks: "New month"},
			{AssetID: 2, BoughtPrice: 99999, CurrentPrice: 7200, Remarks: "Cash"},
			{AssetID: 3, BoughtPrice: 5100, CurrentPrice: 5300, Remarks: "Re-added inactive"},
		},
	})
	if err != nil {
		t.Fatalf("CreateSnapshot returned error: %v", err)
	}
	if result.Offset != 0 {
		t.Fatalf("expected latest offset 0, got %d", result.Offset)
	}

	snapshot, err := repo.GetSnapshotByOffset(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetSnapshotByOffset returned error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected created snapshot")
	}
	if got := snapshot.RecordDate.Format("2006-01-02"); got != "2026-05-12" {
		t.Fatalf("expected snapshot date 2026-05-12, got %s", got)
	}
	if len(snapshot.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(snapshot.Items))
	}

	var cashBoughtPrice float64
	if err := database.Get(&cashBoughtPrice, `
		SELECT bought_price
		FROM record_items ri
		INNER JOIN record_snapshots rs ON rs.id = ri.snapshot_id
		WHERE rs.record_date = '2026-05-12'
		  AND ri.asset_id = 2
		  AND rs.deleted_at IS NULL
		  AND ri.deleted_at IS NULL
	`); err != nil {
		t.Fatalf("get created cash bought price: %v", err)
	}
	if cashBoughtPrice != 0 {
		t.Fatalf("expected created cash bought price 0, got %f", cashBoughtPrice)
	}
}

func TestRecordSnapshotRepository_CreateSnapshotReturnsConflictForDuplicateDate(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	_, err := repo.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 1, CurrentPrice: 1},
		},
	})
	if !errors.Is(err, recorderr.ErrSnapshotDateAlreadyExists) {
		t.Fatalf("expected duplicate date error, got %v", err)
	}
}

func TestRecordSnapshotRepository_DeleteSnapshotSoftDeletesSnapshotAndItems(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	result, err := repo.DeleteSnapshot(context.Background(), dto.DeleteSnapshotInput{
		SnapshotID: 2,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("DeleteSnapshot returned error: %v", err)
	}
	if !result.HasSnapshots {
		t.Fatal("expected remaining snapshots")
	}
	if result.Offset != 0 {
		t.Fatalf("expected next offset 0, got %d", result.Offset)
	}

	var deletedSnapshots int
	if err := database.Get(&deletedSnapshots, `
		SELECT COUNT(*)
		FROM record_snapshots
		WHERE id = 2
		  AND deleted_at IS NOT NULL
	`); err != nil {
		t.Fatalf("count deleted snapshots: %v", err)
	}
	if deletedSnapshots != 1 {
		t.Fatalf("expected snapshot to be soft deleted, got %d", deletedSnapshots)
	}

	var deletedItems int
	if err := database.Get(&deletedItems, `
		SELECT COUNT(*)
		FROM record_items
		WHERE snapshot_id = 2
		  AND deleted_at IS NOT NULL
	`); err != nil {
		t.Fatalf("count deleted items: %v", err)
	}
	if deletedItems != 2 {
		t.Fatalf("expected 2 deleted items, got %d", deletedItems)
	}

	snapshot, err := repo.GetSnapshotByOffset(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetSnapshotByOffset returned error after delete: %v", err)
	}
	if snapshot == nil || snapshot.RecordDate.Format("2006-01-02") != "2026-03-12" {
		t.Fatalf("expected remaining snapshot 2026-03-12, got %+v", snapshot)
	}
}

func TestRecordSnapshotRepository_DeleteSnapshotReturnsNoSnapshotsWhenDeletingLastSnapshot(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertSingleSnapshotData(t, database)

	repo := NewRecordSnapshotRepository(database)

	result, err := repo.DeleteSnapshot(context.Background(), dto.DeleteSnapshotInput{
		SnapshotID: 1,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("DeleteSnapshot returned error: %v", err)
	}
	if result.HasSnapshots {
		t.Fatal("expected no remaining snapshots")
	}
	if result.Offset != 0 {
		t.Fatalf("expected offset 0, got %d", result.Offset)
	}
}

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	database, err := adapterdb.Open(adapterdb.SQLiteConfig{Path: ":memory:"})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	if err := adapterdb.ApplyMigrations(context.Background(), database, dbfiles.FS); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return database
}

func insertTestSnapshotData(t *testing.T, database *sqlx.DB) {
	t.Helper()

	statements := []string{
		`INSERT INTO asset_types (id, name, ordering) VALUES (1, 'Investment', 2), (2, 'Cash', 1)`,
		`INSERT INTO assets (id, asset_type_id, name, broker, is_cash, is_active, auto_increment, ordering) VALUES
			(1, 1, 'SET50 ETF', 'KKP', FALSE, TRUE, 500, 2),
			(2, 2, 'Wallet', 'SCB', TRUE, TRUE, 0, 1),
			(3, 1, 'Old Asset', 'IBKR', FALSE, FALSE, 0, 1)`,
		`INSERT INTO record_snapshots (id, record_date) VALUES (1, '2026-03-12'), (2, '2026-04-12')`,
		`INSERT INTO record_items (id, snapshot_id, asset_id, bought_price, current_price, remarks) VALUES
			(1, 1, 1, 10000, 14000, 'Prev'),
			(2, 1, 3, 5000, 4500, 'Old'),
			(3, 2, 1, 12000, 15000, 'Current'),
			(4, 2, 2, 7000, 7000, 'Cash')`,
	}

	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			t.Fatalf("exec statement %q: %v", statement, err)
		}
	}
}

func insertSingleSnapshotData(t *testing.T, database *sqlx.DB) {
	t.Helper()

	statements := []string{
		`INSERT INTO asset_types (id, name, ordering) VALUES (1, 'Investment', 1)`,
		`INSERT INTO assets (id, asset_type_id, name, broker, is_cash, is_active, auto_increment, ordering) VALUES
			(1, 1, 'SET50 ETF', 'KKP', FALSE, TRUE, 0, 1)`,
		`INSERT INTO record_snapshots (id, record_date) VALUES (1, '2026-04-12')`,
		`INSERT INTO record_items (id, snapshot_id, asset_id, bought_price, current_price, remarks) VALUES
			(1, 1, 1, 10000, 12000, 'Only')`,
	}

	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			t.Fatalf("exec statement %q: %v", statement, err)
		}
	}
}
