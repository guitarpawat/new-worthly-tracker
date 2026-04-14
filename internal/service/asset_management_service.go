package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type AssetManagementReader interface {
	GetPage(ctx context.Context) (dto.AssetManagementPage, error)
	CreateAssetType(ctx context.Context, input dto.CreateAssetTypeInput) (dto.AssetTypeMutationResult, error)
	UpdateAssetType(ctx context.Context, input dto.UpdateAssetTypeInput) (dto.AssetTypeMutationResult, error)
	CreateAsset(ctx context.Context, input dto.CreateAssetInput) (dto.AssetMutationResult, error)
	UpdateAsset(ctx context.Context, input dto.UpdateAssetInput) (dto.AssetMutationResult, error)
	ReorderAssetTypes(ctx context.Context, input dto.ReorderAssetTypesInput) error
	ReorderAssets(ctx context.Context, input dto.ReorderAssetInput) error
}

type AssetManagementService struct {
	repository AssetManagementReader
}

func NewAssetManagementService(repository AssetManagementReader) *AssetManagementService {
	return &AssetManagementService{repository: repository}
}

func (s *AssetManagementService) GetPage(ctx context.Context) (dto.AssetManagementPage, error) {
	page, err := s.repository.GetPage(ctx)
	if err != nil {
		return dto.AssetManagementPage{}, fmt.Errorf("get asset management page: %w", err)
	}
	return page, nil
}

func (s *AssetManagementService) CreateAssetType(ctx context.Context, input dto.CreateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	result, err := s.repository.CreateAssetType(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNameExists):
			return dto.AssetTypeMutationResult{}, recorderr.ErrAssetTypeNameExists
		}
		return dto.AssetTypeMutationResult{}, fmt.Errorf("create asset type: %w", err)
	}
	return result, nil
}

func (s *AssetManagementService) UpdateAssetType(ctx context.Context, input dto.UpdateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	result, err := s.repository.UpdateAssetType(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNameExists):
			return dto.AssetTypeMutationResult{}, recorderr.ErrAssetTypeNameExists
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return dto.AssetTypeMutationResult{}, recorderr.ErrAssetTypeNotFound
		default:
			return dto.AssetTypeMutationResult{}, fmt.Errorf("update asset type: %w", err)
		}
	}
	return result, nil
}

func (s *AssetManagementService) CreateAsset(ctx context.Context, input dto.CreateAssetInput) (dto.AssetMutationResult, error) {
	result, err := s.repository.CreateAsset(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrAssetNameExists):
			return dto.AssetMutationResult{}, recorderr.ErrAssetNameExists
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return dto.AssetMutationResult{}, recorderr.ErrAssetTypeNotFound
		case errors.Is(err, recorderr.ErrAssetTypeInactive):
			return dto.AssetMutationResult{}, recorderr.ErrAssetTypeInactive
		default:
			return dto.AssetMutationResult{}, fmt.Errorf("create asset: %w", err)
		}
	}
	return result, nil
}

func (s *AssetManagementService) UpdateAsset(ctx context.Context, input dto.UpdateAssetInput) (dto.AssetMutationResult, error) {
	result, err := s.repository.UpdateAsset(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrAssetNotFound):
			return dto.AssetMutationResult{}, recorderr.ErrAssetNotFound
		case errors.Is(err, recorderr.ErrAssetNameExists):
			return dto.AssetMutationResult{}, recorderr.ErrAssetNameExists
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return dto.AssetMutationResult{}, recorderr.ErrAssetTypeNotFound
		case errors.Is(err, recorderr.ErrAssetTypeInactive):
			return dto.AssetMutationResult{}, recorderr.ErrAssetTypeInactive
		default:
			return dto.AssetMutationResult{}, fmt.Errorf("update asset: %w", err)
		}
	}
	return result, nil
}

func (s *AssetManagementService) ReorderAssetTypes(ctx context.Context, input dto.ReorderAssetTypesInput) error {
	if err := s.repository.ReorderAssetTypes(ctx, input); err != nil {
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return recorderr.ErrAssetTypeNotFound
		default:
			return fmt.Errorf("reorder asset types: %w", err)
		}
	}
	return nil
}

func (s *AssetManagementService) ReorderAssets(ctx context.Context, input dto.ReorderAssetInput) error {
	if err := s.repository.ReorderAssets(ctx, input); err != nil {
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return recorderr.ErrAssetTypeNotFound
		default:
			return fmt.Errorf("reorder assets: %w", err)
		}
	}
	return nil
}
