package validator

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

// NormalizeCustomAllocation validates percentages with decimal arithmetic before any write.
func NormalizeCustomAllocation(input dto.CustomAllocationChart) (dto.CustomAllocationChart, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.ID < 0 || input.Name == "" || utf8.RuneCountInString(input.Name) > 100 {
		return input, fmt.Errorf("chart name is required and must be at most 100 characters")
	}
	entries := make([]dto.CustomAllocationEntry, 0, len(input.Entries))
	totals := map[int64]decimal.Decimal{}
	seen := map[int64]map[string]bool{}
	labels := map[string]string{}
	for _, entry := range input.Entries {
		entry.Category = strings.TrimSpace(entry.Category)
		key := strings.ToLower(entry.Category)
		if entry.AssetID <= 0 || key == "" || utf8.RuneCountInString(entry.Category) > 100 {
			return input, fmt.Errorf("each allocation needs an asset and a category of at most 100 characters")
		}
		if key == "unallocated" {
			return input, fmt.Errorf("Unallocated is reserved for the remaining percentage")
		}
		if !entry.Percentage.IsPositive() || entry.Percentage.GreaterThan(decimal.NewFromInt(100)) {
			return input, fmt.Errorf("allocation percentage must be greater than 0 and at most 100")
		}
		if hasMoreThanTwoDecimalPlaces(entry.Percentage) {
			return input, fmt.Errorf("allocation percentage allows at most 2 decimal places")
		}
		if seen[entry.AssetID] == nil {
			seen[entry.AssetID] = map[string]bool{}
		}
		if seen[entry.AssetID][key] {
			return input, fmt.Errorf("an asset cannot have the same category twice in one chart")
		}
		seen[entry.AssetID][key] = true
		totals[entry.AssetID] = totals[entry.AssetID].Add(entry.Percentage)
		if totals[entry.AssetID].GreaterThan(decimal.NewFromInt(100)) {
			return input, fmt.Errorf("allocation percentages for each asset must total at most 100%%")
		}
		if label, ok := labels[key]; ok {
			entry.Category = label
		} else {
			labels[key] = entry.Category
		}
		entries = append(entries, entry)
	}
	// Omitted entries mean a chart-name-only edit; preserve existing asset tags.
	if input.Entries != nil {
		input.Entries = entries
	}
	return input, nil
}

func NormalizeAssetChartAllocations(assetID int64, charts []dto.AssetChartAllocation) ([]dto.AssetChartAllocation, error) {
	if charts == nil {
		return nil, nil
	}
	result := make([]dto.AssetChartAllocation, 0, len(charts))
	seen := map[int64]bool{}
	for _, chart := range charts {
		if chart.ChartID <= 0 || seen[chart.ChartID] {
			return nil, fmt.Errorf("chart IDs must be positive and unique")
		}
		seen[chart.ChartID] = true
		if chart.Excluded {
			// Excluding an asset preserves its saved tags rather than replacing them.
			result = append(result, dto.AssetChartAllocation{ChartID: chart.ChartID, Excluded: true})
			continue
		}
		entries := make([]dto.CustomAllocationEntry, 0, len(chart.Entries))
		for _, entry := range chart.Entries {
			entry.AssetID = assetID
			entries = append(entries, entry)
		}
		normalized, err := NormalizeCustomAllocation(dto.CustomAllocationChart{Name: "Asset allocation", Entries: entries})
		if err != nil {
			return nil, err
		}
		result = append(result, dto.AssetChartAllocation{ChartID: chart.ChartID, Entries: normalized.Entries})
	}
	return result, nil
}
