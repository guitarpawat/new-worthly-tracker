package validator

import (
	"fmt"
	"math"
	"strings"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

type AssetValidator struct{}

func (AssetValidator) ValidateCreateAssetTypeInput(input dto.CreateAssetTypeInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("asset type name is required")
	}

	return nil
}

func (AssetValidator) ValidateUpdateAssetTypeInput(input dto.UpdateAssetTypeInput) error {
	if input.ID <= 0 {
		return fmt.Errorf("asset type id must be positive")
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("asset type name is required")
	}

	return nil
}

func (AssetValidator) ValidateCreateAssetInput(input dto.CreateAssetInput) error {
	if err := validateAssetPayload(input.Name, input.AssetTypeID, input.AutoIncrement); err != nil {
		return err
	}

	return nil
}

func (AssetValidator) ValidateUpdateAssetInput(input dto.UpdateAssetInput) error {
	if input.ID <= 0 {
		return fmt.Errorf("asset id must be positive")
	}
	if err := validateAssetPayload(input.Name, input.AssetTypeID, input.AutoIncrement); err != nil {
		return err
	}

	return nil
}

func (AssetValidator) ValidateReorderAssetTypesInput(input dto.ReorderAssetTypesInput) error {
	return validateReorderIDs(input.OrderedIDs)
}

func (AssetValidator) ValidateReorderAssetInput(input dto.ReorderAssetInput) error {
	if input.AssetTypeID <= 0 {
		return fmt.Errorf("asset type id must be positive")
	}
	return validateReorderIDs(input.OrderedIDs)
}

func validateAssetPayload(name string, assetTypeID int64, autoIncrement float64) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("asset name is required")
	}
	if assetTypeID <= 0 {
		return fmt.Errorf("asset type is required")
	}
	if math.IsNaN(autoIncrement) || math.IsInf(autoIncrement, 0) {
		return fmt.Errorf("auto increment must be a valid number")
	}
	if autoIncrement < 0 {
		return fmt.Errorf("auto increment must be non-negative")
	}
	if hasMoreThanTwoDecimalPlaces(autoIncrement) {
		return fmt.Errorf("auto increment allows at most 2 decimal places")
	}

	return nil
}

func validateReorderIDs(ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("reorder list must not be empty")
	}

	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("reorder ids must be positive")
		}
		if _, found := seen[id]; found {
			return fmt.Errorf("reorder ids must be unique")
		}
		seen[id] = struct{}{}
	}

	return nil
}
