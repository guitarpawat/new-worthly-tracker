package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

func TestRecordService_BuildHomePageReturnsOnboardingWhenCurrentSnapshotIsMissing(t *testing.T) {
	t.Parallel()

	service := RecordService{}

	testCases := []struct {
		name    string
		current *dto.Snapshot
	}{
		{
			name:    "nil current snapshot",
			current: nil,
		},
		{
			name: "current snapshot without items",
			current: &dto.Snapshot{
				RecordDate: time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
				Items:      []dto.SnapshotItem{},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			page := service.BuildHomePage(testCase.current, nil)

			if page.HasSnapshot {
				t.Fatal("expected onboarding state")
			}
			if len(page.Groups) != 0 {
				t.Fatalf("expected no groups, got %d", len(page.Groups))
			}
			if page.Comparison != nil {
				t.Fatal("expected no comparison")
			}
		})
	}
}

func TestRecordService_BuildHomePageCalculatesGroupedSummaryAndProfitRules(t *testing.T) {
	t.Parallel()

	service := RecordService{}
	current := &dto.Snapshot{
		RecordDate: time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
		Items: []dto.SnapshotItem{
			{
				AssetID:           3,
				AssetName:         "Emergency Fund",
				AssetTypeName:     "Cash",
				AssetTypeOrdering: 1,
				AssetOrdering:     2,
				Broker:            "SCB",
				IsCash:            true,
				BoughtPrice:       0,
				CurrentPrice:      11000,
				Remarks:           "Payroll",
			},
			{
				AssetID:           1,
				AssetName:         "SET50 ETF",
				AssetTypeName:     "Investment",
				AssetTypeOrdering: 2,
				AssetOrdering:     2,
				Broker:            "KKP",
				BoughtPrice:       15000,
				CurrentPrice:      18000,
				Remarks:           "Long term",
			},
			{
				AssetID:           2,
				AssetName:         "US Index Fund",
				AssetTypeName:     "Investment",
				AssetTypeOrdering: 2,
				AssetOrdering:     1,
				Broker:            "IBKR",
				BoughtPrice:       0,
				CurrentPrice:      5000,
				Remarks:           "Bonus buy",
			},
		},
	}

	page := service.BuildHomePage(current, nil)

	if !page.HasSnapshot {
		t.Fatal("expected snapshot state")
	}
	if len(page.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(page.Groups))
	}
	if page.Groups[0].AssetTypeName != "Cash" {
		t.Fatalf("expected first group to be Cash, got %q", page.Groups[0].AssetTypeName)
	}
	if page.Groups[1].AssetTypeName != "Investment" {
		t.Fatalf("expected second group to be Investment, got %q", page.Groups[1].AssetTypeName)
	}
	if page.Groups[1].Rows[0].AssetName != "US Index Fund" {
		t.Fatalf("expected first investment row to be ordered by asset ordering, got %q", page.Groups[1].Rows[0].AssetName)
	}
	if page.Groups[1].Rows[1].AssetName != "SET50 ETF" {
		t.Fatalf("expected second investment row to be SET50 ETF, got %q", page.Groups[1].Rows[1].AssetName)
	}
	if page.Groups[0].Summary.AssetCount != 1 {
		t.Fatalf("expected cash group asset count 1, got %d", page.Groups[0].Summary.AssetCount)
	}
	if page.Groups[1].Summary.AssetCount != 2 {
		t.Fatalf("expected investment group asset count 2, got %d", page.Groups[1].Summary.AssetCount)
	}
	assertFloatEqual(t, 11000, page.Groups[0].Summary.TotalCurrent)
	assertFloatEqual(t, 23000, page.Groups[1].Summary.TotalCurrent)

	cashRow := page.Groups[0].Rows[0]
	if cashRow.ProfitApplicable {
		t.Fatal("expected cash row to disable profit")
	}
	if cashRow.Profit != 0 {
		t.Fatalf("expected zero cash profit, got %f", cashRow.Profit)
	}
	if cashRow.ProfitPercent != 0 {
		t.Fatalf("expected zero cash profit percent, got %f", cashRow.ProfitPercent)
	}

	investmentRow := page.Groups[1].Rows[0]
	if !investmentRow.ProfitApplicable {
		t.Fatal("expected investment row to calculate profit")
	}
	if investmentRow.Profit != 5000 {
		t.Fatalf("expected profit 5000, got %f", investmentRow.Profit)
	}
	if investmentRow.ProfitPercent != 0 {
		t.Fatalf("expected zero profit percent when bought price is zero, got %f", investmentRow.ProfitPercent)
	}

	assertFloatEqual(t, 15000, page.Summary.TotalBought)
	assertFloatEqual(t, 34000, page.Summary.TotalCurrent)
	assertFloatEqual(t, 8000, page.Summary.TotalProfit)
	assertFloatInDelta(t, 0.5333333333, page.Summary.TotalProfitRate)
	assertFloatEqual(t, 11000, page.Summary.TotalCash)
	assertFloatEqual(t, 23000, page.Summary.TotalNonCash)
	assertFloatInDelta(t, 0.3235294117, page.Summary.CashRatio)
	if page.Comparison != nil {
		t.Fatal("expected no comparison for first snapshot")
	}
}

func TestRecordService_BuildHomePageCalculatesComparisonUsingOverlappingAssetsOnly(t *testing.T) {
	t.Parallel()

	service := RecordService{}
	current := &dto.Snapshot{
		RecordDate: time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
		Items: []dto.SnapshotItem{
			{
				AssetID:           1,
				AssetName:         "SET50 ETF",
				AssetTypeName:     "Investment",
				AssetTypeOrdering: 1,
				AssetOrdering:     1,
				Broker:            "KKP",
				BoughtPrice:       12000,
				CurrentPrice:      15000,
			},
			{
				AssetID:           2,
				AssetName:         "Emergency Fund",
				AssetTypeName:     "Cash",
				AssetTypeOrdering: 2,
				AssetOrdering:     1,
				Broker:            "SCB",
				IsCash:            true,
				BoughtPrice:       0,
				CurrentPrice:      7000,
			},
			{
				AssetID:           3,
				AssetName:         "New Asset",
				AssetTypeName:     "Investment",
				AssetTypeOrdering: 1,
				AssetOrdering:     2,
				Broker:            "IBKR",
				BoughtPrice:       3000,
				CurrentPrice:      3500,
			},
		},
	}
	previous := &dto.Snapshot{
		RecordDate: time.Date(2026, time.March, 12, 0, 0, 0, 0, time.UTC),
		Items: []dto.SnapshotItem{
			{
				AssetID:           1,
				AssetName:         "SET50 ETF",
				AssetTypeName:     "Investment",
				AssetTypeOrdering: 1,
				AssetOrdering:     1,
				Broker:            "KKP",
				BoughtPrice:       10000,
				CurrentPrice:      14000,
			},
			{
				AssetID:           2,
				AssetName:         "Emergency Fund",
				AssetTypeName:     "Cash",
				AssetTypeOrdering: 2,
				AssetOrdering:     1,
				Broker:            "SCB",
				IsCash:            true,
				BoughtPrice:       0,
				CurrentPrice:      6000,
			},
			{
				AssetID:           99,
				AssetName:         "Removed Asset",
				AssetTypeName:     "Investment",
				AssetTypeOrdering: 1,
				AssetOrdering:     3,
				Broker:            "Old Broker",
				BoughtPrice:       8000,
				CurrentPrice:      7500,
			},
		},
	}

	page := service.BuildHomePage(current, previous)

	if page.PreviousSnapshot == nil {
		t.Fatal("expected previous snapshot date")
	}
	if page.Comparison == nil {
		t.Fatal("expected comparison")
	}
	if page.Comparison.PreviousSnapshotDate != previous.RecordDate {
		t.Fatalf("expected comparison date %v, got %v", previous.RecordDate, page.Comparison.PreviousSnapshotDate)
	}
	assertFloatEqual(t, 2000, page.Comparison.BoughtChange)
	assertFloatEqual(t, 2000, page.Comparison.CurrentChange)
	assertFloatEqual(t, -1000, page.Comparison.ProfitChange)
	assertFloatInDelta(t, -0.15, page.Comparison.ProfitRateChange)
	assertFloatEqual(t, 1000, page.Comparison.CashChange)
	assertFloatEqual(t, 1000, page.Comparison.NonCashChange)
	assertFloatInDelta(t, 0.0181818182, page.Comparison.CashRatioChange)
}

func TestRecordService_GetHomePageSetsNavigationFlagsFromSnapshotOffset(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{
		snapshots: map[int]*dto.Snapshot{
			0: {
				RecordDate: time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
				Items: []dto.SnapshotItem{
					{AssetID: 1, AssetName: "ETF", AssetTypeName: "Investment", BoughtPrice: 10, CurrentPrice: 12},
				},
			},
			1: {
				RecordDate: time.Date(2026, time.March, 12, 0, 0, 0, 0, time.UTC),
				Items: []dto.SnapshotItem{
					{AssetID: 1, AssetName: "ETF", AssetTypeName: "Investment", BoughtPrice: 9, CurrentPrice: 11},
				},
			},
		},
	})

	page, err := service.GetHomePage(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetHomePage returned error: %v", err)
	}

	if !page.CanNavigateForward {
		t.Fatal("expected forward navigation when offset is greater than zero")
	}
	if page.CanNavigateBack {
		t.Fatal("expected no older snapshot after the second latest one")
	}
	if len(page.SnapshotOptions) != 2 {
		t.Fatalf("expected 2 snapshot options, got %d", len(page.SnapshotOptions))
	}
}

func TestRecordService_GetEditSnapshotPageBuildsGroupedRowsAndAvailableAssets(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{
		editableSnapshot: &dto.EditableSnapshot{
			ID:         7,
			RecordDate: time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
			Items: []dto.EditableSnapshotItem{
				{
					AssetID:           2,
					AssetName:         "Wallet",
					AssetTypeName:     "Cash",
					AssetTypeOrdering: 2,
					AssetOrdering:     1,
					Broker:            "SCB",
					IsCash:            true,
					IsActive:          true,
					CurrentPrice:      5000,
				},
				{
					AssetID:           1,
					AssetName:         "SET50 ETF",
					AssetTypeName:     "Investment",
					AssetTypeOrdering: 1,
					AssetOrdering:     2,
					Broker:            "KKP",
					IsCash:            false,
					IsActive:          true,
					BoughtPrice:       12000,
					CurrentPrice:      15000,
					Remarks:           "Core",
				},
			},
			AvailableAssets: []dto.EditableAssetOption{
				{
					AssetID:           3,
					AssetName:         "US Index Fund",
					AssetTypeName:     "Investment",
					AssetTypeOrdering: 1,
					AssetOrdering:     1,
					Broker:            "IBKR",
					IsActive:          false,
				},
			},
		},
	})

	page, err := service.GetEditSnapshotPage(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetEditSnapshotPage returned error: %v", err)
	}

	if page.SnapshotID != 7 {
		t.Fatalf("expected snapshot id 7, got %d", page.SnapshotID)
	}
	if page.SnapshotDate != "2026-04-12" {
		t.Fatalf("expected snapshot date 2026-04-12, got %s", page.SnapshotDate)
	}
	if len(page.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(page.Groups))
	}
	if page.Groups[0].AssetTypeName != "Investment" {
		t.Fatalf("expected investment group first, got %s", page.Groups[0].AssetTypeName)
	}
	if len(page.AvailableAssetGroups) != 1 || page.AvailableAssetGroups[0].Options[0].AssetID != 3 {
		t.Fatalf("expected available asset id 3, got %+v", page.AvailableAssetGroups)
	}
}

func TestRecordService_GetNewSnapshotPageBuildsNewModeDraft(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{
		newSnapshotDraft: &dto.EditableSnapshot{
			ID:         0,
			RecordDate: time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
			Items: []dto.EditableSnapshotItem{
				{
					AssetID:           2,
					AssetName:         "Wallet",
					AssetTypeName:     "Cash",
					AssetTypeOrdering: 2,
					AssetOrdering:     1,
					Broker:            "SCB",
					IsCash:            true,
					IsActive:          true,
					CurrentPrice:      5000,
				},
				{
					AssetID:           1,
					AssetName:         "SET50 ETF",
					AssetTypeName:     "Investment",
					AssetTypeOrdering: 1,
					AssetOrdering:     2,
					Broker:            "KKP",
					IsCash:            false,
					IsActive:          true,
					BoughtPrice:       12500,
					CurrentPrice:      15000,
					Remarks:           "Core",
				},
			},
			AvailableAssets: []dto.EditableAssetOption{
				{
					AssetID:           3,
					AssetName:         "Old Asset",
					AssetTypeName:     "Investment",
					AssetTypeOrdering: 1,
					AssetOrdering:     1,
					Broker:            "IBKR",
					IsActive:          false,
				},
			},
		},
	})

	page, err := service.GetNewSnapshotPage(
		context.Background(),
		time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("GetNewSnapshotPage returned error: %v", err)
	}

	if page.Mode != "new" {
		t.Fatalf("expected new mode, got %q", page.Mode)
	}
	if page.SnapshotID != 0 {
		t.Fatalf("expected draft snapshot id 0, got %d", page.SnapshotID)
	}
	if page.SnapshotDate != "2026-04-12" {
		t.Fatalf("expected snapshot date 2026-04-12, got %s", page.SnapshotDate)
	}
	if len(page.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(page.Groups))
	}
	if page.Groups[0].AssetTypeName != "Investment" {
		t.Fatalf("expected investment group first, got %s", page.Groups[0].AssetTypeName)
	}
	if len(page.AvailableAssetGroups) != 1 || page.AvailableAssetGroups[0].Options[0].AssetID != 3 {
		t.Fatalf("expected available asset id 3, got %+v", page.AvailableAssetGroups)
	}
}

func TestRecordService_SaveSnapshotRejectsEmptyRows(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{})

	_, err := service.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   1,
		SnapshotDate: "2026-04-12",
	})
	if !errors.Is(err, recorderr.ErrSnapshotHasNoRows) {
		t.Fatalf("expected ErrSnapshotHasNoRows, got %v", err)
	}
}

func TestRecordService_SaveSnapshotReturnsRepositoryResult(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{
		saveResult: dto.SaveSnapshotResult{Offset: 3},
	})

	result, err := service.SaveSnapshot(context.Background(), dto.SaveSnapshotInput{
		SnapshotID:   1,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 10, CurrentPrice: 12},
		},
	})
	if err != nil {
		t.Fatalf("SaveSnapshot returned error: %v", err)
	}
	if result.Offset != 3 {
		t.Fatalf("expected offset 3, got %d", result.Offset)
	}
}

func TestRecordService_CreateSnapshotRejectsEmptyRows(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{})

	_, err := service.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-04-12",
	})
	if !errors.Is(err, recorderr.ErrSnapshotHasNoRows) {
		t.Fatalf("expected ErrSnapshotHasNoRows, got %v", err)
	}
}

func TestRecordService_CreateSnapshotReturnsRepositoryResult(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{
		createResult: dto.CreateSnapshotResult{Offset: 0},
	})

	result, err := service.CreateSnapshot(context.Background(), dto.CreateSnapshotInput{
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 10, CurrentPrice: 12},
		},
	})
	if err != nil {
		t.Fatalf("CreateSnapshot returned error: %v", err)
	}
	if result.Offset != 0 {
		t.Fatalf("expected offset 0, got %d", result.Offset)
	}
}

func TestRecordService_DeleteSnapshotReturnsRepositoryResult(t *testing.T) {
	t.Parallel()

	service := NewRecordService(snapshotReaderStub{
		deleteResult: dto.DeleteSnapshotResult{
			Offset:       0,
			HasSnapshots: true,
		},
	})

	result, err := service.DeleteSnapshot(context.Background(), dto.DeleteSnapshotInput{
		SnapshotID: 2,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("DeleteSnapshot returned error: %v", err)
	}
	if !result.HasSnapshots {
		t.Fatal("expected remaining snapshots")
	}
	if result.Offset != 0 {
		t.Fatalf("expected offset 0, got %d", result.Offset)
	}
}

func assertFloatEqual(t *testing.T, want float64, got float64) {
	t.Helper()
	if want != got {
		t.Fatalf("expected %f, got %f", want, got)
	}
}

func assertFloatInDelta(t *testing.T, want float64, got float64) {
	t.Helper()
	const tolerance = 0.0000001
	delta := want - got
	if delta < 0 {
		delta = -delta
	}
	if delta > tolerance {
		t.Fatalf("expected %f, got %f", want, got)
	}
}

type snapshotReaderStub struct {
	snapshots        map[int]*dto.Snapshot
	snapshotOptions  []dto.SnapshotOption
	editableSnapshot *dto.EditableSnapshot
	newSnapshotDraft *dto.EditableSnapshot
	saveResult       dto.SaveSnapshotResult
	saveErr          error
	createResult     dto.CreateSnapshotResult
	createErr        error
	deleteResult     dto.DeleteSnapshotResult
	deleteErr        error
}

func (s snapshotReaderStub) GetSnapshotByOffset(
	_ context.Context,
	offset int,
) (*dto.Snapshot, error) {
	return s.snapshots[offset], nil
}

func (s snapshotReaderStub) ListSnapshotOptions(context.Context) ([]dto.SnapshotOption, error) {
	if len(s.snapshotOptions) > 0 {
		return s.snapshotOptions, nil
	}

	options := make([]dto.SnapshotOption, 0, len(s.snapshots))
	for offset := range s.snapshots {
		options = append(options, dto.SnapshotOption{
			Offset: offset,
			Label:  s.snapshots[offset].RecordDate.Format("02 Jan 2006"),
		})
	}

	return options, nil
}

func (s snapshotReaderStub) GetEditableSnapshotByOffset(
	context.Context,
	int,
) (*dto.EditableSnapshot, error) {
	return s.editableSnapshot, nil
}

func (s snapshotReaderStub) GetNewSnapshotDraft(
	context.Context,
	time.Time,
) (*dto.EditableSnapshot, error) {
	return s.newSnapshotDraft, nil
}

func (s snapshotReaderStub) SaveSnapshot(
	context.Context,
	dto.SaveSnapshotInput,
) (dto.SaveSnapshotResult, error) {
	if s.saveErr != nil {
		return dto.SaveSnapshotResult{}, s.saveErr
	}

	return s.saveResult, nil
}

func (s snapshotReaderStub) CreateSnapshot(
	context.Context,
	dto.CreateSnapshotInput,
) (dto.CreateSnapshotResult, error) {
	if s.createErr != nil {
		return dto.CreateSnapshotResult{}, s.createErr
	}

	return s.createResult, nil
}

func (s snapshotReaderStub) DeleteSnapshot(
	context.Context,
	dto.DeleteSnapshotInput,
) (dto.DeleteSnapshotResult, error) {
	if s.deleteErr != nil {
		return dto.DeleteSnapshotResult{}, s.deleteErr
	}

	return s.deleteResult, nil
}
