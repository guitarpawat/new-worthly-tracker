package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/jmoiron/sqlx"
)

type CustomAllocationRepository struct {
	db *sqlx.DB
}

func NewCustomAllocationRepository(db *sqlx.DB) *CustomAllocationRepository {
	return &CustomAllocationRepository{db: db}
}

func (r *CustomAllocationRepository) ListCharts(ctx context.Context) ([]dto.CustomAllocationChart, error) {
	// One read transaction keeps chart names and their entries consistent.
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin allocation read: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	charts := []dto.CustomAllocationChart{}
	rows := []struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}{}
	if err := tx.SelectContext(ctx, &rows,
		`SELECT id, name FROM custom_allocation_charts WHERE deleted_at IS NULL ORDER BY name_key, id`); err != nil {
		return nil, fmt.Errorf("list allocation charts: %w", err)
	}
	for _, row := range rows {
		entries := []dto.CustomAllocationEntry{}
		if err := tx.SelectContext(ctx, &entries, `
			SELECT asset_id, category, percentage FROM custom_allocation_entries
			WHERE chart_id = ? AND deleted_at IS NULL ORDER BY asset_id, id`, row.ID); err != nil {
			return nil, fmt.Errorf("list allocation entries: %w", err)
		}
		excluded := []int64{}
		if err := tx.SelectContext(ctx, &excluded, `SELECT asset_id FROM custom_allocation_exclusions
			WHERE chart_id = ? AND deleted_at IS NULL ORDER BY asset_id`, row.ID); err != nil {
			return nil, fmt.Errorf("list chart exclusions: %w", err)
		}
		charts = append(charts, dto.CustomAllocationChart{ID: row.ID, Name: row.Name, Entries: entries, ExcludedAssetIDs: excluded})
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("finish allocation read: %w", err)
	}
	return charts, nil
}

func (r *CustomAllocationRepository) SaveChart(ctx context.Context, input dto.CustomAllocationChart) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin allocation save: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var duplicate bool
	if err := tx.GetContext(ctx, &duplicate, `SELECT EXISTS (
		SELECT 1 FROM custom_allocation_charts WHERE name_key = ? AND id <> ? AND deleted_at IS NULL
	)`, strings.ToLower(input.Name), input.ID); err != nil {
		return 0, fmt.Errorf("check allocation name: %w", err)
	}
	if duplicate {
		return 0, fmt.Errorf("chart name already exists")
	}
	if input.ID == 0 {
		result, err := tx.ExecContext(ctx,
			`INSERT INTO custom_allocation_charts(name, name_key) VALUES (?, ?)`, input.Name, strings.ToLower(input.Name))
		if err != nil {
			return 0, fmt.Errorf("create allocation chart: %w", err)
		}
		input.ID, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("allocation chart id: %w", err)
		}
	} else {
		result, err := tx.ExecContext(ctx, `UPDATE custom_allocation_charts SET name = ?, name_key = ?
			WHERE id = ? AND deleted_at IS NULL`, input.Name, strings.ToLower(input.Name), input.ID)
		if err != nil {
			return 0, fmt.Errorf("update allocation chart: %w", err)
		}
		count, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, fmt.Errorf("chart no longer exists")
		}
	}
	if input.Entries != nil {
		// Keep removed assignments as soft-deleted rows, and replace the active set atomically.
		if _, err := tx.ExecContext(ctx, `UPDATE custom_allocation_entries SET deleted_at = CURRENT_TIMESTAMP
		WHERE chart_id = ? AND deleted_at IS NULL`, input.ID); err != nil {
			return 0, fmt.Errorf("replace allocation entries: %w", err)
		}
	}
	for _, entry := range input.Entries {
		var available bool
		if err := tx.GetContext(ctx, &available, `SELECT EXISTS (
			SELECT 1 FROM assets a JOIN asset_types at ON at.id = a.asset_type_id
			WHERE a.id = ? AND a.deleted_at IS NULL AND at.deleted_at IS NULL
		)`, entry.AssetID); err != nil {
			return 0, fmt.Errorf("check allocation asset: %w", err)
		}
		if !available {
			return 0, fmt.Errorf("an allocation asset is no longer available; reload before saving")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO custom_allocation_entries
			(chart_id, asset_id, category, category_key, percentage) VALUES (?, ?, ?, ?, ?)`,
			input.ID, entry.AssetID, entry.Category, strings.ToLower(entry.Category), entry.Percentage); err != nil {
			return 0, fmt.Errorf("save allocation entry: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit allocation chart: %w", err)
	}
	return input.ID, nil
}

func (r *CustomAllocationRepository) DeleteChart(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE custom_allocation_charts SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ? AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("delete allocation chart: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("chart no longer exists")
	}
	return nil
}
