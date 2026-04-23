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
	page.ProjectionPoints = buildProjectionPoints(page.AllocationSnapshots)
	page.GoalEstimates = buildGoalEstimates(page.TrendPoints, page.AllocationSnapshots, goals)

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

func buildProjectionPoints(allocationSnapshots []dto.AllocationSnapshot) []dto.ProjectionPoint {
	model, ok := buildHybridProjectionModel(allocationSnapshots)
	if !ok || !model.hasPositiveNetWorthTrend(36) {
		return nil
	}

	horizonMonths := 36

	projection := make([]dto.ProjectionPoint, 0, horizonMonths+1)
	for month := 0; month <= horizonMonths; month++ {
		projectedDate := model.latestDate.AddDate(0, month, 0)
		projectedCash := model.projectCash(float64(month))
		projectedNonCash := model.projectNonCash(float64(month))
		projectedLiabilities := model.projectLiabilities(float64(month))
		projection = append(projection, dto.ProjectionPoint{
			SnapshotDate: projectedDate.Format("2006-01-02"),
			TotalCurrent: projectedCash + projectedNonCash + projectedLiabilities,
			TotalCash:    projectedCash,
			TotalNonCash: projectedNonCash,
			Liabilities:  projectedLiabilities,
		})
	}

	return projection
}

func buildGoalEstimates(
	points []dto.ProgressPoint,
	allocationSnapshots []dto.AllocationSnapshot,
	goals []dto.GoalRow,
) []dto.GoalEstimate {
	estimates := make([]dto.GoalEstimate, 0, len(goals))
	if len(goals) == 0 || len(points) == 0 {
		return estimates
	}

	latestSnapshotDate := points[len(points)-1].SnapshotDate
	latestCurrent := points[len(points)-1].TotalCurrent
	model, modelAvailable := buildHybridProjectionModel(allocationSnapshots)
	if modelAvailable {
		latestSnapshotDate = model.latestDate.Format("2006-01-02")
	}

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
			estimate.EstimatedDate = latestSnapshotDate
			estimates = append(estimates, estimate)
			continue
		}

		if !modelAvailable {
			estimate.Status = "Needs positive trend"
			estimates = append(estimates, estimate)
			continue
		}

		estimatedDate, ok := model.findTargetDate(goal.TargetAmount, 600)
		if !ok {
			estimate.Status = "Needs positive trend"
			estimates = append(estimates, estimate)
			continue
		}

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

type projectionCategorySnapshot struct {
	snapshotDate time.Time
	cash         float64
	nonCash      float64
	liabilities  float64
}

type hybridProjectionModel struct {
	latestDate                time.Time
	latestCash                float64
	latestNonCash             float64
	latestLiabilities         float64
	cashDeltaPerMonth         float64
	liabilityDeltaPerMonth    float64
	nonCashGrowthRatePerMonth float64
}

func buildHybridProjectionModel(
	allocationSnapshots []dto.AllocationSnapshot,
) (hybridProjectionModel, bool) {
	if len(allocationSnapshots) < 3 {
		return hybridProjectionModel{}, false
	}

	categorySnapshots, ok := buildProjectionCategorySnapshots(allocationSnapshots)
	if !ok || len(categorySnapshots) < 3 {
		return hybridProjectionModel{}, false
	}

	cashDeltas := make([]float64, 0, len(categorySnapshots)-1)
	liabilityDeltas := make([]float64, 0, len(categorySnapshots)-1)
	nonCashGrowthRates := make([]float64, 0, len(categorySnapshots)-1)
	for index := 1; index < len(categorySnapshots); index++ {
		previous := categorySnapshots[index-1]
		current := categorySnapshots[index]
		monthsBetween := diffSnapshotMonths(previous.snapshotDate, current.snapshotDate)
		if monthsBetween <= 0 {
			continue
		}

		cashDeltas = append(
			cashDeltas,
			(current.cash-previous.cash)/monthsBetween,
		)
		liabilityDeltas = append(
			liabilityDeltas,
			(current.liabilities-previous.liabilities)/monthsBetween,
		)

		if previous.nonCash <= 0 || current.nonCash <= 0 {
			continue
		}

		nonCashGrowthRates = append(
			nonCashGrowthRates,
			math.Pow(current.nonCash/previous.nonCash, 1/monthsBetween)-1,
		)
	}

	if len(cashDeltas) == 0 && len(liabilityDeltas) == 0 && len(nonCashGrowthRates) == 0 {
		return hybridProjectionModel{}, false
	}

	latest := categorySnapshots[len(categorySnapshots)-1]
	return hybridProjectionModel{
		latestDate:                latest.snapshotDate,
		latestCash:                latest.cash,
		latestNonCash:             latest.nonCash,
		latestLiabilities:         latest.liabilities,
		cashDeltaPerMonth:         averageFloat(cashDeltas),
		liabilityDeltaPerMonth:    medianFloat(liabilityDeltas),
		nonCashGrowthRatePerMonth: averageFloat(nonCashGrowthRates),
	}, true
}

func buildProjectionCategorySnapshots(
	allocationSnapshots []dto.AllocationSnapshot,
) ([]projectionCategorySnapshot, bool) {
	categorySnapshots := make([]projectionCategorySnapshot, 0, len(allocationSnapshots))
	for _, snapshot := range allocationSnapshots {
		snapshotDate, err := time.Parse("2006-01-02", snapshot.SnapshotDate)
		if err != nil {
			return nil, false
		}

		categorySnapshots = append(categorySnapshots, projectionCategorySnapshot{
			snapshotDate: snapshotDate,
			cash:         resolveAllocationValue(snapshot.ByCategory, "Cash"),
			nonCash:      resolveAllocationValue(snapshot.ByCategory, "Non Cash Asset"),
			liabilities:  resolveAllocationValue(snapshot.ByCategory, "Liabilities"),
		})
	}
	return categorySnapshots, true
}

func resolveAllocationValue(rows []dto.AllocationSlice, targetName string) float64 {
	for _, row := range rows {
		if row.Name == targetName {
			return row.Value
		}
	}
	return 0
}

func diffSnapshotMonths(startDate time.Time, endDate time.Time) float64 {
	const averageDaysPerMonth = 365.25 / 12

	return endDate.Sub(startDate).Hours() / 24 / averageDaysPerMonth
}

func averageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	total := 0.0
	for _, value := range values {
		total += value
	}

	return total / float64(len(values))
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64{}, values...)
	slices.Sort(sorted)

	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}

	return (sorted[middle-1] + sorted[middle]) / 2
}

func (m hybridProjectionModel) projectNetWorth(monthsFromLatest float64) float64 {
	return m.projectCash(monthsFromLatest) +
		m.projectNonCash(monthsFromLatest) +
		m.projectLiabilities(monthsFromLatest)
}

func (m hybridProjectionModel) projectCash(monthsFromLatest float64) float64 {
	return m.latestCash + (m.cashDeltaPerMonth * monthsFromLatest)
}

func (m hybridProjectionModel) projectNonCash(monthsFromLatest float64) float64 {
	projectedNonCash := m.latestNonCash
	if projectedNonCash <= 0 {
		return projectedNonCash
	}

	return projectedNonCash * math.Pow(1+m.nonCashGrowthRatePerMonth, monthsFromLatest)
}

func (m hybridProjectionModel) projectLiabilities(monthsFromLatest float64) float64 {
	projectedLiabilities := m.latestLiabilities + (m.liabilityDeltaPerMonth * monthsFromLatest)
	if projectedLiabilities > 0 {
		return 0
	}

	return projectedLiabilities
}

func (m hybridProjectionModel) hasPositiveNetWorthTrend(horizonMonths int) bool {
	latestNetWorth := m.projectNetWorth(0)
	for month := 1; month <= horizonMonths; month++ {
		if m.projectNetWorth(float64(month)) > latestNetWorth {
			return true
		}
	}

	return false
}

func (m hybridProjectionModel) findTargetDate(
	targetAmount float64,
	maxMonths int,
) (time.Time, bool) {
	previousDate := m.latestDate
	previousNetWorth := m.projectNetWorth(0)
	if previousNetWorth >= targetAmount {
		return previousDate, true
	}

	for month := 1; month <= maxMonths; month++ {
		projectedDate := m.latestDate.AddDate(0, month, 0)
		projectedNetWorth := m.projectNetWorth(float64(month))
		if projectedNetWorth >= targetAmount {
			return interpolateProjectionDate(
				previousDate,
				previousNetWorth,
				projectedDate,
				projectedNetWorth,
				targetAmount,
			), true
		}

		previousDate = projectedDate
		previousNetWorth = projectedNetWorth
	}

	return time.Time{}, false
}

func interpolateProjectionDate(
	startDate time.Time,
	startValue float64,
	endDate time.Time,
	endValue float64,
	targetValue float64,
) time.Time {
	growth := endValue - startValue
	if growth <= 0 {
		return endDate
	}

	progress := (targetValue - startValue) / growth
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	return startDate.Add(time.Duration(float64(endDate.Sub(startDate)) * progress))
}

func maxFloat(left float64, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
