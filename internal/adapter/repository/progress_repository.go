package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

type ProgressRepository struct {
	db *sqlx.DB
}

func NewProgressRepository(db *sqlx.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

func (r *ProgressRepository) ListSnapshotDates(ctx context.Context) ([]string, error) {
	dates := []string{}
	if err := r.db.SelectContext(ctx, &dates, `
		SELECT CAST(rs.record_date AS TEXT) AS record_date
		FROM record_snapshots rs
		WHERE rs.deleted_at IS NULL
		  AND EXISTS (
			  SELECT 1
			  FROM record_items ri
			  WHERE ri.snapshot_id = rs.id
			    AND ri.deleted_at IS NULL
		  )
		ORDER BY rs.record_date ASC, rs.id ASC
	`); err != nil {
		return nil, fmt.Errorf("select snapshot dates: %w", err)
	}

	return dates, nil
}

func (r *ProgressRepository) ListSnapshotItemsInRange(
	ctx context.Context,
	startDate string,
	endDate string,
) ([]dto.ProgressSnapshotItem, error) {
	rows := []dto.ProgressSnapshotItem{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT
			rs.id AS snapshot_id,
			CAST(rs.record_date AS TEXT) AS snapshot_date,
			a.name AS asset_name,
			COALESCE(at.name, '') AS asset_type_name,
			ri.current_price,
			ri.bought_price,
			a.is_cash AS is_cash
		FROM record_snapshots rs
		INNER JOIN record_items ri ON ri.snapshot_id = rs.id
		INNER JOIN assets a ON a.id = ri.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		WHERE rs.deleted_at IS NULL
		  AND ri.deleted_at IS NULL
		  AND rs.record_date >= ?
		  AND rs.record_date <= ?
		ORDER BY rs.record_date ASC, rs.id ASC, COALESCE(at.ordering, 0), a.ordering, a.name
	`, startDate, endDate); err != nil {
		return nil, fmt.Errorf("select progress snapshot items: %w", err)
	}

	return rows, nil
}
