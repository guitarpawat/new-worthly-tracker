package service

import (
	"context"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

type progressReaderStub struct {
	dates []string
	rows  []dto.ProgressSnapshotItem
}

func (s progressReaderStub) ListSnapshotDates(context.Context) ([]string, error) {
	return s.dates, nil
}

func (s progressReaderStub) ListSnapshotItemsInRange(context.Context, string, string) ([]dto.ProgressSnapshotItem, error) {
	return s.rows, nil
}

type goalReaderStub struct {
	goals []dto.GoalRow
}

func (s goalReaderStub) ListGoals(context.Context) ([]dto.GoalRow, error) {
	return s.goals, nil
}

func TestProgressService_GetPageUsesLatestTwelveDatesAsDefaultRange(t *testing.T) {
	t.Parallel()

	service := NewProgressService(progressReaderStub{
		dates: []string{
			"2025-01-12", "2025-02-12", "2025-03-12", "2025-04-12",
			"2025-05-12", "2025-06-12", "2025-07-12", "2025-08-12",
			"2025-09-12", "2025-10-12", "2025-11-12", "2025-12-12",
			"2026-01-12",
		},
		rows: []dto.ProgressSnapshotItem{
			{SnapshotDate: "2026-01-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 1000, BoughtPrice: 900},
		},
	}, goalReaderStub{})

	page, err := service.GetPage(context.Background(), dto.ProgressFilter{})
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	if page.Filter.StartDate != "2025-02-12" || page.Filter.EndDate != "2026-01-12" {
		t.Fatalf("unexpected default range: %+v", page.Filter)
	}
}

func TestProgressService_GetPageBuildsTrendSummaryProjectionAllocationAndGoalEstimates(t *testing.T) {
	t.Parallel()

	service := NewProgressService(progressReaderStub{
		dates: []string{"2026-01-12", "2026-02-12", "2026-03-12"},
		rows: []dto.ProgressSnapshotItem{
			{SnapshotDate: "2026-01-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 10000, BoughtPrice: 9000},
			{SnapshotDate: "2026-01-12", AssetName: "Wallet", AssetTypeName: "Cash", CurrentPrice: 2000, BoughtPrice: 500, IsCash: true},
			{SnapshotDate: "2026-01-12", AssetName: "Visa", AssetTypeName: "Credit Card", CurrentPrice: -1000, BoughtPrice: 0, IsCash: true},
			{SnapshotDate: "2026-02-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 11000, BoughtPrice: 9500},
			{SnapshotDate: "2026-02-12", AssetName: "Wallet", AssetTypeName: "Cash", CurrentPrice: 2400, BoughtPrice: 500, IsCash: true},
			{SnapshotDate: "2026-02-12", AssetName: "Visa", AssetTypeName: "Credit Card", CurrentPrice: -900, BoughtPrice: 0, IsCash: true},
			{SnapshotDate: "2026-03-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 12500, BoughtPrice: 10000},
			{SnapshotDate: "2026-03-12", AssetName: "Wallet", AssetTypeName: "Cash", CurrentPrice: 2600, BoughtPrice: 500, IsCash: true},
			{SnapshotDate: "2026-03-12", AssetName: "Visa", AssetTypeName: "Credit Card", CurrentPrice: -800, BoughtPrice: 0, IsCash: true},
		},
	}, goalReaderStub{
		goals: []dto.GoalRow{
			{ID: 1, Name: "First Million", TargetAmount: 20000},
			{ID: 2, Name: "Trip Fund", TargetAmount: 10000, TargetDate: "2026-12-31"},
		},
	})

	page, err := service.GetPage(context.Background(), dto.ProgressFilter{
		StartDate: "2026-01-12",
		EndDate:   "2026-03-12",
	})
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	if !page.HasData {
		t.Fatal("expected data")
	}
	if len(page.TrendPoints) != 3 {
		t.Fatalf("expected 3 trend points, got %d", len(page.TrendPoints))
	}
	assertProgressFloatEqual(t, 14300, page.Summary.CurrentNetWorth)
	assertProgressFloatEqual(t, 4300, page.Summary.CurrentProfit)
	assertProgressFloatInDelta(t, 0.4300, page.Summary.ProfitRate)
	assertProgressFloatInDelta(t, 0.1258741258, page.Summary.CashRatio)

	if len(page.ProjectionPoints) < 2 {
		t.Fatalf("expected projection points, got %d", len(page.ProjectionPoints))
	}
	if len(page.AllocationSnapshots) != 3 {
		t.Fatalf("expected 3 allocation snapshots, got %d", len(page.AllocationSnapshots))
	}
	if page.AllocationSnapshots[2].ByAssetType[0].Name != "Investment" {
		t.Fatalf("expected investment to be largest allocation, got %+v", page.AllocationSnapshots[2].ByAssetType)
	}
	if page.AllocationSnapshots[2].ByCategory[2].Name != "Liabilities" || page.AllocationSnapshots[2].ByCategory[2].Value >= 0 {
		t.Fatalf("expected liabilities category to stay negative, got %+v", page.AllocationSnapshots[2].ByCategory)
	}
	if len(page.GoalEstimates) != 2 {
		t.Fatalf("expected 2 goal estimates, got %d", len(page.GoalEstimates))
	}
	if page.GoalEstimates[0].Status != "Projected" && page.GoalEstimates[0].Status != "On track" {
		t.Fatalf("expected projected status, got %+v", page.GoalEstimates[0])
	}
}

func TestProgressService_GetPageMarksGoalsAsReachedOrNeedsPositiveTrend(t *testing.T) {
	t.Parallel()

	service := NewProgressService(progressReaderStub{
		dates: []string{"2026-01-12", "2026-02-12", "2026-03-12"},
		rows: []dto.ProgressSnapshotItem{
			{SnapshotDate: "2026-01-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 10000, BoughtPrice: 9000},
			{SnapshotDate: "2026-02-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 9500, BoughtPrice: 9000},
			{SnapshotDate: "2026-03-12", AssetName: "ETF", AssetTypeName: "Investment", CurrentPrice: 9000, BoughtPrice: 9000},
		},
	}, goalReaderStub{
		goals: []dto.GoalRow{
			{ID: 1, Name: "Already There", TargetAmount: 8000},
			{ID: 2, Name: "Too Far", TargetAmount: 12000},
		},
	})

	page, err := service.GetPage(context.Background(), dto.ProgressFilter{
		StartDate: "2026-01-12",
		EndDate:   "2026-03-12",
	})
	if err != nil {
		t.Fatalf("GetPage returned error: %v", err)
	}

	if page.GoalEstimates[0].Status != "Reached" {
		t.Fatalf("expected reached status, got %+v", page.GoalEstimates[0])
	}
	if page.GoalEstimates[1].Status != "Needs positive trend" {
		t.Fatalf("expected needs positive trend status, got %+v", page.GoalEstimates[1])
	}
	if len(page.ProjectionPoints) != 0 {
		t.Fatalf("expected no projection points, got %d", len(page.ProjectionPoints))
	}
}

func assertProgressFloatEqual(t *testing.T, want float64, got float64) {
	t.Helper()
	if want != got {
		t.Fatalf("expected %f, got %f", want, got)
	}
}

func assertProgressFloatInDelta(t *testing.T, want float64, got float64) {
	t.Helper()
	const delta = 0.00001
	if got < want-delta || got > want+delta {
		t.Fatalf("expected %f +/- %f, got %f", want, delta, got)
	}
}
