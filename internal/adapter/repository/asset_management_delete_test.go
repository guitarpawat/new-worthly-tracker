package repository

import (
	"context"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

func TestDeleteAssetAndTypeGuards(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	insertTestSnapshotData(t, db)
	ctx := context.Background()
	repo := NewAssetManagementRepository(db)
	for _, id := range []int64{1, 2, 3} {
		if err := repo.DeleteAsset(ctx, id); err == nil {
			t.Fatalf("asset %d with records deleted", id)
		}
	}
	if _, err := db.Exec(`UPDATE record_items SET deleted_at = CURRENT_TIMESTAMP; UPDATE record_snapshots SET deleted_at = CURRENT_TIMESTAMP`); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteAsset(ctx, 1); err == nil {
		t.Fatal("archived snapshot references ignored")
	}
	if err := repo.DeleteAssetType(ctx, 1); err == nil {
		t.Fatal("nonempty type deleted")
	}
	typeResult, err := repo.CreateAssetType(ctx, dto.CreateAssetTypeInput{Name: "Temporary", IsActive: true})
	if err != nil {
		t.Fatal(err)
	}
	asset, err := repo.CreateAsset(ctx, dto.CreateAssetInput{Name: "Unused", AssetTypeID: typeResult.ID, IsActive: false})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteAssetType(ctx, typeResult.ID); err == nil {
		t.Fatal("inactive asset failed to block type deletion")
	}
	page, err := repo.GetPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range page.Assets {
		if row.CanDelete != (row.ID == asset.ID) {
			t.Fatalf("incorrect deletion eligibility: %+v", row)
		}
	}
	if err := repo.DeleteAsset(ctx, asset.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteAssetType(ctx, typeResult.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM assets WHERE id = ? AND deleted_at IS NOT NULL`, asset.ID); err != nil || count != 1 {
		t.Fatalf("asset not soft deleted: %d %v", count, err)
	}
	if err := db.Get(&count, `SELECT COUNT(*) FROM asset_types WHERE id = ? AND deleted_at IS NOT NULL`, typeResult.ID); err != nil || count != 1 {
		t.Fatalf("type not soft deleted: %d %v", count, err)
	}
	if err := repo.DeleteAsset(ctx, asset.ID); err == nil {
		t.Fatal("deleted asset accepted")
	}
	if err := repo.DeleteAssetType(ctx, 99999); err == nil {
		t.Fatal("unknown type accepted")
	}
}
