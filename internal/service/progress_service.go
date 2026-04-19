package service

import (
	"context"
	"fmt"
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
		byAssetType []dto.AllocationSlice
		byAsset     []dto.AllocationSlice
		byCategory  []dto.AllocationSlice
		typeIndex   map[string]int
		assetIndex  map[string]int
		categoryMap map[string]int
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
				typeIndex:   make(map[string]int),
				assetIndex:  make(map[string]int),
				categoryMap: make(map[string]int),
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
		accumulateAllocationSlice(&current.byAssetType, current.typeIndex, assetTypeName, row.CurrentPrice)
		accumulateAllocationSlice(&current.byAsset, current.assetIndex, row.AssetName, row.CurrentPrice)
		switch {
		case row.IsLiability:
			accumulateAllocationSlice(&current.byCategory, current.categoryMap, "Liabilities", row.CurrentPrice)
		case row.IsCash:
			accumulateAllocationSlice(&current.byCategory, current.categoryMap, "Cash", row.CurrentPrice)
		default:
			accumulateAllocationSlice(&current.byCategory, current.categoryMap, "Non Cash Asset", row.CurrentPrice)
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
			ByAssetType:  current.byAssetType,
			ByAsset:      current.byAsset,
			ByCategory:   normalizeCategorySlices(current.byCategory),
		})
	}

	return trendPoints, allocationSnapshots
}

func accumulateAllocationSlice(target *[]dto.AllocationSlice, indexMap map[string]int, name string, value float64) {
	index, found := indexMap[name]
	if !found {
		index = len(*target)
		indexMap[name] = index
		*target = append(*target, dto.AllocationSlice{Name: name, Value: value})
		return
	}
	(*target)[index].Value += value
}

func normalizeCategorySlices(rows []dto.AllocationSlice) []dto.AllocationSlice {
	order := []string{"Cash", "Non Cash Asset", "Liabilities"}
	indexMap := make(map[string]int, len(rows))
	for index, row := range rows {
		indexMap[row.Name] = index
	}

	ordered := make([]dto.AllocationSlice, 0, len(rows))
	for _, name := range order {
		index, found := indexMap[name]
		if !found {
			continue
		}
		ordered = append(ordered, rows[index])
	}
	for _, row := range rows {
		if row.Name == "Cash" || row.Name == "Non Cash Asset" || row.Name == "Liabilities" {
			continue
		}
		ordered = append(ordered, row)
	}
	return ordered
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
