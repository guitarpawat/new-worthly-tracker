package service

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

type ProgressReader interface {
	ListSnapshotDates(ctx context.Context) ([]string, error)
	ListSnapshotItemsInRange(ctx context.Context, startDate string, endDate string) ([]dto.ProgressSnapshotItem, error)
}

type GoalReader interface {
	ListGoals(ctx context.Context) ([]dto.GoalRow, error)
}

type ProgressService struct {
	repository     ProgressReader
	goalRepository GoalReader
}

func NewProgressService(repository ProgressReader, goalRepository GoalReader) *ProgressService {
	return &ProgressService{
		repository:     repository,
		goalRepository: goalRepository,
	}
}

func (s *ProgressService) GetPage(ctx context.Context, filter dto.ProgressFilter) (dto.ProgressPage, error) {
	availableDates, err := s.repository.ListSnapshotDates(ctx)
	if err != nil {
		return dto.ProgressPage{}, fmt.Errorf("list snapshot dates: %w", err)
	}

	goals, err := s.goalRepository.ListGoals(ctx)
	if err != nil {
		return dto.ProgressPage{}, fmt.Errorf("list goals: %w", err)
	}

	page := dto.ProgressPage{
		Filter:         filter,
		AvailableDates: availableDates,
		Goals:          goals,
	}
	if len(availableDates) == 0 {
		return page, nil
	}

	normalizedFilter, err := normalizeProgressFilter(availableDates, filter)
	if err != nil {
		return dto.ProgressPage{}, err
	}
	page.Filter = normalizedFilter

	rows, err := s.repository.ListSnapshotItemsInRange(ctx, normalizedFilter.StartDate, normalizedFilter.EndDate)
	if err != nil {
		return dto.ProgressPage{}, fmt.Errorf("list snapshot items in range: %w", err)
	}
	if len(rows) == 0 {
		return page, nil
	}

	page.HasData = true
	page.TrendPoints, page.AllocationSnapshots = buildProgressAggregates(rows)
	page.Summary = buildProgressSummary(page.TrendPoints)
	page.ProjectionPoints = buildProjectionPoints(page.TrendPoints, goals)
	page.GoalEstimates = buildGoalEstimates(page.TrendPoints, goals)

	return page, nil
}

func normalizeProgressFilter(availableDates []string, filter dto.ProgressFilter) (dto.ProgressFilter, error) {
	if len(availableDates) == 0 {
		return dto.ProgressFilter{}, nil
	}

	if filter.StartDate == "" && filter.EndDate == "" {
		endDate := availableDates[len(availableDates)-1]
		startIndex := len(availableDates) - 12
		if startIndex < 0 {
			startIndex = 0
		}

		return dto.ProgressFilter{
			StartDate: availableDates[startIndex],
			EndDate:   endDate,
		}, nil
	}

	return filter, nil
}

func buildProgressAggregates(rows []dto.ProgressSnapshotItem) ([]dto.ProgressPoint, []dto.AllocationSnapshot) {
	type aggregate struct {
		point       dto.ProgressPoint
		byAssetType map[string]float64
		byAsset     map[string]float64
		byCategory  map[string]float64
	}

	aggregates := make([]aggregate, 0, len(rows))
	indexByDate := make(map[string]int, len(rows))
	for _, row := range rows {
		index, found := indexByDate[row.SnapshotDate]
		if !found {
			index = len(aggregates)
			indexByDate[row.SnapshotDate] = index
			aggregates = append(aggregates, aggregate{
				point: dto.ProgressPoint{
					SnapshotDate: row.SnapshotDate,
				},
				byAssetType: make(map[string]float64),
				byAsset:     make(map[string]float64),
				byCategory:  make(map[string]float64),
			})
		}

		current := &aggregates[index]
		current.point.TotalCurrent += row.CurrentPrice
		if !row.IsCash {
			current.point.TotalBought += row.BoughtPrice
		}
		if row.IsCash {
			current.point.TotalCash += row.CurrentPrice
		} else {
			current.point.TotalNonCash += row.CurrentPrice
		}

		assetTypeName := row.AssetTypeName
		if assetTypeName == "" {
			assetTypeName = "Uncategorized"
		}
		current.byAssetType[assetTypeName] += row.CurrentPrice
		current.byAsset[row.AssetName] += row.CurrentPrice
		switch {
		case row.CurrentPrice < 0:
			current.byCategory["Liabilities"] += row.CurrentPrice
		case row.IsCash:
			current.byCategory["Cash"] += row.CurrentPrice
		default:
			current.byCategory["Non Cash Asset"] += row.CurrentPrice
		}
	}

	trendPoints := make([]dto.ProgressPoint, 0, len(aggregates))
	allocationSnapshots := make([]dto.AllocationSnapshot, 0, len(aggregates))
	for _, current := range aggregates {
		current.point.TotalProfit = current.point.TotalCurrent - current.point.TotalBought
		if current.point.TotalBought != 0 {
			current.point.ProfitRate = current.point.TotalProfit / current.point.TotalBought
		}
		if current.point.TotalCurrent != 0 {
			current.point.CashRatio = current.point.TotalCash / current.point.TotalCurrent
		}

		trendPoints = append(trendPoints, current.point)
		allocationSnapshots = append(allocationSnapshots, dto.AllocationSnapshot{
			SnapshotDate: current.point.SnapshotDate,
			ByAssetType:  mapToAllocationSlices(current.byAssetType),
			ByAsset:      mapToAllocationSlices(current.byAsset),
			ByCategory:   mapToAllocationSlices(current.byCategory),
		})
	}

	return trendPoints, allocationSnapshots
}

func mapToAllocationSlices(values map[string]float64) []dto.AllocationSlice {
	slicesOut := make([]dto.AllocationSlice, 0, len(values))
	for name, value := range values {
		slicesOut = append(slicesOut, dto.AllocationSlice{
			Name:  name,
			Value: value,
		})
	}

	slices.SortFunc(slicesOut, func(left dto.AllocationSlice, right dto.AllocationSlice) int {
		leftAbs := math.Abs(left.Value)
		rightAbs := math.Abs(right.Value)
		switch {
		case leftAbs > rightAbs:
			return -1
		case leftAbs < rightAbs:
			return 1
		case left.Name < right.Name:
			return -1
		case left.Name > right.Name:
			return 1
		default:
			return 0
		}
	})

	return slicesOut
}

func buildProgressSummary(points []dto.ProgressPoint) dto.ProgressSummary {
	if len(points) == 0 {
		return dto.ProgressSummary{}
	}

	last := points[len(points)-1]
	return dto.ProgressSummary{
		CurrentNetWorth: last.TotalCurrent,
		CurrentProfit:   last.TotalProfit,
		ProfitRate:      last.ProfitRate,
		CashRatio:       last.CashRatio,
	}
}

func buildProjectionPoints(points []dto.ProgressPoint, goals []dto.GoalRow) []dto.ProjectionPoint {
	latestDate, latestCurrent, slopePerDay, ok := calculateProjectionTrend(points)
	if !ok {
		return nil
	}

	horizonMonths := 36

	projection := make([]dto.ProjectionPoint, 0, horizonMonths+1)
	for month := 0; month <= horizonMonths; month++ {
		projectedDate := latestDate.AddDate(0, month, 0)
		daysSinceLatest := projectedDate.Sub(latestDate).Hours() / 24
		projection = append(projection, dto.ProjectionPoint{
			SnapshotDate: projectedDate.Format("2006-01-02"),
			TotalCurrent: latestCurrent + (slopePerDay * daysSinceLatest),
		})
	}

	return projection
}

func buildGoalEstimates(points []dto.ProgressPoint, goals []dto.GoalRow) []dto.GoalEstimate {
	estimates := make([]dto.GoalEstimate, 0, len(goals))
	if len(goals) == 0 || len(points) == 0 {
		return estimates
	}

	latestDate, latestCurrent, slopePerDay, projectionAvailable := calculateProjectionTrend(points)
	for _, goal := range goals {
		estimate := dto.GoalEstimate{
			GoalID:         goal.ID,
			Name:           goal.Name,
			TargetAmount:   goal.TargetAmount,
			TargetDate:     goal.TargetDate,
			RemainingValue: maxFloat(goal.TargetAmount-latestCurrent, 0),
		}

		if latestCurrent >= goal.TargetAmount {
			estimate.Status = "Reached"
			estimate.EstimatedDate = latestDate.Format("2006-01-02")
			estimates = append(estimates, estimate)
			continue
		}

		if !projectionAvailable {
			estimate.Status = "Needs positive trend"
			estimates = append(estimates, estimate)
			continue
		}

		daysNeeded := (goal.TargetAmount - latestCurrent) / slopePerDay
		if daysNeeded < 0 {
			daysNeeded = 0
		}
		estimatedDate := latestDate.Add(time.Duration(daysNeeded*24) * time.Hour)
		estimate.EstimatedDate = estimatedDate.Format("2006-01-02")
		estimate.Status = "Projected"

		if goal.TargetDate != "" {
			targetDate, err := time.Parse("2006-01-02", goal.TargetDate)
			if err == nil {
				if estimatedDate.After(targetDate) {
					estimate.Status = "Behind target"
				} else {
					estimate.Status = "On track"
				}
			}
		}

		estimates = append(estimates, estimate)
	}

	return estimates
}

func calculateProjectionTrend(points []dto.ProgressPoint) (time.Time, float64, float64, bool) {
	if len(points) < 3 {
		return time.Time{}, 0, 0, false
	}

	firstDate, err := time.Parse("2006-01-02", points[0].SnapshotDate)
	if err != nil {
		return time.Time{}, 0, 0, false
	}
	lastDate, err := time.Parse("2006-01-02", points[len(points)-1].SnapshotDate)
	if err != nil {
		return time.Time{}, 0, 0, false
	}
	days := lastDate.Sub(firstDate).Hours() / 24
	if days <= 0 {
		return time.Time{}, 0, 0, false
	}

	firstCurrent := points[0].TotalCurrent
	lastCurrent := points[len(points)-1].TotalCurrent
	slopePerDay := (lastCurrent - firstCurrent) / days
	if slopePerDay <= 0 {
		return lastDate, lastCurrent, 0, false
	}

	return lastDate, lastCurrent, slopePerDay, true
}

func maxFloat(left float64, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
