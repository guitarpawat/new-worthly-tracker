package repository

import (
	"context"
	"fmt"
)

func (r *AssetManagementRepository) DeleteAsset(ctx context.Context, id int64) error {
	// Count archived records too: deleting a snapshot must not enable deleting its assets.
	result, err := r.db.ExecContext(ctx, `UPDATE assets SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
		AND NOT EXISTS (SELECT 1 FROM record_items WHERE asset_id = assets.id)`, id)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("asset cannot be deleted: it has snapshot records or no longer exists")
	}
	return nil
}

func (r *AssetManagementRepository) DeleteAssetType(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE asset_types SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL
		AND NOT EXISTS (SELECT 1 FROM assets WHERE asset_type_id = asset_types.id AND deleted_at IS NULL)`, id)
	if err != nil {
		return fmt.Errorf("delete asset type: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("type cannot be deleted: it contains assets or no longer exists")
	}
	return nil
}
