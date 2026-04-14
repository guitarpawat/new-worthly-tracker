package validator

import (
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

func TestRecordValidator_ValidateSaveSnapshotInputAcceptsValidPayload(t *testing.T) {
	t.Parallel()

	validator := RecordValidator{}

	err := validator.ValidateSaveSnapshotInput(dto.SaveSnapshotInput{
		SnapshotID:   1,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 1234.50, CurrentPrice: -55.25, Remarks: "valid"},
		},
	})
	if err != nil {
		t.Fatalf("ValidateSaveSnapshotInput returned error: %v", err)
	}
}

func TestRecordValidator_ValidateSaveSnapshotInputRejectsMoreThanTwoDecimalPlaces(t *testing.T) {
	t.Parallel()

	validator := RecordValidator{}

	err := validator.ValidateSaveSnapshotInput(dto.SaveSnapshotInput{
		SnapshotID:   1,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 123.456, CurrentPrice: 1},
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRecordValidator_ValidateSaveSnapshotInputRejectsDuplicateAssetIDs(t *testing.T) {
	t.Parallel()

	validator := RecordValidator{}

	err := validator.ValidateSaveSnapshotInput(dto.SaveSnapshotInput{
		SnapshotID:   1,
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 1, CurrentPrice: 1},
			{AssetID: 1, BoughtPrice: 2, CurrentPrice: 2},
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRecordValidator_ValidateCreateSnapshotInputAcceptsValidPayload(t *testing.T) {
	t.Parallel()

	validator := RecordValidator{}

	err := validator.ValidateCreateSnapshotInput(dto.CreateSnapshotInput{
		SnapshotDate: "2026-04-12",
		Items: []dto.SaveSnapshotItemInput{
			{AssetID: 1, BoughtPrice: 1234.50, CurrentPrice: -55.25, Remarks: "valid"},
		},
	})
	if err != nil {
		t.Fatalf("ValidateCreateSnapshotInput returned error: %v", err)
	}
}

func TestRecordValidator_ValidateCreateSnapshotInputRejectsEmptyPayload(t *testing.T) {
	t.Parallel()

	validator := RecordValidator{}

	err := validator.ValidateCreateSnapshotInput(dto.CreateSnapshotInput{
		SnapshotDate: "2026-04-12",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
