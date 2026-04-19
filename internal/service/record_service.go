package service

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type SnapshotReader interface {
	GetSnapshotByOffset(ctx context.Context, offset int) (*dto.Snapshot, error)
	ListSnapshotOptions(ctx context.Context) ([]dto.SnapshotOption, error)
	GetEditableSnapshotByOffset(ctx context.Context, offset int) (*dto.EditableSnapshot, error)
	GetNewSnapshotDraft(ctx context.Context, recordDate time.Time) (*dto.EditableSnapshot, error)
	SaveSnapshot(ctx context.Context, input dto.SaveSnapshotInput) (dto.SaveSnapshotResult, error)
	CreateSnapshot(ctx context.Context, input dto.CreateSnapshotInput) (dto.CreateSnapshotResult, error)
	DeleteSnapshot(ctx context.Context, input dto.DeleteSnapshotInput) (dto.DeleteSnapshotResult, error)
}

type RecordService struct {
	repository SnapshotReader
}

func NewRecordService(repository SnapshotReader) *RecordService {
	return &RecordService{repository: repository}
}

func (s *RecordService) GetHomePage(
	ctx context.Context,
	offset int,
) (dto.HomePage, error) {
	if offset < 0 {
		return dto.HomePage{}, fmt.Errorf("offset must be non-negative")
	}

	current, err := s.repository.GetSnapshotByOffset(ctx, offset)
	if err != nil {
		return dto.HomePage{}, fmt.Errorf("get current snapshot: %w", err)
	}
	if current == nil {
		return dto.HomePage{}, nil
	}

	previous, err := s.repository.GetSnapshotByOffset(ctx, offset+1)
	if err != nil {
		return dto.HomePage{}, fmt.Errorf("get previous snapshot: %w", err)
	}

	page := s.BuildHomePage(current, previous)
	options, err := s.repository.ListSnapshotOptions(ctx)
	if err != nil {
		return dto.HomePage{}, fmt.Errorf("list snapshot options: %w", err)
	}
	page.SnapshotOptions = options
	page.CanNavigateForward = offset > 0
	page.CanNavigateBack = previous != nil

	return page, nil
}

func (s *RecordService) GetEditSnapshotPage(
	ctx context.Context,
	offset int,
) (dto.EditSnapshotPage, error) {
	if offset < 0 {
		return dto.EditSnapshotPage{}, fmt.Errorf("offset must be non-negative")
	}

	snapshot, err := s.repository.GetEditableSnapshotByOffset(ctx, offset)
	if err != nil {
		return dto.EditSnapshotPage{}, fmt.Errorf("get editable snapshot: %w", err)
	}
	if snapshot == nil {
		return dto.EditSnapshotPage{}, recorderr.ErrSnapshotNotFound
	}

	return buildSnapshotFormPage("edit", snapshot), nil
}

func (s *RecordService) GetNewSnapshotPage(
	ctx context.Context,
	recordDate time.Time,
) (dto.EditSnapshotPage, error) {
	snapshot, err := s.repository.GetNewSnapshotDraft(ctx, recordDate)
	if err != nil {
		return dto.EditSnapshotPage{}, fmt.Errorf("get new snapshot draft: %w", err)
	}

	return buildSnapshotFormPage("new", snapshot), nil
}

func (s *RecordService) SaveSnapshot(
	ctx context.Context,
	input dto.SaveSnapshotInput,
) (dto.SaveSnapshotResult, error) {
	if len(input.Items) == 0 {
		return dto.SaveSnapshotResult{}, recorderr.ErrSnapshotHasNoRows
	}

	result, err := s.repository.SaveSnapshot(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrSnapshotDateAlreadyExists):
			return dto.SaveSnapshotResult{}, recorderr.ErrSnapshotDateAlreadyExists
		case errors.Is(err, recorderr.ErrSnapshotNotFound):
			return dto.SaveSnapshotResult{}, recorderr.ErrSnapshotNotFound
		case errors.Is(err, recorderr.ErrAssetUnavailable):
			return dto.SaveSnapshotResult{}, recorderr.ErrAssetUnavailable
		default:
			return dto.SaveSnapshotResult{}, fmt.Errorf("save snapshot: %w", err)
		}
	}

	return result, nil
}

func (s *RecordService) CreateSnapshot(
	ctx context.Context,
	input dto.CreateSnapshotInput,
) (dto.CreateSnapshotResult, error) {
	if len(input.Items) == 0 {
		return dto.CreateSnapshotResult{}, recorderr.ErrSnapshotHasNoRows
	}

	result, err := s.repository.CreateSnapshot(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrSnapshotDateAlreadyExists):
			return dto.CreateSnapshotResult{}, recorderr.ErrSnapshotDateAlreadyExists
		case errors.Is(err, recorderr.ErrAssetUnavailable):
			return dto.CreateSnapshotResult{}, recorderr.ErrAssetUnavailable
		default:
			return dto.CreateSnapshotResult{}, fmt.Errorf("create snapshot: %w", err)
		}
	}

	return result, nil
}

func (s *RecordService) DeleteSnapshot(
	ctx context.Context,
	input dto.DeleteSnapshotInput,
) (dto.DeleteSnapshotResult, error) {
	if input.SnapshotID <= 0 {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("snapshot id must be positive")
	}
	if input.Offset < 0 {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("offset must be non-negative")
	}

	result, err := s.repository.DeleteSnapshot(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrSnapshotNotFound):
			return dto.DeleteSnapshotResult{}, recorderr.ErrSnapshotNotFound
		default:
			return dto.DeleteSnapshotResult{}, fmt.Errorf("delete snapshot: %w", err)
		}
	}

	return result, nil
}

func (*RecordService) BuildHomePage(
	current *dto.Snapshot,
	previous *dto.Snapshot,
) dto.HomePage {
	if current == nil || len(current.Items) == 0 {
		return dto.HomePage{}
	}

	page := dto.HomePage{
		SnapshotID:         current.ID,
		HasSnapshot:        true,
		SnapshotDate:       current.RecordDate,
		Groups:             buildHomeGroups(current.Items),
		Summary:            calculateHomeSummary(current.Items),
		CanNavigateBack:    previous != nil,
		CanNavigateForward: false,
	}

	if previous == nil || len(previous.Items) == 0 {
		return page
	}

	page.PreviousSnapshot = &previous.RecordDate

	currentOverlap, previousOverlap := overlappingItems(current.Items, previous.Items)
	if len(currentOverlap) == 0 || len(previousOverlap) == 0 {
		return page
	}

	currentSummary := calculateHomeSummary(currentOverlap)
	previousSummary := calculateHomeSummary(previousOverlap)
	page.Comparison = &dto.HomeSummaryDelta{
		PreviousSnapshotDate: previous.RecordDate,
		BoughtChange:         currentSummary.TotalBought - previousSummary.TotalBought,
		CurrentChange:        currentSummary.TotalCurrent - previousSummary.TotalCurrent,
		ProfitChange:         currentSummary.TotalProfit - previousSummary.TotalProfit,
		ProfitRateChange:     currentSummary.TotalProfitRate - previousSummary.TotalProfitRate,
		CashChange:           currentSummary.TotalCash - previousSummary.TotalCash,
		NonCashChange:        currentSummary.TotalNonCash - previousSummary.TotalNonCash,
		CashRatioChange:      currentSummary.CashRatio - previousSummary.CashRatio,
	}

	return page
}

func buildHomeGroups(items []dto.SnapshotItem) []dto.HomeAssetGroup {
	sortedItems := slices.Clone(items)
	slices.SortFunc(sortedItems, func(left dto.SnapshotItem, right dto.SnapshotItem) int {
		if diff := cmp.Compare(left.AssetTypeOrdering, right.AssetTypeOrdering); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(left.AssetTypeName, right.AssetTypeName); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(left.AssetOrdering, right.AssetOrdering); diff != 0 {
			return diff
		}
		return cmp.Compare(left.AssetName, right.AssetName)
	})

	groups := make([]dto.HomeAssetGroup, 0, len(sortedItems))
	for _, item := range sortedItems {
		row := dto.HomeAssetRow{
			AssetID:          item.AssetID,
			AssetName:        item.AssetName,
			Broker:           item.Broker,
			BoughtPrice:      item.BoughtPrice,
			CurrentPrice:     item.CurrentPrice,
			Profit:           calculateProfit(item),
			ProfitPercent:    calculateProfitPercent(item),
			ProfitApplicable: !item.IsCash,
			Remarks:          item.Remarks,
		}

		lastGroupIndex := len(groups) - 1
		if len(groups) == 0 || groups[lastGroupIndex].AssetTypeName != item.AssetTypeName {
			groups = append(groups, dto.HomeAssetGroup{
				AssetTypeName: item.AssetTypeName,
				Summary:       dto.HomeAssetGroupSummary{},
				Rows:          []dto.HomeAssetRow{row},
			})
			updateHomeGroupSummary(&groups[len(groups)-1].Summary, item)
			continue
		}

		groups[lastGroupIndex].Rows = append(groups[lastGroupIndex].Rows, row)
		updateHomeGroupSummary(&groups[lastGroupIndex].Summary, item)
	}

	return groups
}

func updateHomeGroupSummary(summary *dto.HomeAssetGroupSummary, item dto.SnapshotItem) {
	summary.AssetCount += 1
	summary.TotalCurrent += item.CurrentPrice
}

func buildSnapshotFormPage(mode string, snapshot *dto.EditableSnapshot) dto.EditSnapshotPage {
	page := dto.EditSnapshotPage{
		Mode:         mode,
		SnapshotID:   snapshot.ID,
		SnapshotDate: snapshot.RecordDate.Format("2006-01-02"),
		Groups:       buildEditGroups(snapshot.Items),
	}
	page.AvailableAssetGroups = buildAvailableAssetGroups(snapshot.AvailableAssets)

	return page
}

func buildEditGroups(items []dto.EditableSnapshotItem) []dto.EditSnapshotGroup {
	sortedItems := slices.Clone(items)
	slices.SortFunc(sortedItems, func(left dto.EditableSnapshotItem, right dto.EditableSnapshotItem) int {
		if diff := cmp.Compare(left.AssetTypeOrdering, right.AssetTypeOrdering); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(left.AssetTypeName, right.AssetTypeName); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(left.AssetOrdering, right.AssetOrdering); diff != 0 {
			return diff
		}
		return cmp.Compare(left.AssetName, right.AssetName)
	})

	groups := make([]dto.EditSnapshotGroup, 0, len(sortedItems))
	for _, item := range sortedItems {
		row := dto.EditSnapshotRow{
			AssetTypeOrdering: item.AssetTypeOrdering,
			AssetOrdering:     item.AssetOrdering,
			AssetID:           item.AssetID,
			AssetName:         item.AssetName,
			Broker:            item.Broker,
			IsCash:            item.IsCash,
			IsActive:          item.IsActive,
			BoughtPrice:       item.BoughtPrice,
			CurrentPrice:      item.CurrentPrice,
			Remarks:           item.Remarks,
		}

		lastGroupIndex := len(groups) - 1
		if len(groups) == 0 || groups[lastGroupIndex].AssetTypeName != item.AssetTypeName {
			groups = append(groups, dto.EditSnapshotGroup{
				AssetTypeName: item.AssetTypeName,
				Rows:          []dto.EditSnapshotRow{row},
			})
			continue
		}

		groups[lastGroupIndex].Rows = append(groups[lastGroupIndex].Rows, row)
	}

	return groups
}

func buildAvailableAssetGroups(items []dto.EditableAssetOption) []dto.EditAssetOptionGroup {
	sortedItems := slices.Clone(items)
	slices.SortFunc(sortedItems, func(left dto.EditableAssetOption, right dto.EditableAssetOption) int {
		if diff := cmp.Compare(left.AssetTypeOrdering, right.AssetTypeOrdering); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(left.AssetTypeName, right.AssetTypeName); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(left.AssetOrdering, right.AssetOrdering); diff != 0 {
			return diff
		}
		return cmp.Compare(left.AssetName, right.AssetName)
	})

	groups := make([]dto.EditAssetOptionGroup, 0, len(sortedItems))
	for _, item := range sortedItems {
		option := dto.EditAssetOption{
			AssetTypeOrdering: item.AssetTypeOrdering,
			AssetOrdering:     item.AssetOrdering,
			AssetID:           item.AssetID,
			AssetName:         item.AssetName,
			Broker:            item.Broker,
			IsCash:            item.IsCash,
			IsActive:          item.IsActive,
		}

		lastGroupIndex := len(groups) - 1
		if len(groups) == 0 || groups[lastGroupIndex].AssetTypeName != item.AssetTypeName {
			groups = append(groups, dto.EditAssetOptionGroup{
				AssetTypeName: item.AssetTypeName,
				Options:       []dto.EditAssetOption{option},
			})
			continue
		}

		groups[lastGroupIndex].Options = append(groups[lastGroupIndex].Options, option)
	}

	return groups
}

func calculateHomeSummary(items []dto.SnapshotItem) dto.HomeSummary {
	summary := dto.HomeSummary{}
	var investedBought float64

	for _, item := range items {
		summary.TotalCurrent += item.CurrentPrice

		if item.IsCash {
			summary.TotalCash += item.CurrentPrice
			continue
		}

		summary.TotalBought += item.BoughtPrice
		summary.TotalNonCash += item.CurrentPrice
		summary.TotalProfit += calculateProfit(item)
		investedBought += item.BoughtPrice
	}

	summary.TotalProfitRate = safeDivide(summary.TotalProfit, investedBought)
	summary.CashRatio = safeDivide(summary.TotalCash, summary.TotalCurrent)

	return summary
}

func overlappingItems(
	current []dto.SnapshotItem,
	previous []dto.SnapshotItem,
) ([]dto.SnapshotItem, []dto.SnapshotItem) {
	previousByAssetID := make(map[int64]dto.SnapshotItem, len(previous))
	for _, item := range previous {
		previousByAssetID[item.AssetID] = item
	}

	currentOverlap := make([]dto.SnapshotItem, 0, len(current))
	previousOverlap := make([]dto.SnapshotItem, 0, len(current))
	for _, item := range current {
		previousItem, found := previousByAssetID[item.AssetID]
		if !found {
			continue
		}

		currentOverlap = append(currentOverlap, item)
		previousOverlap = append(previousOverlap, previousItem)
	}

	return currentOverlap, previousOverlap
}

func calculateProfit(item dto.SnapshotItem) float64 {
	if item.IsCash {
		return 0
	}

	return item.CurrentPrice - item.BoughtPrice
}

func calculateProfitPercent(item dto.SnapshotItem) float64 {
	if item.IsCash {
		return 0
	}

	return safeDivide(calculateProfit(item), item.BoughtPrice)
}

func safeDivide(numerator float64, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
