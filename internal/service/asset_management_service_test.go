package service

import (
	"context"
	"errors"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type stubAssetManagementRepository struct {
	createAssetTypeErr error
	updateAssetErr     error
	reorderAssetsErr   error
}

func (s stubAssetManagementRepository) GetPage(context.Context) (dto.AssetManagementPage, error) {
	return dto.AssetManagementPage{}, nil
}

func (s stubAssetManagementRepository) CreateAssetType(context.Context, dto.CreateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	return dto.AssetTypeMutationResult{}, s.createAssetTypeErr
}

func (s stubAssetManagementRepository) UpdateAssetType(context.Context, dto.UpdateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	return dto.AssetTypeMutationResult{}, nil
}

func (s stubAssetManagementRepository) CreateAsset(context.Context, dto.CreateAssetInput) (dto.AssetMutationResult, error) {
	return dto.AssetMutationResult{}, nil
}

func (s stubAssetManagementRepository) UpdateAsset(context.Context, dto.UpdateAssetInput) (dto.AssetMutationResult, error) {
	return dto.AssetMutationResult{}, s.updateAssetErr
}

func (s stubAssetManagementRepository) ReorderAssetTypes(context.Context, dto.ReorderAssetTypesInput) error {
	return nil
}

func (s stubAssetManagementRepository) ReorderAssets(context.Context, dto.ReorderAssetInput) error {
	return s.reorderAssetsErr
}

func TestAssetManagementService_CreateAssetTypePreservesDuplicateNameError(t *testing.T) {
	t.Parallel()

	service := NewAssetManagementService(stubAssetManagementRepository{
		createAssetTypeErr: recorderr.ErrAssetTypeNameExists,
	})

	_, err := service.CreateAssetType(context.Background(), dto.CreateAssetTypeInput{
		Name:     "Investment",
		IsActive: true,
	})
	if !errors.Is(err, recorderr.ErrAssetTypeNameExists) {
		t.Fatalf("expected duplicate asset type name error, got %v", err)
	}
}

func TestAssetManagementService_UpdateAssetPreservesDuplicateAssetNameError(t *testing.T) {
	t.Parallel()

	service := NewAssetManagementService(stubAssetManagementRepository{
		updateAssetErr: recorderr.ErrAssetNameExists,
	})

	_, err := service.UpdateAsset(context.Background(), dto.UpdateAssetInput{
		ID:            1,
		Name:          "Wallet",
		AssetTypeID:   2,
		Broker:        "SCB",
		IsCash:        true,
		IsActive:      true,
		AutoIncrement: 0,
	})
	if !errors.Is(err, recorderr.ErrAssetNameExists) {
		t.Fatalf("expected duplicate asset name error, got %v", err)
	}
}

func TestAssetManagementService_ReorderAssetsPreservesAssetTypeNotFound(t *testing.T) {
	t.Parallel()

	service := NewAssetManagementService(stubAssetManagementRepository{
		reorderAssetsErr: recorderr.ErrAssetTypeNotFound,
	})

	err := service.ReorderAssets(context.Background(), dto.ReorderAssetInput{
		AssetTypeID: 1,
		OrderedIDs:  []int64{1, 2},
	})
	if !errors.Is(err, recorderr.ErrAssetTypeNotFound) {
		t.Fatalf("expected asset type not found error, got %v", err)
	}
}
