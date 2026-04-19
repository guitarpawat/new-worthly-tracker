package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type AssetManagementRepository struct {
	db *sqlx.DB
}

type assetTypeListRow struct {
	ID         int64  `db:"id"`
	Name       string `db:"name"`
	IsActive   bool   `db:"is_active"`
	Ordering   int    `db:"ordering"`
	AssetCount int    `db:"asset_count"`
}

type assetListRow struct {
	ID            int64   `db:"id"`
	Name          string  `db:"name"`
	AssetTypeID   int64   `db:"asset_type_id"`
	AssetTypeName string  `db:"asset_type_name"`
	Broker        string  `db:"broker"`
	IsCash        bool    `db:"is_cash"`
	IsLiability   bool    `db:"is_liability"`
	IsActive      bool    `db:"is_active"`
	Ordering      int     `db:"ordering"`
	AutoIncrement float64 `db:"auto_increment"`
}

type assetTypeMetaRow struct {
	ID        int64        `db:"id"`
	IsActive  bool         `db:"is_active"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type assetMetaRow struct {
	ID               int64        `db:"id"`
	AssetTypeID      int64        `db:"asset_type_id"`
	IsCash           bool         `db:"is_cash"`
	IsLiability      bool         `db:"is_liability"`
	DeletedAt        sql.NullTime `db:"deleted_at"`
	AssetTypeDeleted sql.NullTime `db:"asset_type_deleted_at"`
	AssetTypeActive  bool         `db:"asset_type_is_active"`
}

func NewAssetManagementRepository(db *sqlx.DB) *AssetManagementRepository {
	return &AssetManagementRepository{db: db}
}

func (r *AssetManagementRepository) GetPage(ctx context.Context) (dto.AssetManagementPage, error) {
	assetTypeRows := []assetTypeListRow{}
	if err := r.db.SelectContext(ctx, &assetTypeRows, `
		SELECT
			at.id,
			at.name,
			at.is_active,
			at.ordering,
			COUNT(a.id) AS asset_count
		FROM asset_types at
		LEFT JOIN assets a ON a.asset_type_id = at.id
		  AND a.deleted_at IS NULL
		WHERE at.deleted_at IS NULL
		GROUP BY at.id, at.name, at.is_active, at.ordering
		ORDER BY at.ordering, at.name
	`); err != nil {
		return dto.AssetManagementPage{}, fmt.Errorf("select asset types: %w", err)
	}

	assetRows := []assetListRow{}
	if err := r.db.SelectContext(ctx, &assetRows, `
		SELECT
			a.id,
			a.name,
			a.asset_type_id,
			at.name AS asset_type_name,
			a.broker,
			a.is_cash,
			a.is_liability,
			a.is_active,
			a.ordering,
			a.auto_increment
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		WHERE a.deleted_at IS NULL
		  AND at.deleted_at IS NULL
		ORDER BY at.ordering, at.name, a.ordering, a.name
	`); err != nil {
		return dto.AssetManagementPage{}, fmt.Errorf("select assets: %w", err)
	}

	activeTypeRows := []dto.AssetTypeOption{}
	if err := r.db.SelectContext(ctx, &activeTypeRows, `
		SELECT
			id,
			name
		FROM asset_types
		WHERE deleted_at IS NULL
		  AND is_active = TRUE
		ORDER BY ordering, name
	`); err != nil {
		return dto.AssetManagementPage{}, fmt.Errorf("select active asset type options: %w", err)
	}

	page := dto.AssetManagementPage{
		AssetTypes:       make([]dto.AssetTypeRow, 0, len(assetTypeRows)),
		Assets:           make([]dto.AssetRow, 0, len(assetRows)),
		ActiveAssetTypes: activeTypeRows,
	}
	for _, row := range assetTypeRows {
		page.AssetTypes = append(page.AssetTypes, dto.AssetTypeRow{
			ID:         row.ID,
			Name:       row.Name,
			IsActive:   row.IsActive,
			Ordering:   row.Ordering,
			AssetCount: row.AssetCount,
		})
	}
	for _, row := range assetRows {
		page.Assets = append(page.Assets, dto.AssetRow{
			ID:            row.ID,
			Name:          row.Name,
			AssetTypeID:   row.AssetTypeID,
			AssetTypeName: row.AssetTypeName,
			Broker:        row.Broker,
			IsCash:        row.IsCash,
			IsLiability:   row.IsLiability,
			IsActive:      row.IsActive,
			Ordering:      row.Ordering,
			AutoIncrement: row.AutoIncrement,
		})
	}

	return page, nil
}

func (r *AssetManagementRepository) CreateAssetType(ctx context.Context, input dto.CreateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	exists, err := r.assetTypeNameExists(ctx, input.Name, 0)
	if err != nil {
		return dto.AssetTypeMutationResult{}, err
	}
	if exists {
		return dto.AssetTypeMutationResult{}, recorderr.ErrAssetTypeNameExists
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO asset_types (name, is_active)
		VALUES (?, ?)
	`, input.Name, input.IsActive)
	if err != nil {
		return dto.AssetTypeMutationResult{}, fmt.Errorf("insert asset type: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return dto.AssetTypeMutationResult{}, fmt.Errorf("asset type last insert id: %w", err)
	}

	return dto.AssetTypeMutationResult{ID: id}, nil
}

func (r *AssetManagementRepository) UpdateAssetType(ctx context.Context, input dto.UpdateAssetTypeInput) (dto.AssetTypeMutationResult, error) {
	exists, err := r.assetTypeNameExists(ctx, input.Name, input.ID)
	if err != nil {
		return dto.AssetTypeMutationResult{}, err
	}
	if exists {
		return dto.AssetTypeMutationResult{}, recorderr.ErrAssetTypeNameExists
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE asset_types
		SET name = ?,
		    is_active = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`, input.Name, input.IsActive, input.ID)
	if err != nil {
		return dto.AssetTypeMutationResult{}, fmt.Errorf("update asset type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dto.AssetTypeMutationResult{}, fmt.Errorf("asset type rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return dto.AssetTypeMutationResult{}, recorderr.ErrAssetTypeNotFound
	}

	return dto.AssetTypeMutationResult{ID: input.ID}, nil
}

func (r *AssetManagementRepository) CreateAsset(ctx context.Context, input dto.CreateAssetInput) (dto.AssetMutationResult, error) {
	typeMeta, err := r.loadAssetTypeMeta(ctx, r.db, input.AssetTypeID)
	if err != nil {
		return dto.AssetMutationResult{}, err
	}
	if typeMeta.DeletedAt.Valid || !typeMeta.IsActive {
		return dto.AssetMutationResult{}, recorderr.ErrAssetTypeInactive
	}
	exists, err := r.assetNameExistsInType(ctx, input.AssetTypeID, input.Name, 0)
	if err != nil {
		return dto.AssetMutationResult{}, err
	}
	if exists {
		return dto.AssetMutationResult{}, recorderr.ErrAssetNameExists
	}

	autoIncrement := input.AutoIncrement
	if input.IsCash {
		autoIncrement = 0
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO assets (
			asset_type_id,
			name,
			broker,
			is_cash,
			is_liability,
			is_active,
			auto_increment
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.AssetTypeID, input.Name, input.Broker, input.IsCash, input.IsLiability, input.IsActive, autoIncrement)
	if err != nil {
		return dto.AssetMutationResult{}, fmt.Errorf("insert asset: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return dto.AssetMutationResult{}, fmt.Errorf("asset last insert id: %w", err)
	}

	return dto.AssetMutationResult{ID: id}, nil
}

func (r *AssetManagementRepository) UpdateAsset(ctx context.Context, input dto.UpdateAssetInput) (dto.AssetMutationResult, error) {
	assetMeta, err := r.loadAssetMeta(ctx, r.db, input.ID)
	if err != nil {
		return dto.AssetMutationResult{}, err
	}
	if assetMeta.DeletedAt.Valid || assetMeta.AssetTypeDeleted.Valid {
		return dto.AssetMutationResult{}, recorderr.ErrAssetNotFound
	}

	typeMeta, err := r.loadAssetTypeMeta(ctx, r.db, input.AssetTypeID)
	if err != nil {
		return dto.AssetMutationResult{}, err
	}
	if typeMeta.DeletedAt.Valid {
		return dto.AssetMutationResult{}, recorderr.ErrAssetTypeNotFound
	}
	if !typeMeta.IsActive && input.AssetTypeID != assetMeta.AssetTypeID {
		return dto.AssetMutationResult{}, recorderr.ErrAssetTypeInactive
	}
	exists, err := r.assetNameExistsInType(ctx, input.AssetTypeID, input.Name, input.ID)
	if err != nil {
		return dto.AssetMutationResult{}, err
	}
	if exists {
		return dto.AssetMutationResult{}, recorderr.ErrAssetNameExists
	}

	autoIncrement := input.AutoIncrement
	if input.IsCash {
		autoIncrement = 0
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE assets
		SET asset_type_id = ?,
		    name = ?,
		    broker = ?,
		    is_cash = ?,
		    is_liability = ?,
		    is_active = ?,
		    auto_increment = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`, input.AssetTypeID, input.Name, input.Broker, input.IsCash, input.IsLiability, input.IsActive, autoIncrement, input.ID)
	if err != nil {
		return dto.AssetMutationResult{}, fmt.Errorf("update asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dto.AssetMutationResult{}, fmt.Errorf("asset rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return dto.AssetMutationResult{}, recorderr.ErrAssetNotFound
	}

	return dto.AssetMutationResult{ID: input.ID}, nil
}

func (r *AssetManagementRepository) ReorderAssetTypes(ctx context.Context, input dto.ReorderAssetTypesInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rows := []assetTypeListRow{}
	if err := sqlx.SelectContext(ctx, tx, &rows, `
		SELECT
			id,
			name,
			is_active,
			ordering
		FROM asset_types
		WHERE deleted_at IS NULL
		ORDER BY ordering, name
	`); err != nil {
		return fmt.Errorf("select asset types for reorder: %w", err)
	}

	finalOrder, err := mergeVisibleOrderIDs(extractVisibleAssetTypeIDs(rows, input.ActiveOnly), extractAllAssetTypeIDs(rows), input.OrderedIDs)
	if err != nil {
		return err
	}

	if err := r.persistAssetTypeOrdering(ctx, tx, finalOrder); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *AssetManagementRepository) ReorderAssets(ctx context.Context, input dto.ReorderAssetInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	typeMeta, err := r.loadAssetTypeMeta(ctx, tx, input.AssetTypeID)
	if err != nil {
		return err
	}
	if typeMeta.DeletedAt.Valid {
		return recorderr.ErrAssetTypeNotFound
	}

	rows := []assetListRow{}
	if err := sqlx.SelectContext(ctx, tx, &rows, `
		SELECT
			a.id,
			a.name,
			a.asset_type_id,
			at.name AS asset_type_name,
			a.broker,
			a.is_cash,
			a.is_liability,
			a.is_active,
			a.ordering,
			a.auto_increment
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		WHERE a.deleted_at IS NULL
		  AND at.deleted_at IS NULL
		  AND a.asset_type_id = ?
		ORDER BY a.ordering, a.name
	`, input.AssetTypeID); err != nil {
		return fmt.Errorf("select assets for reorder: %w", err)
	}

	finalOrder, err := mergeVisibleOrderIDs(extractVisibleAssetIDs(rows, input.ActiveOnly), extractAllAssetIDs(rows), input.OrderedIDs)
	if err != nil {
		return err
	}

	if err := r.persistAssetOrdering(ctx, tx, finalOrder); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *AssetManagementRepository) loadAssetTypeMeta(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	assetTypeID int64,
) (assetTypeMetaRow, error) {
	var row assetTypeMetaRow
	if err := sqlx.GetContext(ctx, queryer, &row, `
		SELECT id, is_active, deleted_at
		FROM asset_types
		WHERE id = ?
	`, assetTypeID); err != nil {
		if err == sql.ErrNoRows {
			return assetTypeMetaRow{}, recorderr.ErrAssetTypeNotFound
		}
		return assetTypeMetaRow{}, fmt.Errorf("select asset type meta: %w", err)
	}

	return row, nil
}

func (r *AssetManagementRepository) assetTypeNameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	var count int
	if err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(1)
		FROM asset_types
		WHERE deleted_at IS NULL
		  AND lower(trim(name)) = lower(trim(?))
		  AND (? = 0 OR id <> ?)
	`, strings.TrimSpace(name), excludeID, excludeID); err != nil {
		return false, fmt.Errorf("check duplicate asset type name: %w", err)
	}

	return count > 0, nil
}

func extractVisibleAssetTypeIDs(rows []assetTypeListRow, activeOnly bool) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if activeOnly && !row.IsActive {
			continue
		}
		ids = append(ids, row.ID)
	}
	return ids
}

func extractAllAssetTypeIDs(rows []assetTypeListRow) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

func extractVisibleAssetIDs(rows []assetListRow, activeOnly bool) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if activeOnly && !row.IsActive {
			continue
		}
		ids = append(ids, row.ID)
	}
	return ids
}

func extractAllAssetIDs(rows []assetListRow) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

func mergeVisibleOrderIDs(visibleIDs []int64, allIDs []int64, reorderedVisibleIDs []int64) ([]int64, error) {
	if len(visibleIDs) != len(reorderedVisibleIDs) {
		return nil, fmt.Errorf("reorder payload does not match visible rows")
	}

	visibleSet := make(map[int64]struct{}, len(visibleIDs))
	for _, id := range visibleIDs {
		visibleSet[id] = struct{}{}
	}

	seen := make(map[int64]struct{}, len(reorderedVisibleIDs))
	for _, id := range reorderedVisibleIDs {
		if _, found := visibleSet[id]; !found {
			return nil, fmt.Errorf("reorder payload contains invalid row")
		}
		if _, found := seen[id]; found {
			return nil, fmt.Errorf("reorder payload contains duplicate row")
		}
		seen[id] = struct{}{}
	}

	finalOrder := make([]int64, 0, len(allIDs))
	visibleIndex := 0
	for _, id := range allIDs {
		if _, found := visibleSet[id]; found {
			finalOrder = append(finalOrder, reorderedVisibleIDs[visibleIndex])
			visibleIndex++
			continue
		}
		finalOrder = append(finalOrder, id)
	}

	return finalOrder, nil
}

func (r *AssetManagementRepository) persistAssetTypeOrdering(ctx context.Context, tx *sqlx.Tx, orderedIDs []int64) error {
	for index, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx, `
			UPDATE asset_types
			SET ordering = ?
			WHERE id = ?
			  AND deleted_at IS NULL
		`, index+1, id); err != nil {
			return fmt.Errorf("update asset type ordering: %w", err)
		}
	}
	return nil
}

func (r *AssetManagementRepository) persistAssetOrdering(ctx context.Context, tx *sqlx.Tx, orderedIDs []int64) error {
	for index, id := range orderedIDs {
		if _, err := tx.ExecContext(ctx, `
			UPDATE assets
			SET ordering = ?
			WHERE id = ?
			  AND deleted_at IS NULL
		`, index+1, id); err != nil {
			return fmt.Errorf("update asset ordering: %w", err)
		}
	}
	return nil
}

func (r *AssetManagementRepository) loadAssetMeta(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	assetID int64,
) (assetMetaRow, error) {
	var row assetMetaRow
	if err := sqlx.GetContext(ctx, queryer, &row, `
		SELECT
			a.id,
			a.asset_type_id,
			a.is_cash,
			a.is_liability,
			a.deleted_at,
			at.deleted_at AS asset_type_deleted_at,
			at.is_active AS asset_type_is_active
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		WHERE a.id = ?
	`, assetID); err != nil {
		if err == sql.ErrNoRows {
			return assetMetaRow{}, recorderr.ErrAssetNotFound
		}
		return assetMetaRow{}, fmt.Errorf("select asset meta: %w", err)
	}

	return row, nil
}

func (r *AssetManagementRepository) assetNameExistsInType(ctx context.Context, assetTypeID int64, name string, excludeID int64) (bool, error) {
	var count int
	if err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(1)
		FROM assets
		WHERE deleted_at IS NULL
		  AND asset_type_id = ?
		  AND lower(trim(name)) = lower(trim(?))
		  AND (? = 0 OR id <> ?)
	`, assetTypeID, strings.TrimSpace(name), excludeID, excludeID); err != nil {
		return false, fmt.Errorf("check duplicate asset name: %w", err)
	}

	return count > 0, nil
}
