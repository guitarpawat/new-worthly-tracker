package validator

import (
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

func TestNormalizeCustomAllocation(t *testing.T) {
	t.Parallel()
	entry := func(asset int64, category, percentage string) dto.CustomAllocationEntry {
		return dto.CustomAllocationEntry{AssetID: asset, Category: category, Percentage: decimal.RequireFromString(percentage)}
	}
	tests := []struct {
		name  string
		chart dto.CustomAllocationChart
		valid bool
	}{
		{name: "empty chart", chart: dto.CustomAllocationChart{Name: "Country"}, valid: true},
		{name: "missing name", chart: dto.CustomAllocationChart{Name: "  "}},
		{name: "partial", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "70")}}, valid: true},
		{name: "exact hundred", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "33.33"), entry(1, "China", "66.67")}}, valid: true},
		{name: "over hundred", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "33.34"), entry(1, "China", "66.67")}}},
		{name: "independent assets", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "100"), entry(2, "USA", "100")}}, valid: true},
		{name: "duplicate category", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, " USA ", "10"), entry(1, "usa", "10")}}},
		{name: "missing asset", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(0, "USA", "10")}}},
		{name: "missing category", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, " ", "10")}}},
		{name: "reserved category", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, " UNALLOCATED ", "10")}}},
		{name: "zero", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "0")}}},
		{name: "negative", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "-1")}}},
		{name: "excess precision", chart: dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{entry(1, "USA", "0.001")}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NormalizeCustomAllocation(tc.chart)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
	normalized, err := NormalizeCustomAllocation(dto.CustomAllocationChart{Name: " Country ", Entries: []dto.CustomAllocationEntry{
		entry(1, " USA ", "10"), entry(2, "usa", "100"),
	}})
	if err != nil || normalized.Name != "Country" || normalized.Entries[1].Category != "USA" {
		t.Fatalf("normalization failed: %+v, %v", normalized, err)
	}
}

func TestNormalizeAssetChartAllocations(t *testing.T) {
	t.Parallel()
	input := []dto.AssetChartAllocation{{ChartID: 1, Entries: []dto.CustomAllocationEntry{
		{AssetID: 999, Category: " USA ", Percentage: decimal.NewFromInt(50)},
	}}}
	normalized, err := NormalizeAssetChartAllocations(2, input)
	if err != nil || normalized[0].Entries[0].AssetID != 2 || normalized[0].Entries[0].Category != "USA" {
		t.Fatalf("normalize asset chart: %+v %v", normalized, err)
	}
	if input[0].Entries[0].AssetID != 999 {
		t.Fatal("caller payload mutated")
	}
	for _, tc := range []struct {
		name   string
		charts []dto.AssetChartAllocation
	}{
		{name: "duplicate chart", charts: append(input, input[0])},
		{name: "invalid chart", charts: []dto.AssetChartAllocation{{ChartID: 0}}},
		{name: "excess allocation", charts: []dto.AssetChartAllocation{{ChartID: 1, Entries: []dto.CustomAllocationEntry{
			{Category: "USA", Percentage: decimal.NewFromInt(101)},
		}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NormalizeAssetChartAllocations(2, tc.charts); err == nil {
				t.Fatal("invalid allocations accepted")
			}
		})
	}
	chart, err := NormalizeCustomAllocation(dto.CustomAllocationChart{ID: 1, Name: "Rename"})
	if err != nil || chart.Entries != nil {
		t.Fatalf("omitted entries must remain omitted: %+v %v", chart, err)
	}
}
