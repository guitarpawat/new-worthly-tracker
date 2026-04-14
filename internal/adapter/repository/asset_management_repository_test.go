package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

func TestAssetManagementRepository_GetPageReturnsTypesAssetsAndActiveTypeOptions(t *testing.T) {
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

	repo := NewAssetManagementRepository(database)

	page, err := repo.GetPage(context.Background())
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	if len(page.AssetTypes) != 2 {
		t.Fatalf("expected 2 asset types, got %d", len(page.AssetTypes))
	}
	if len(page.Assets) != 3 {
		t.Fatalf("expected 3 assets, got %d", len(page.Assets))
	}
	if len(page.ActiveAssetTypes) != 1 {
		t.Fatalf("expected 1 active asset type option, got %d", len(page.ActiveAssetTypes))
	}
	if page.ActiveAssetTypes[0].Name != "Cash" {
		t.Fatalf("expected only cash active option, got %+v", page.ActiveAssetTypes)
	}
}

func TestAssetManagementRepository_CreateAndUpdateAssetType(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	createResult, err := repo.CreateAssetType(context.Background(), dto.CreateAssetTypeInput{
		Name:     "Property",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateAssetType returned error: %v", err)
	}

	updateResult, err := repo.UpdateAssetType(context.Background(), dto.UpdateAssetTypeInput{
		ID:       createResult.ID,
		Name:     "Property & Land",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("UpdateAssetType returned error: %v", err)
	}
	if updateResult.ID != createResult.ID {
		t.Fatalf("expected same id after update, got %d", updateResult.ID)
	}

	page, err := repo.GetPage(context.Background())
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	var found bool
	for _, item := range page.AssetTypes {
		if item.ID == createResult.ID {
			found = true
			if item.Name != "Property & Land" {
				t.Fatalf("expected updated name, got %q", item.Name)
			}
			if item.IsActive {
				t.Fatal("expected updated asset type to be inactive")
			}
		}
	}
	if !found {
		t.Fatal("expected created asset type in page")
	}
}

func TestAssetManagementRepository_CreateAndUpdateAsset(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	createResult, err := repo.CreateAsset(context.Background(), dto.CreateAssetInput{
		Name:          "Thai Bond Fund",
		AssetTypeID:   1,
		Broker:        "KKP",
		IsCash:        false,
		IsActive:      true,
		AutoIncrement: 2500,
	})
	if err != nil {
		t.Fatalf("CreateAsset returned error: %v", err)
	}

	updateResult, err := repo.UpdateAsset(context.Background(), dto.UpdateAssetInput{
		ID:            createResult.ID,
		Name:          "Cash Bucket",
		AssetTypeID:   2,
		Broker:        "SCB",
		IsCash:        true,
		IsActive:      false,
		AutoIncrement: 9999,
	})
	if err != nil {
		t.Fatalf("UpdateAsset returned error: %v", err)
	}
	if updateResult.ID != createResult.ID {
		t.Fatalf("expected same id after update, got %d", updateResult.ID)
	}

	page, err := repo.GetPage(context.Background())
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	var found bool
	for _, item := range page.Assets {
		if item.ID == createResult.ID {
			found = true
			if item.Name != "Cash Bucket" {
				t.Fatalf("expected updated name, got %q", item.Name)
			}
			if !item.IsCash {
				t.Fatal("expected updated asset to be cash")
			}
			if item.IsActive {
				t.Fatal("expected updated asset to be inactive")
			}
			if item.AutoIncrement != 0 {
				t.Fatalf("expected cash auto increment forced to 0, got %f", item.AutoIncrement)
			}
		}
	}
	if !found {
		t.Fatal("expected created asset in page")
	}
}

func TestAssetManagementRepository_CreateAssetRejectsInactiveAssetType(t *testing.T) {
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

	repo := NewAssetManagementRepository(database)

	_, err := repo.CreateAsset(context.Background(), dto.CreateAssetInput{
		Name:          "Should Fail",
		AssetTypeID:   1,
		Broker:        "KKP",
		IsCash:        false,
		IsActive:      true,
		AutoIncrement: 100,
	})
	if !errors.Is(err, recorderr.ErrAssetTypeInactive) {
		t.Fatalf("expected inactive asset type error, got %v", err)
	}
}

func TestAssetManagementRepository_CreateAssetTypeRejectsDuplicateName(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	_, err := repo.CreateAssetType(context.Background(), dto.CreateAssetTypeInput{
		Name:     " investment ",
		IsActive: true,
	})
	if !errors.Is(err, recorderr.ErrAssetTypeNameExists) {
		t.Fatalf("expected duplicate asset type name error, got %v", err)
	}
}

func TestAssetManagementRepository_UpdateAssetTypeRejectsDuplicateName(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	_, err := repo.UpdateAssetType(context.Background(), dto.UpdateAssetTypeInput{
		ID:       2,
		Name:     "investment",
		IsActive: true,
	})
	if !errors.Is(err, recorderr.ErrAssetTypeNameExists) {
		t.Fatalf("expected duplicate asset type name error, got %v", err)
	}
}

func TestAssetManagementRepository_CreateAssetRejectsDuplicateNameInsideSameType(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	_, err := repo.CreateAsset(context.Background(), dto.CreateAssetInput{
		Name:          " wallet ",
		AssetTypeID:   2,
		Broker:        "SCB",
		IsCash:        true,
		IsActive:      true,
		AutoIncrement: 0,
	})
	if !errors.Is(err, recorderr.ErrAssetNameExists) {
		t.Fatalf("expected duplicate asset name error, got %v", err)
	}
}

func TestAssetManagementRepository_UpdateAssetRejectsDuplicateNameInsideSameType(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	_, err := repo.UpdateAsset(context.Background(), dto.UpdateAssetInput{
		ID:            3,
		Name:          "set50 etf",
		AssetTypeID:   1,
		Broker:        "IBKR",
		IsCash:        false,
		IsActive:      true,
		AutoIncrement: 100,
	})
	if !errors.Is(err, recorderr.ErrAssetNameExists) {
		t.Fatalf("expected duplicate asset name error, got %v", err)
	}
}

func TestAssetManagementRepository_ReorderAssetTypesUpdatesOrdering(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)

	repo := NewAssetManagementRepository(database)

	err := repo.ReorderAssetTypes(context.Background(), dto.ReorderAssetTypesInput{
		OrderedIDs: []int64{1, 2},
	})
	if err != nil {
		t.Fatalf("ReorderAssetTypes returned error: %v", err)
	}

	page, err := repo.GetPage(context.Background())
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	if page.AssetTypes[0].ID != 1 || page.AssetTypes[1].ID != 2 {
		t.Fatalf("expected reordered asset types, got %+v", page.AssetTypes)
	}
}

func TestAssetManagementRepository_ReorderAssetTypesActiveOnlyPreservesInactiveSlots(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	if _, err := database.Exec(`
		INSERT INTO asset_types (id, name, is_active, ordering)
		VALUES (3, 'Property', TRUE, 3)
	`); err != nil {
		t.Fatalf("insert asset type: %v", err)
	}
	if _, err := database.Exec(`
		UPDATE asset_types
		SET is_active = FALSE
		WHERE id = 2
	`); err != nil {
		t.Fatalf("deactivate asset type: %v", err)
	}

	repo := NewAssetManagementRepository(database)

	err := repo.ReorderAssetTypes(context.Background(), dto.ReorderAssetTypesInput{
		OrderedIDs: []int64{3, 1},
		ActiveOnly: true,
	})
	if err != nil {
		t.Fatalf("ReorderAssetTypes returned error: %v", err)
	}

	page, err := repo.GetPage(context.Background())
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	got := []int64{page.AssetTypes[0].ID, page.AssetTypes[1].ID, page.AssetTypes[2].ID}
	want := []int64{2, 3, 1}
	if got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("expected inactive slot to stay stable, got %v want %v", got, want)
	}
}

func TestAssetManagementRepository_ReorderAssetsActiveOnlyPreservesInactiveSlotsWithinType(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	if _, err := database.Exec(`
		INSERT INTO assets (id, asset_type_id, name, broker, is_cash, is_active, auto_increment, ordering)
		VALUES (4, 1, 'Growth Fund', 'KKP', FALSE, TRUE, 1000, 3)
	`); err != nil {
		t.Fatalf("insert asset: %v", err)
	}

	repo := NewAssetManagementRepository(database)

	err := repo.ReorderAssets(context.Background(), dto.ReorderAssetInput{
		AssetTypeID: 1,
		OrderedIDs:  []int64{4, 1},
		ActiveOnly:  true,
	})
	if err != nil {
		t.Fatalf("ReorderAssets returned error: %v", err)
	}

	page, err := repo.GetPage(context.Background())
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	var got []int64
	for _, asset := range page.Assets {
		if asset.AssetTypeID == 1 {
			got = append(got, asset.ID)
		}
	}

	want := []int64{3, 4, 1}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("expected inactive slot to stay stable within asset type, got %v want %v", got, want)
	}
}
