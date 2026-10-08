package service

import (
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

func TestCustomAllocationsWeightEachAssetAndIncludeRemainder(t *testing.T) {
	t.Parallel()
	charts := []dto.CustomAllocationChart{
		{ID: 1, Name: "Country", Entries: []dto.CustomAllocationEntry{
			{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(50)},
			{AssetID: 1, Category: "China", Percentage: decimal.NewFromInt(15)},
			{AssetID: 1, Category: "Japan", Percentage: decimal.NewFromInt(5)},
			{AssetID: 2, Category: "China", Percentage: decimal.NewFromInt(100)},
		}},
		{ID: 2, Name: "Type", Entries: []dto.CustomAllocationEntry{
			{AssetID: 1, Category: "Stock", Percentage: decimal.NewFromInt(100)},
			{AssetID: 2, Category: "Bond", Percentage: decimal.NewFromInt(100)},
			{AssetID: 3, Category: "Gold", Percentage: decimal.NewFromInt(100)},
		}},
	}
	items := []dto.ProgressSnapshotItem{
		{AssetID: 1, AssetName: "Same name", CurrentPrice: decimal.NewFromInt(1000)},
		{AssetID: 2, AssetName: "Same name", CurrentPrice: decimal.NewFromInt(2000)},
		{AssetID: 3, CurrentPrice: decimal.NewFromInt(3000)},
	}
	got := buildCustomAllocations(charts, items)
	assertAllocationValues(t, got[0].Rows, map[string]string{"China": "2150", "Japan": "50", "USA": "500", "Unallocated": "3300"})
	assertAllocationValues(t, got[1].Rows, map[string]string{"Stock": "1000", "Bond": "2000", "Gold": "3000"})
}

func TestCustomAllocationsPreserveSignedValuesAndPrecision(t *testing.T) {
	t.Parallel()
	charts := []dto.CustomAllocationChart{{ID: 1, Name: "Country", Entries: []dto.CustomAllocationEntry{
		{AssetID: 1, Category: "USA", Percentage: decimal.RequireFromString("33.33")},
	}}}
	for _, amount := range []string{"-123.45", "0", "0.00000000000000000001", "9007199254740993.01"} {
		t.Run(amount, func(t *testing.T) {
			value := decimal.RequireFromString(amount)
			got := buildCustomAllocations(charts, []dto.ProgressSnapshotItem{{AssetID: 1, CurrentPrice: value}})
			var sum decimal.Decimal
			for _, row := range got[0].Rows {
				sum = sum.Add(row.Value)
			}
			if !sum.Equal(value) {
				t.Fatalf("total changed: %s != %s", sum, value)
			}
			if !got[0].Rows[0].Value.Equal(value.Mul(decimal.RequireFromString("0.3333"))) {
				t.Fatalf("wrong weighted value: %s", got[0].Rows[0].Value)
			}
		})
	}
}

func assertAllocationValues(t *testing.T, rows []dto.AllocationSlice, want map[string]string) {
	t.Helper()
	if len(rows) != len(want) {
		t.Fatalf("rows: %+v", rows)
	}
	for _, row := range rows {
		expected, ok := want[row.Name]
		if !ok || !row.Value.Equal(decimal.RequireFromString(expected)) {
			t.Fatalf("unexpected slice: %+v", row)
		}
	}
}

func TestCustomAllocationsExcludeAssetsOnlyFromSelectedChart(t *testing.T) {
	t.Parallel()
	charts := []dto.CustomAllocationChart{
		{ID: 1, Name: "Country", ExcludedAssetIDs: []int64{1, 3}, Entries: []dto.CustomAllocationEntry{
			{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(50)},
			{AssetID: 2, Category: "China", Percentage: decimal.NewFromInt(70)},
		}},
		{ID: 2, Name: "Default included"},
		{ID: 3, Name: "All excluded", ExcludedAssetIDs: []int64{1, 2, 3}},
	}
	items := []dto.ProgressSnapshotItem{
		{AssetID: 1, CurrentPrice: decimal.NewFromInt(100)},
		{AssetID: 2, CurrentPrice: decimal.NewFromInt(200)},
		{AssetID: 3, CurrentPrice: decimal.NewFromInt(-50)},
	}
	got := buildCustomAllocations(charts, items)
	assertAllocationValues(t, got[0].Rows, map[string]string{"China": "140", "Unallocated": "60"})
	assertAllocationValues(t, got[1].Rows, map[string]string{"Unallocated": "250"})
	assertAllocationValues(t, got[2].Rows, map[string]string{"Unallocated": "0"})
}
