package validator

import (
	"fmt"
	"math"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

type RecordValidator struct{}

func (RecordValidator) ValidateSnapshotOffset(offset int) error {
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}

func (RecordValidator) ValidateSaveSnapshotInput(input dto.SaveSnapshotInput) error {
	if input.SnapshotID <= 0 {
		return fmt.Errorf("snapshot id must be positive")
	}

	if err := validateSnapshotPayload(input.SnapshotDate, input.Items); err != nil {
		return err
	}

	return nil
}

func (RecordValidator) ValidateCreateSnapshotInput(input dto.CreateSnapshotInput) error {
	if err := validateSnapshotPayload(input.SnapshotDate, input.Items); err != nil {
		return err
	}

	return nil
}

func validateSnapshotPayload(snapshotDate string, items []dto.SaveSnapshotItemInput) error {
	if _, err := time.Parse("2006-01-02", snapshotDate); err != nil {
		return fmt.Errorf("snapshot date must use YYYY-MM-DD")
	}

	if len(items) == 0 {
		return fmt.Errorf("snapshot must contain at least one asset")
	}

	seenAssetIDs := make(map[int64]struct{}, len(items))
	for index, item := range items {
		if item.AssetID <= 0 {
			return fmt.Errorf("item %d asset id must be positive", index+1)
		}
		if _, found := seenAssetIDs[item.AssetID]; found {
			return fmt.Errorf("asset %d is duplicated in snapshot", item.AssetID)
		}
		seenAssetIDs[item.AssetID] = struct{}{}

		if !isFiniteNumber(item.BoughtPrice) || !isFiniteNumber(item.CurrentPrice) {
			return fmt.Errorf("item %d contains invalid numeric values", index+1)
		}
		if hasMoreThanTwoDecimalPlaces(item.BoughtPrice) || hasMoreThanTwoDecimalPlaces(item.CurrentPrice) {
			return fmt.Errorf("item %d allows at most 2 decimal places", index+1)
		}
	}

	return nil
}

func (RecordValidator) ValidateDeleteSnapshotInput(input dto.DeleteSnapshotInput) error {
	if input.SnapshotID <= 0 {
		return fmt.Errorf("snapshot id must be positive")
	}
	if input.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	return nil
}

func isFiniteNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func hasMoreThanTwoDecimalPlaces(value float64) bool {
	scaled := math.Abs(value * 100)
	return math.Abs(scaled-math.Round(scaled)) > 0.0000001
}
