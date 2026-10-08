package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/jmoiron/sqlx"
)

// Save only the edited asset's assignments, inside the asset update transaction.
func saveAssetChartAllocations(ctx context.Context, tx *sqlx.Tx, input dto.UpdateAssetInput) error {
	for _, chart := range input.ChartAllocations {
		var available bool
		if err := tx.GetContext(ctx, &available, `SELECT EXISTS (
			SELECT 1 FROM custom_allocation_charts WHERE id = ? AND deleted_at IS NULL
		)`, chart.ChartID); err != nil {
			return fmt.Errorf("check allocation chart: %w", err)
		}
		if !available {
			return fmt.Errorf("an allocation chart no longer exists; reload before saving")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE custom_allocation_exclusions SET deleted_at = CURRENT_TIMESTAMP
			WHERE chart_id = ? AND asset_id = ? AND deleted_at IS NULL`, chart.ChartID, input.ID); err != nil {
			return fmt.Errorf("update chart inclusion: %w", err)
		}
		if chart.Excluded {
			if _, err := tx.ExecContext(ctx, `INSERT INTO custom_allocation_exclusions (chart_id, asset_id) VALUES (?, ?)`,
				chart.ChartID, input.ID); err != nil {
				return fmt.Errorf("exclude asset from chart: %w", err)
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE custom_allocation_entries SET deleted_at = CURRENT_TIMESTAMP
			WHERE chart_id = ? AND asset_id = ? AND deleted_at IS NULL`, chart.ChartID, input.ID); err != nil {
			return fmt.Errorf("replace asset allocations: %w", err)
		}
		for _, entry := range chart.Entries {
			if _, err := tx.ExecContext(ctx, `INSERT INTO custom_allocation_entries
				(chart_id, asset_id, category, category_key, percentage) VALUES (?, ?, ?, ?, ?)`,
				chart.ChartID, input.ID, entry.Category, strings.ToLower(entry.Category), entry.Percentage); err != nil {
				return fmt.Errorf("save asset allocation: %w", err)
			}
		}
	}
	return nil
}
