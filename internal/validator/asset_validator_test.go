package validator

import (
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

func TestAssetValidator_ValidateCreateAssetTypeInputRejectsBlankName(t *testing.T) {
	t.Parallel()

	validator := AssetValidator{}

	err := validator.ValidateCreateAssetTypeInput(dto.CreateAssetTypeInput{
		Name: "   ",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAssetValidator_ValidateCreateAssetInputRejectsNegativeAutoIncrement(t *testing.T) {
	t.Parallel()

	validator := AssetValidator{}

	err := validator.ValidateCreateAssetInput(dto.CreateAssetInput{
		Name:          "ETF",
		AssetTypeID:   1,
		AutoIncrement: -1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAssetValidator_ValidateUpdateAssetInputAcceptsValidPayload(t *testing.T) {
	t.Parallel()

	validator := AssetValidator{}

	err := validator.ValidateUpdateAssetInput(dto.UpdateAssetInput{
		ID:            1,
		Name:          "ETF",
		AssetTypeID:   2,
		Broker:        "KKP",
		IsCash:        false,
		IsActive:      true,
		AutoIncrement: 6500,
	})
	if err != nil {
		t.Fatalf("ValidateUpdateAssetInput returned error: %v", err)
	}
}

func TestAssetValidator_ValidateReorderAssetInputRejectsDuplicateIDs(t *testing.T) {
	t.Parallel()

	validator := AssetValidator{}

	err := validator.ValidateReorderAssetInput(dto.ReorderAssetInput{
		AssetTypeID: 1,
		OrderedIDs:  []int64{1, 1},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAssetValidator_ValidateReorderAssetTypesInputAcceptsValidPayload(t *testing.T) {
	t.Parallel()

	validator := AssetValidator{}

	err := validator.ValidateReorderAssetTypesInput(dto.ReorderAssetTypesInput{
		OrderedIDs: []int64{2, 1},
		ActiveOnly: true,
	})
	if err != nil {
		t.Fatalf("ValidateReorderAssetTypesInput returned error: %v", err)
	}
}
