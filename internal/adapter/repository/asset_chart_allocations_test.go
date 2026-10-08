package repository

import (
	"context"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

func TestAssetChartAllocationsSaveTogetherAndPreserveOtherAssets(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	insertTestSnapshotData(t, db)
	ctx := context.Background()
	charts := NewCustomAllocationRepository(db)
	id, err := charts.SaveChart(ctx, dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{
		{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(70)},
		{AssetID: 2, Category: "China", Percentage: decimal.NewFromInt(100)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	// Chart management submits only the name, without overwriting any assignments.
	if _, err := charts.SaveChart(ctx, dto.CustomAllocationChart{ID: id, Name: "Asset Country"}); err != nil {
		t.Fatal(err)
	}
	list, err := charts.ListCharts(ctx)
	if err != nil || len(list[0].Entries) != 2 {
		t.Fatalf("rename lost entries: %+v %v", list, err)
	}
	assets := NewAssetManagementRepository(db)
	input := dto.UpdateAssetInput{ID: 1, Name: "Edited Fund", AssetTypeID: 1, IsActive: true,
		ChartAllocations: []dto.AssetChartAllocation{{ChartID: id, Entries: []dto.CustomAllocationEntry{
			{AssetID: 1, Category: "Japan", Percentage: decimal.NewFromInt(50)},
		}}},
	}
	if _, err := assets.UpdateAsset(ctx, input); err != nil {
		t.Fatal(err)
	}
	list, err = charts.ListCharts(ctx)
	if err != nil || len(list[0].Entries) != 2 {
		t.Fatalf("saved entries: %+v %v", list, err)
	}
	if list[0].Entries[0].Category != "Japan" || list[0].Entries[1].Category != "China" {
		t.Fatalf("other asset modified: %+v", list[0].Entries)
	}
	input.Name = "Should roll back"
	input.ChartAllocations[0].Entries[0].Category = "Changed"
	input.ChartAllocations = append(input.ChartAllocations, dto.AssetChartAllocation{ChartID: 9999})
	if _, err := assets.UpdateAsset(ctx, input); err == nil {
		t.Fatal("missing chart accepted")
	}
	page, err := assets.GetPage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, asset := range page.Assets {
		if asset.ID == 1 && asset.Name != "Edited Fund" {
			t.Fatalf("asset write was not rolled back: %+v", asset)
		}
	}
	list, err = charts.ListCharts(ctx)
	if err != nil || list[0].Entries[0].Category != "Japan" {
		t.Fatalf("tag write was not rolled back: %+v %v", list, err)
	}
	input.Name = "Edited Fund"
	input.ChartAllocations = []dto.AssetChartAllocation{{ChartID: id, Entries: []dto.CustomAllocationEntry{}}}
	if _, err := assets.UpdateAsset(ctx, input); err != nil {
		t.Fatal(err)
	}
	list, err = charts.ListCharts(ctx)
	if err != nil || len(list[0].Entries) != 1 || list[0].Entries[0].AssetID != 2 {
		t.Fatalf("clear did not preserve other asset: %+v %v", list, err)
	}
	if err := charts.DeleteChart(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := assets.UpdateAsset(ctx, input); err == nil {
		t.Fatal("deleted chart accepted")
	}
}

func TestAssetChartExclusionDefaultsAndSavedTags(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	insertTestSnapshotData(t, db)
	ctx := context.Background()
	charts := NewCustomAllocationRepository(db)
	entry := dto.CustomAllocationEntry{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(70)}
	id, err := charts.SaveChart(ctx, dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry}})
	if err != nil {
		t.Fatal(err)
	}
	list, err := charts.ListCharts(ctx)
	if err != nil || len(list[0].ExcludedAssetIDs) != 0 {
		t.Fatalf("assets must default to included: %+v %v", list, err)
	}
	assets := NewAssetManagementRepository(db)
	input := dto.UpdateAssetInput{ID: 1, Name: "Fund", AssetTypeID: 1, IsActive: true,
		ChartAllocations: []dto.AssetChartAllocation{{ChartID: id, Excluded: true}},
	}
	if _, err := assets.UpdateAsset(ctx, input); err != nil {
		t.Fatal(err)
	}
	if _, err := charts.SaveChart(ctx, dto.CustomAllocationChart{ID: id, Name: "Renamed"}); err != nil {
		t.Fatal(err)
	}
	list, err = charts.ListCharts(ctx)
	if err != nil || len(list[0].ExcludedAssetIDs) != 1 || list[0].ExcludedAssetIDs[0] != 1 || len(list[0].Entries) != 1 {
		t.Fatalf("exclusion or saved tags lost after rename: %+v %v", list, err)
	}
	input.ChartAllocations = []dto.AssetChartAllocation{
		{ChartID: id, Entries: []dto.CustomAllocationEntry{entry}}, {ChartID: 9999},
	}
	if _, err := assets.UpdateAsset(ctx, input); err == nil {
		t.Fatal("invalid chart accepted")
	}
	list, err = charts.ListCharts(ctx)
	if err != nil || len(list[0].ExcludedAssetIDs) != 1 {
		t.Fatalf("failed save changed exclusion: %+v %v", list, err)
	}
	input.ChartAllocations = input.ChartAllocations[:1]
	if _, err := assets.UpdateAsset(ctx, input); err != nil {
		t.Fatal(err)
	}
	list, err = charts.ListCharts(ctx)
	if err != nil || len(list[0].ExcludedAssetIDs) != 0 || len(list[0].Entries) != 1 {
		t.Fatalf("re-inclusion failed: %+v %v", list, err)
	}
	var archived int
	if err := db.Get(&archived, `SELECT COUNT(*) FROM custom_allocation_exclusions WHERE deleted_at IS NOT NULL`); err != nil || archived != 1 {
		t.Fatalf("exclusion was not soft deleted: %d %v", archived, err)
	}
}
