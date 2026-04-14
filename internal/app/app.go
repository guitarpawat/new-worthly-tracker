package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
	"github.com/guitarpawat/worthly-tracker/internal/service"
	"github.com/guitarpawat/worthly-tracker/internal/validator"
)

type App struct {
	ctx                    context.Context
	logger                 *slog.Logger
	recordService          *service.RecordService
	assetManagementService *service.AssetManagementService
	progressService        *service.ProgressService
	goalService            *service.GoalService
	demoData               *service.DemoDataService
	validator              validator.RecordValidator
	assetValidator         validator.AssetValidator
	progressValidator      validator.ProgressValidator
	goalValidator          validator.GoalValidator
}

func New(
	logger *slog.Logger,
	recordService *service.RecordService,
	assetManagementService *service.AssetManagementService,
	progressService *service.ProgressService,
	goalService *service.GoalService,
	demoData *service.DemoDataService,
) *App {
	return &App{
		logger:                 logger,
		recordService:          recordService,
		assetManagementService: assetManagementService,
		progressService:        progressService,
		goalService:            goalService,
		demoData:               demoData,
		validator:              validator.RecordValidator{},
		assetValidator:         validator.AssetValidator{},
		progressValidator:      validator.ProgressValidator{},
		goalValidator:          validator.GoalValidator{},
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("application startup completed")
}

func (a *App) GetHomePage(offset int) (dto.HomePage, error) {
	if err := a.validator.ValidateSnapshotOffset(offset); err != nil {
		return dto.HomePage{}, err
	}

	page, err := a.recordService.GetHomePage(a.ctx, offset)
	if err != nil {
		a.logger.Error("load home page", "offset", offset, "err", err)
		return dto.HomePage{}, fmt.Errorf("load home page: %w", err)
	}

	if page.HasSnapshot {
		a.logger.Info(
			"loaded home page",
			"offset", offset,
			"snapshot_date", page.SnapshotDate.Format("2006-01-02"),
			"snapshot_options", len(page.SnapshotOptions),
			"can_navigate_back", page.CanNavigateBack,
			"can_navigate_forward", page.CanNavigateForward,
		)
	} else {
		a.logger.Info("loaded home page", "offset", offset, "snapshot_date", "", "snapshot_options", len(page.SnapshotOptions))
	}

	return page, nil
}

func (a *App) GetEditSnapshotPage(offset int) (dto.EditSnapshotPage, error) {
	if err := a.validator.ValidateSnapshotOffset(offset); err != nil {
		return dto.EditSnapshotPage{}, err
	}

	page, err := a.recordService.GetEditSnapshotPage(a.ctx, offset)
	if err != nil {
		a.logger.Error("load edit snapshot page", "offset", offset, "err", err)
		return dto.EditSnapshotPage{}, fmt.Errorf("load edit snapshot page: %w", err)
	}

	a.logger.Info(
		"loaded edit snapshot page",
		"offset", offset,
		"snapshot_id", page.SnapshotID,
		"snapshot_date", page.SnapshotDate,
		"groups", len(page.Groups),
		"available_asset_groups", len(page.AvailableAssetGroups),
	)

	return page, nil
}

func (a *App) GetNewSnapshotPage() (dto.EditSnapshotPage, error) {
	recordDate := time.Now().In(time.Local)

	page, err := a.recordService.GetNewSnapshotPage(a.ctx, recordDate)
	if err != nil {
		a.logger.Error("load new snapshot page", "err", err)
		return dto.EditSnapshotPage{}, fmt.Errorf("load new snapshot page: %w", err)
	}

	a.logger.Info(
		"loaded new snapshot page",
		"snapshot_date", page.SnapshotDate,
		"groups", len(page.Groups),
		"available_asset_groups", len(page.AvailableAssetGroups),
	)

	return page, nil
}

func (a *App) GetAssetManagementPage() (dto.AssetManagementPage, error) {
	page, err := a.assetManagementService.GetPage(a.ctx)
	if err != nil {
		a.logger.Error("load asset management page", "err", err)
		return dto.AssetManagementPage{}, fmt.Errorf("load asset management page: %w", err)
	}

	a.logger.Info(
		"loaded asset management page",
		"asset_types", len(page.AssetTypes),
		"assets", len(page.Assets),
		"active_asset_types", len(page.ActiveAssetTypes),
	)

	return page, nil
}

func (a *App) GetProgressPage(filter dto.ProgressFilter) (dto.ProgressPage, error) {
	if err := a.progressValidator.ValidateProgressFilter(filter); err != nil {
		return dto.ProgressPage{}, err
	}

	page, err := a.progressService.GetPage(a.ctx, filter)
	if err != nil {
		a.logger.Error("load progress page", "start_date", filter.StartDate, "end_date", filter.EndDate, "err", err)
		return dto.ProgressPage{}, fmt.Errorf("load progress page: %w", err)
	}

	a.logger.Info(
		"loaded progress page",
		"has_data", page.HasData,
		"start_date", page.Filter.StartDate,
		"end_date", page.Filter.EndDate,
		"trend_points", len(page.TrendPoints),
		"goals", len(page.Goals),
	)

	return page, nil
}

func (a *App) CreateGoal(input dto.CreateGoalInput) (dto.GoalMutationResult, error) {
	if err := a.goalValidator.ValidateCreateGoalInput(input); err != nil {
		return dto.GoalMutationResult{}, err
	}

	result, err := a.goalService.CreateGoal(a.ctx, input)
	if err != nil {
		a.logger.Error("create goal", "name", input.Name, "err", err)
		return dto.GoalMutationResult{}, fmt.Errorf("create goal: %w", err)
	}

	a.logger.Info("created goal", "goal_id", result.ID, "name", input.Name)
	return result, nil
}

func (a *App) UpdateGoal(input dto.UpdateGoalInput) (dto.GoalMutationResult, error) {
	if err := a.goalValidator.ValidateUpdateGoalInput(input); err != nil {
		return dto.GoalMutationResult{}, err
	}

	result, err := a.goalService.UpdateGoal(a.ctx, input)
	if err != nil {
		a.logger.Error("update goal", "goal_id", input.ID, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrGoalNotFound):
			return dto.GoalMutationResult{}, fmt.Errorf("goal no longer exists")
		default:
			return dto.GoalMutationResult{}, fmt.Errorf("update goal: %w", err)
		}
	}

	a.logger.Info("updated goal", "goal_id", input.ID)
	return result, nil
}

func (a *App) DeleteGoal(input dto.DeleteGoalInput) error {
	if err := a.goalValidator.ValidateDeleteGoalInput(input); err != nil {
		return err
	}

	if err := a.goalService.DeleteGoal(a.ctx, input); err != nil {
		a.logger.Error("delete goal", "goal_id", input.ID, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrGoalNotFound):
			return fmt.Errorf("goal no longer exists")
		default:
			return fmt.Errorf("delete goal: %w", err)
		}
	}

	a.logger.Info("deleted goal", "goal_id", input.ID)
	return nil
}

func (a *App) CreateAssetType(input dto.CreateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	if err := a.assetValidator.ValidateCreateAssetTypeInput(input); err != nil {
		return dto.AssetTypeMutationResult{}, err
	}

	result, err := a.assetManagementService.CreateAssetType(a.ctx, input)
	if err != nil {
		a.logger.Error("create asset type", "name", input.Name, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNameExists):
			return dto.AssetTypeMutationResult{}, fmt.Errorf("asset type name already exists")
		default:
			return dto.AssetTypeMutationResult{}, fmt.Errorf("create asset type: %w", err)
		}
	}

	a.logger.Info("created asset type", "asset_type_id", result.ID, "name", input.Name, "is_active", input.IsActive)
	return result, nil
}

func (a *App) UpdateAssetType(input dto.UpdateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	if err := a.assetValidator.ValidateUpdateAssetTypeInput(input); err != nil {
		return dto.AssetTypeMutationResult{}, err
	}

	result, err := a.assetManagementService.UpdateAssetType(a.ctx, input)
	if err != nil {
		a.logger.Error("update asset type", "asset_type_id", input.ID, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNameExists):
			return dto.AssetTypeMutationResult{}, fmt.Errorf("asset type name already exists")
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return dto.AssetTypeMutationResult{}, fmt.Errorf("asset type no longer exists")
		default:
			return dto.AssetTypeMutationResult{}, fmt.Errorf("update asset type: %w", err)
		}
	}

	a.logger.Info("updated asset type", "asset_type_id", input.ID, "is_active", input.IsActive)
	return result, nil
}

func (a *App) CreateAsset(input dto.CreateAssetInput) (dto.AssetMutationResult, error) {
	if err := a.assetValidator.ValidateCreateAssetInput(input); err != nil {
		return dto.AssetMutationResult{}, err
	}

	result, err := a.assetManagementService.CreateAsset(a.ctx, input)
	if err != nil {
		a.logger.Error("create asset", "name", input.Name, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrAssetNameExists):
			return dto.AssetMutationResult{}, fmt.Errorf("asset name already exists in this asset type")
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return dto.AssetMutationResult{}, fmt.Errorf("asset type no longer exists")
		case errors.Is(err, recorderr.ErrAssetTypeInactive):
			return dto.AssetMutationResult{}, fmt.Errorf("inactive asset type cannot be used for a new asset")
		default:
			return dto.AssetMutationResult{}, fmt.Errorf("create asset: %w", err)
		}
	}

	a.logger.Info("created asset", "asset_id", result.ID, "name", input.Name, "asset_type_id", input.AssetTypeID)
	return result, nil
}

func (a *App) UpdateAsset(input dto.UpdateAssetInput) (dto.AssetMutationResult, error) {
	if err := a.assetValidator.ValidateUpdateAssetInput(input); err != nil {
		return dto.AssetMutationResult{}, err
	}

	result, err := a.assetManagementService.UpdateAsset(a.ctx, input)
	if err != nil {
		a.logger.Error("update asset", "asset_id", input.ID, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrAssetNotFound):
			return dto.AssetMutationResult{}, fmt.Errorf("asset no longer exists")
		case errors.Is(err, recorderr.ErrAssetNameExists):
			return dto.AssetMutationResult{}, fmt.Errorf("asset name already exists in this asset type")
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return dto.AssetMutationResult{}, fmt.Errorf("asset type no longer exists")
		case errors.Is(err, recorderr.ErrAssetTypeInactive):
			return dto.AssetMutationResult{}, fmt.Errorf("inactive asset type cannot be assigned to this asset")
		default:
			return dto.AssetMutationResult{}, fmt.Errorf("update asset: %w", err)
		}
	}

	a.logger.Info("updated asset", "asset_id", input.ID, "asset_type_id", input.AssetTypeID)
	return result, nil
}

func (a *App) ReorderAssetTypes(input dto.ReorderAssetTypesInput) error {
	if err := a.assetValidator.ValidateReorderAssetTypesInput(input); err != nil {
		return err
	}

	if err := a.assetManagementService.ReorderAssetTypes(a.ctx, input); err != nil {
		a.logger.Error("reorder asset types", "count", len(input.OrderedIDs), "active_only", input.ActiveOnly, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return fmt.Errorf("one or more asset types no longer exist")
		default:
			return fmt.Errorf("reorder asset types: %w", err)
		}
	}

	a.logger.Info("reordered asset types", "count", len(input.OrderedIDs), "active_only", input.ActiveOnly)
	return nil
}

func (a *App) ReorderAssets(input dto.ReorderAssetInput) error {
	if err := a.assetValidator.ValidateReorderAssetInput(input); err != nil {
		return err
	}

	if err := a.assetManagementService.ReorderAssets(a.ctx, input); err != nil {
		a.logger.Error("reorder assets", "asset_type_id", input.AssetTypeID, "count", len(input.OrderedIDs), "active_only", input.ActiveOnly, "err", err)
		switch {
		case errors.Is(err, recorderr.ErrAssetTypeNotFound):
			return fmt.Errorf("asset type no longer exists")
		default:
			return fmt.Errorf("reorder assets: %w", err)
		}
	}

	a.logger.Info("reordered assets", "asset_type_id", input.AssetTypeID, "count", len(input.OrderedIDs), "active_only", input.ActiveOnly)
	return nil
}

func (a *App) SaveSnapshot(input dto.SaveSnapshotInput) (dto.SaveSnapshotResult, error) {
	if err := a.validator.ValidateSaveSnapshotInput(input); err != nil {
		return dto.SaveSnapshotResult{}, err
	}

	result, err := a.recordService.SaveSnapshot(a.ctx, input)
	if err != nil {
		a.logger.Error("save snapshot", "snapshot_id", input.SnapshotID, "snapshot_date", input.SnapshotDate, "err", err)

		switch {
		case errors.Is(err, recorderr.ErrSnapshotDateAlreadyExists):
			return dto.SaveSnapshotResult{}, fmt.Errorf("snapshot date already exists, choose another date")
		case errors.Is(err, recorderr.ErrSnapshotNotFound):
			return dto.SaveSnapshotResult{}, fmt.Errorf("snapshot no longer exists")
		case errors.Is(err, recorderr.ErrSnapshotHasNoRows):
			return dto.SaveSnapshotResult{}, fmt.Errorf("snapshot must contain at least one asset")
		case errors.Is(err, recorderr.ErrAssetUnavailable):
			return dto.SaveSnapshotResult{}, fmt.Errorf("one or more selected assets are no longer available")
		default:
			return dto.SaveSnapshotResult{}, fmt.Errorf("save snapshot: %w", err)
		}
	}

	a.logger.Info(
		"saved snapshot",
		"snapshot_id", input.SnapshotID,
		"snapshot_date", input.SnapshotDate,
		"items", len(input.Items),
		"offset", result.Offset,
	)

	return result, nil
}

func (a *App) CreateSnapshot(input dto.CreateSnapshotInput) (dto.CreateSnapshotResult, error) {
	if err := a.validator.ValidateCreateSnapshotInput(input); err != nil {
		return dto.CreateSnapshotResult{}, err
	}

	result, err := a.recordService.CreateSnapshot(a.ctx, input)
	if err != nil {
		a.logger.Error("create snapshot", "snapshot_date", input.SnapshotDate, "err", err)

		switch {
		case errors.Is(err, recorderr.ErrSnapshotDateAlreadyExists):
			return dto.CreateSnapshotResult{}, fmt.Errorf("snapshot date already exists, choose another date")
		case errors.Is(err, recorderr.ErrSnapshotHasNoRows):
			return dto.CreateSnapshotResult{}, fmt.Errorf("snapshot must contain at least one asset")
		case errors.Is(err, recorderr.ErrAssetUnavailable):
			return dto.CreateSnapshotResult{}, fmt.Errorf("one or more selected assets are no longer available")
		default:
			return dto.CreateSnapshotResult{}, fmt.Errorf("create snapshot: %w", err)
		}
	}

	a.logger.Info(
		"created snapshot",
		"snapshot_date", input.SnapshotDate,
		"items", len(input.Items),
		"offset", result.Offset,
	)

	return result, nil
}

func (a *App) DeleteSnapshot(input dto.DeleteSnapshotInput) (dto.DeleteSnapshotResult, error) {
	if err := a.validator.ValidateDeleteSnapshotInput(input); err != nil {
		return dto.DeleteSnapshotResult{}, err
	}

	result, err := a.recordService.DeleteSnapshot(a.ctx, input)
	if err != nil {
		a.logger.Error("delete snapshot", "snapshot_id", input.SnapshotID, "offset", input.Offset, "err", err)

		switch {
		case errors.Is(err, recorderr.ErrSnapshotNotFound):
			return dto.DeleteSnapshotResult{}, fmt.Errorf("snapshot no longer exists")
		default:
			return dto.DeleteSnapshotResult{}, fmt.Errorf("delete snapshot: %w", err)
		}
	}

	a.logger.Info(
		"deleted snapshot",
		"snapshot_id", input.SnapshotID,
		"offset", input.Offset,
		"next_offset", result.Offset,
		"has_snapshots", result.HasSnapshots,
	)

	return result, nil
}

func (a *App) InsertDemoData() error {
	a.logger.Info("ui action", "action", "insert_demo_data")

	if err := a.demoData.InsertOnboardingData(a.ctx); err != nil {
		a.logger.Error("insert demo data", "err", err)
		return fmt.Errorf("insert demo data: %w", err)
	}

	a.logger.Info("inserted demo onboarding data")
	return nil
}

func (a *App) LogHomeAction(action string, offset int) {
	a.logger.Info("ui action", "action", action, "offset", offset)
}
