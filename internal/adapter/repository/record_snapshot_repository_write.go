package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type saveAssetMetaRow struct {
	ID               int64        `db:"id"`
	IsCash           bool         `db:"is_cash"`
	IsActive         bool         `db:"is_active"`
	AssetDeletedAt   sql.NullTime `db:"asset_deleted_at"`
	AssetTypeDeleted sql.NullTime `db:"asset_type_deleted_at"`
	AssetTypeActive  bool         `db:"asset_type_is_active"`
}

func (r *RecordSnapshotRepository) SaveSnapshot(
	ctx context.Context,
	input dto.SaveSnapshotInput,
) (dto.SaveSnapshotResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return dto.SaveSnapshotResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var snapshotExists bool
	if err := tx.GetContext(ctx, &snapshotExists, `
		SELECT EXISTS(
			SELECT 1
			FROM record_snapshots
			WHERE id = ?
			  AND deleted_at IS NULL
		)
	`, input.SnapshotID); err != nil {
		return dto.SaveSnapshotResult{}, fmt.Errorf("check snapshot exists: %w", err)
	}
	if !snapshotExists {
		return dto.SaveSnapshotResult{}, recorderr.ErrSnapshotNotFound
	}

	latestSnapshotID, err := r.findLatestSnapshotID(ctx, tx)
	if err != nil {
		return dto.SaveSnapshotResult{}, err
	}
	wasLatestSnapshot := latestSnapshotID == input.SnapshotID

	var duplicateExists bool
	if err := tx.GetContext(ctx, &duplicateExists, `
		SELECT EXISTS(
			SELECT 1
			FROM record_snapshots
			WHERE record_date = ?
			  AND id <> ?
			  AND deleted_at IS NULL
		)
	`, input.SnapshotDate, input.SnapshotID); err != nil {
		return dto.SaveSnapshotResult{}, fmt.Errorf("check duplicate snapshot date: %w", err)
	}
	if duplicateExists {
		return dto.SaveSnapshotResult{}, recorderr.ErrSnapshotDateAlreadyExists
	}

	existingAssetIDs, err := r.listSnapshotAssetIDs(ctx, tx, input.SnapshotID)
	if err != nil {
		return dto.SaveSnapshotResult{}, err
	}
	existingAssetSet := make(map[int64]struct{}, len(existingAssetIDs))
	for _, assetID := range existingAssetIDs {
		existingAssetSet[assetID] = struct{}{}
	}

	assetIDs := make([]int64, 0, len(input.Items))
	for _, item := range input.Items {
		assetIDs = append(assetIDs, item.AssetID)
	}

	assetMetaByID, err := r.loadSaveAssetMeta(ctx, tx, assetIDs)
	if err != nil {
		return dto.SaveSnapshotResult{}, err
	}

	for _, item := range input.Items {
		meta, found := assetMetaByID[item.AssetID]
		if !found {
			return dto.SaveSnapshotResult{}, recorderr.ErrAssetUnavailable
		}

		if _, existed := existingAssetSet[item.AssetID]; !existed &&
			(meta.AssetDeletedAt.Valid || meta.AssetTypeDeleted.Valid || !meta.IsActive || !meta.AssetTypeActive) {
			return dto.SaveSnapshotResult{}, recorderr.ErrAssetUnavailable
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE record_snapshots
		SET record_date = ?
		WHERE id = ?
		  AND deleted_at IS NULL
	`, input.SnapshotDate, input.SnapshotID); err != nil {
		if isSQLiteUniqueError(err) {
			return dto.SaveSnapshotResult{}, recorderr.ErrSnapshotDateAlreadyExists
		}
		return dto.SaveSnapshotResult{}, fmt.Errorf("update snapshot date: %w", err)
	}

	if err := r.softDeleteRemovedItems(ctx, tx, input.SnapshotID, assetIDs); err != nil {
		return dto.SaveSnapshotResult{}, err
	}

	removedAssetIDs := diffAssetIDs(existingAssetIDs, assetIDs)

	for _, item := range input.Items {
		meta := assetMetaByID[item.AssetID]
		boughtPrice := item.BoughtPrice
		if meta.IsCash {
			boughtPrice = 0
		}

		result, err := tx.ExecContext(ctx, `
			UPDATE record_items
			SET bought_price = ?,
			    current_price = ?,
			    remarks = ?
			WHERE snapshot_id = ?
			  AND asset_id = ?
			  AND deleted_at IS NULL
		`, boughtPrice, item.CurrentPrice, item.Remarks, input.SnapshotID, item.AssetID)
		if err != nil {
			return dto.SaveSnapshotResult{}, fmt.Errorf("update record item: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return dto.SaveSnapshotResult{}, fmt.Errorf("record rows affected: %w", err)
		}
		if rowsAffected > 0 {
			continue
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO record_items (
				snapshot_id,
				asset_id,
				bought_price,
				current_price,
				remarks
			) VALUES (?, ?, ?, ?, ?)
		`, input.SnapshotID, item.AssetID, boughtPrice, item.CurrentPrice, item.Remarks); err != nil {
			return dto.SaveSnapshotResult{}, fmt.Errorf("insert record item: %w", err)
		}
	}

	if wasLatestSnapshot {
		willRemainLatest, err := r.willSnapshotBeLatest(ctx, tx, input.SnapshotID, input.SnapshotDate)
		if err != nil {
			return dto.SaveSnapshotResult{}, err
		}
		if willRemainLatest {
			if err := r.deactivateAssets(ctx, tx, removedAssetIDs); err != nil {
				return dto.SaveSnapshotResult{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.SaveSnapshotResult{}, fmt.Errorf("commit tx: %w", err)
	}

	offset, err := r.GetSnapshotOffsetByID(ctx, input.SnapshotID)
	if err != nil {
		return dto.SaveSnapshotResult{}, err
	}

	return dto.SaveSnapshotResult{Offset: offset}, nil
}

func (r *RecordSnapshotRepository) CreateSnapshot(
	ctx context.Context,
	input dto.CreateSnapshotInput,
) (dto.CreateSnapshotResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return dto.CreateSnapshotResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var duplicateExists bool
	if err := tx.GetContext(ctx, &duplicateExists, `
		SELECT EXISTS(
			SELECT 1
			FROM record_snapshots
			WHERE record_date = ?
			  AND deleted_at IS NULL
		)
	`, input.SnapshotDate); err != nil {
		return dto.CreateSnapshotResult{}, fmt.Errorf("check duplicate snapshot date: %w", err)
	}
	if duplicateExists {
		return dto.CreateSnapshotResult{}, recorderr.ErrSnapshotDateAlreadyExists
	}

	latestSnapshotID, err := r.findLatestSnapshotID(ctx, tx)
	if err != nil {
		return dto.CreateSnapshotResult{}, err
	}
	shouldBecomeLatest, err := r.willSnapshotBeLatest(ctx, tx, 0, input.SnapshotDate)
	if err != nil {
		return dto.CreateSnapshotResult{}, err
	}

	assetIDs := make([]int64, 0, len(input.Items))
	for _, item := range input.Items {
		assetIDs = append(assetIDs, item.AssetID)
	}

	assetMetaByID, err := r.loadSaveAssetMeta(ctx, tx, assetIDs)
	if err != nil {
		return dto.CreateSnapshotResult{}, err
	}
	for _, item := range input.Items {
		meta, found := assetMetaByID[item.AssetID]
		if !found || meta.AssetDeletedAt.Valid || meta.AssetTypeDeleted.Valid || !meta.IsActive || !meta.AssetTypeActive {
			return dto.CreateSnapshotResult{}, recorderr.ErrAssetUnavailable
		}
	}

	removedFromPreviousLatest := []int64{}
	if shouldBecomeLatest && latestSnapshotID > 0 {
		latestAssetIDs, err := r.listSnapshotAssetIDs(ctx, tx, latestSnapshotID)
		if err != nil {
			return dto.CreateSnapshotResult{}, err
		}
		removedFromPreviousLatest = diffAssetIDs(latestAssetIDs, assetIDs)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO record_snapshots (record_date)
		VALUES (?)
	`, input.SnapshotDate)
	if err != nil {
		if isSQLiteUniqueError(err) {
			return dto.CreateSnapshotResult{}, recorderr.ErrSnapshotDateAlreadyExists
		}
		return dto.CreateSnapshotResult{}, fmt.Errorf("insert snapshot: %w", err)
	}

	snapshotID, err := result.LastInsertId()
	if err != nil {
		return dto.CreateSnapshotResult{}, fmt.Errorf("snapshot last insert id: %w", err)
	}

	for _, item := range input.Items {
		meta := assetMetaByID[item.AssetID]
		boughtPrice := item.BoughtPrice
		if meta.IsCash {
			boughtPrice = 0
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO record_items (
				snapshot_id,
				asset_id,
				bought_price,
				current_price,
				remarks
			) VALUES (?, ?, ?, ?, ?)
		`, snapshotID, item.AssetID, boughtPrice, item.CurrentPrice, item.Remarks); err != nil {
			return dto.CreateSnapshotResult{}, fmt.Errorf("insert record item: %w", err)
		}
	}

	if shouldBecomeLatest {
		if err := r.deactivateAssets(ctx, tx, removedFromPreviousLatest); err != nil {
			return dto.CreateSnapshotResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return dto.CreateSnapshotResult{}, fmt.Errorf("commit tx: %w", err)
	}

	offset, err := r.GetSnapshotOffsetByID(ctx, snapshotID)
	if err != nil {
		return dto.CreateSnapshotResult{}, err
	}

	return dto.CreateSnapshotResult{Offset: offset}, nil
}

func (r *RecordSnapshotRepository) DeleteSnapshot(
	ctx context.Context,
	input dto.DeleteSnapshotInput,
) (dto.DeleteSnapshotResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var snapshotExists bool
	if err := tx.GetContext(ctx, &snapshotExists, `
		SELECT EXISTS(
			SELECT 1
			FROM record_snapshots
			WHERE id = ?
			  AND deleted_at IS NULL
		)
	`, input.SnapshotID); err != nil {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("check snapshot exists: %w", err)
	}
	if !snapshotExists {
		return dto.DeleteSnapshotResult{}, recorderr.ErrSnapshotNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE record_items
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE snapshot_id = ?
		  AND deleted_at IS NULL
	`, input.SnapshotID); err != nil {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("soft delete record items: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE record_snapshots
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND deleted_at IS NULL
	`, input.SnapshotID); err != nil {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("soft delete snapshot: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return dto.DeleteSnapshotResult{}, fmt.Errorf("commit tx: %w", err)
	}

	headers, err := r.listSnapshotHeaders(ctx)
	if err != nil {
		return dto.DeleteSnapshotResult{}, err
	}
	if len(headers) == 0 {
		return dto.DeleteSnapshotResult{
			Offset:       0,
			HasSnapshots: false,
		}, nil
	}

	offset := input.Offset
	if offset >= len(headers) {
		offset = len(headers) - 1
	}
	if offset < 0 {
		offset = 0
	}

	return dto.DeleteSnapshotResult{
		Offset:       offset,
		HasSnapshots: true,
	}, nil
}

func (r *RecordSnapshotRepository) loadSaveAssetMeta(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	assetIDs []int64,
) (map[int64]saveAssetMetaRow, error) {
	if len(assetIDs) == 0 {
		return map[int64]saveAssetMetaRow{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT
			a.id,
			a.is_cash,
			a.is_active,
			a.deleted_at AS asset_deleted_at,
			at.deleted_at AS asset_type_deleted_at,
			at.is_active AS asset_type_is_active
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		WHERE a.id IN (?)
	`, assetIDs)
	if err != nil {
		return nil, fmt.Errorf("build asset meta query: %w", err)
	}
	query = r.db.Rebind(query)

	rows := []saveAssetMetaRow{}
	if err := sqlx.SelectContext(ctx, queryer, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select asset meta: %w", err)
	}

	metaByID := make(map[int64]saveAssetMetaRow, len(rows))
	for _, row := range rows {
		metaByID[row.ID] = row
	}

	return metaByID, nil
}

func (r *RecordSnapshotRepository) softDeleteRemovedItems(
	ctx context.Context,
	tx *sqlx.Tx,
	snapshotID int64,
	assetIDs []int64,
) error {
	query := `
		UPDATE record_items
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE snapshot_id = ?
		  AND deleted_at IS NULL
	`
	args := []any{snapshotID}

	if len(assetIDs) > 0 {
		notInQuery, notInArgs, err := sqlx.In(`
			AND asset_id NOT IN (?)
		`, assetIDs)
		if err != nil {
			return fmt.Errorf("build remove items query: %w", err)
		}

		query += notInQuery
		args = append(args, notInArgs...)
	}

	query = tx.Rebind(query)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("soft delete removed items: %w", err)
	}

	return nil
}

func (r *RecordSnapshotRepository) findLatestSnapshotID(
	ctx context.Context,
	queryer sqlx.QueryerContext,
) (int64, error) {
	var snapshotID sql.NullInt64
	if err := sqlx.GetContext(ctx, queryer, &snapshotID, `
		SELECT rs.id
		FROM record_snapshots rs
		WHERE rs.deleted_at IS NULL
		  AND EXISTS (
			  SELECT 1
			  FROM record_items ri
			  WHERE ri.snapshot_id = rs.id
			    AND ri.deleted_at IS NULL
		  )
		ORDER BY rs.record_date DESC, rs.id DESC
		LIMIT 1
	`); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("find latest snapshot id: %w", err)
	}

	if !snapshotID.Valid {
		return 0, nil
	}

	return snapshotID.Int64, nil
}

func (r *RecordSnapshotRepository) willSnapshotBeLatest(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	excludeSnapshotID int64,
	recordDate string,
) (bool, error) {
	var newerExists bool
	if err := sqlx.GetContext(ctx, queryer, &newerExists, `
		SELECT EXISTS(
			SELECT 1
			FROM record_snapshots rs
			WHERE rs.deleted_at IS NULL
			  AND (? = 0 OR rs.id <> ?)
			  AND EXISTS (
				  SELECT 1
				  FROM record_items ri
				  WHERE ri.snapshot_id = rs.id
				    AND ri.deleted_at IS NULL
			  )
			  AND rs.record_date > ?
		)
	`, excludeSnapshotID, excludeSnapshotID, recordDate); err != nil {
		return false, fmt.Errorf("check newer snapshot exists: %w", err)
	}

	return !newerExists, nil
}

func (r *RecordSnapshotRepository) deactivateAssets(
	ctx context.Context,
	tx *sqlx.Tx,
	assetIDs []int64,
) error {
	if len(assetIDs) == 0 {
		return nil
	}

	query, args, err := sqlx.In(`
		UPDATE assets
		SET is_active = FALSE
		WHERE id IN (?)
		  AND deleted_at IS NULL
	`, assetIDs)
	if err != nil {
		return fmt.Errorf("build deactivate assets query: %w", err)
	}
	query = tx.Rebind(query)

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("deactivate removed latest assets: %w", err)
	}

	return nil
}

func diffAssetIDs(existingAssetIDs []int64, keptAssetIDs []int64) []int64 {
	if len(existingAssetIDs) == 0 {
		return nil
	}

	kept := make(map[int64]struct{}, len(keptAssetIDs))
	for _, assetID := range keptAssetIDs {
		kept[assetID] = struct{}{}
	}

	removed := make([]int64, 0, len(existingAssetIDs))
	for _, assetID := range existingAssetIDs {
		if _, found := kept[assetID]; found {
			continue
		}
		removed = append(removed, assetID)
	}

	return removed
}
