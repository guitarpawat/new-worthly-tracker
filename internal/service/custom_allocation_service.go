package service

import (
	"context"
	"sort"
	"strings"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

type CustomAllocationStore interface {
	ListCharts(context.Context) ([]dto.CustomAllocationChart, error)
	SaveChart(context.Context, dto.CustomAllocationChart) (int64, error)
	DeleteChart(context.Context, int64) error
}

type CustomAllocationService struct {
	store     CustomAllocationStore
	snapshots ProgressReader
}

func NewCustomAllocationService(store CustomAllocationStore, snapshots ProgressReader) *CustomAllocationService {
	return &CustomAllocationService{store: store, snapshots: snapshots}
}

func (s *CustomAllocationService) ListCharts(ctx context.Context) ([]dto.CustomAllocationChart, error) {
	return s.store.ListCharts(ctx)
}

func (s *CustomAllocationService) SaveChart(ctx context.Context, input dto.CustomAllocationChart) (int64, error) {
	return s.store.SaveChart(ctx, input)
}

func (s *CustomAllocationService) DeleteChart(ctx context.Context, id int64) error {
	return s.store.DeleteChart(ctx, id)
}

func (s *CustomAllocationService) Breakdowns(
	ctx context.Context, startDate string, endDate string,
) (map[string][]dto.CustomAllocationBreakdown, error) {
	charts, err := s.store.ListCharts(ctx)
	if err != nil {
		return nil, err
	}
	result := map[string][]dto.CustomAllocationBreakdown{}
	if len(charts) == 0 {
		return result, nil
	}
	items, err := s.snapshots.ListSnapshotItemsInRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	byDate := map[string][]dto.ProgressSnapshotItem{}
	for _, item := range items {
		byDate[item.SnapshotDate] = append(byDate[item.SnapshotDate], item)
	}
	for date, rows := range byDate {
		result[date] = buildCustomAllocations(charts, rows)
	}
	return result, nil
}

func buildCustomAllocations(charts []dto.CustomAllocationChart, items []dto.ProgressSnapshotItem) []dto.CustomAllocationBreakdown {
	result := make([]dto.CustomAllocationBreakdown, 0, len(charts))
	for _, chart := range charts {
		excluded := map[int64]bool{}
		for _, id := range chart.ExcludedAssetIDs {
			excluded[id] = true
		}
		entries := map[int64][]dto.CustomAllocationEntry{}
		for _, entry := range chart.Entries {
			entries[entry.AssetID] = append(entries[entry.AssetID], entry)
		}
		totals := map[string]decimal.Decimal{}
		labels := map[string]string{}
		var unallocated decimal.Decimal
		for _, item := range items {
			if excluded[item.AssetID] {
				continue
			}
			var allocated decimal.Decimal
			for _, entry := range entries[item.AssetID] {
				key := strings.ToLower(entry.Category)
				labels[key] = entry.Category
				// Shifting two places divides by 100 exactly, without rounding small values.
				value := item.CurrentPrice.Mul(entry.Percentage).Shift(-2)
				totals[key] = totals[key].Add(value)
				allocated = allocated.Add(value)
			}
			unallocated = unallocated.Add(item.CurrentPrice.Sub(allocated))
		}
		keys := make([]string, 0, len(totals))
		for key := range totals {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		rows := make([]dto.AllocationSlice, 0, len(keys)+1)
		for _, key := range keys {
			rows = append(rows, dto.AllocationSlice{Name: labels[key], Value: totals[key]})
		}
		if !unallocated.IsZero() || len(rows) == 0 {
			rows = append(rows, dto.AllocationSlice{Name: "Unallocated", Value: unallocated})
		}
		result = append(result, dto.CustomAllocationBreakdown{ID: chart.ID, Name: chart.Name, Rows: rows})
	}
	return result
}
