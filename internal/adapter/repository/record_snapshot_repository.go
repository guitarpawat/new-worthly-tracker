package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type RecordSnapshotRepository struct {
	db *sqlx.DB
}

type snapshotRow struct {
	ID         int64          `db:"id"`
	RecordDate sql.NullString `db:"record_date"`
}

type snapshotItemRow struct {
	AssetID           int64          `db:"asset_id"`
	AssetName         string         `db:"asset_name"`
	AssetTypeID       sql.NullInt64  `db:"asset_type_id"`
	AssetTypeName     sql.NullString `db:"asset_type_name"`
	AssetTypeOrdering sql.NullInt64  `db:"asset_type_ordering"`
	AssetOrdering     int            `db:"asset_ordering"`
	Broker            string         `db:"broker"`
	AssetIsCash       bool           `db:"asset_is_cash"`
	AssetIsActive     bool           `db:"asset_is_active"`
	BoughtPrice       float64        `db:"bought_price"`
	CurrentPrice      float64        `db:"current_price"`
	Remarks           string         `db:"remarks"`
}

type availableAssetRow struct {
	AssetID           int64  `db:"asset_id"`
	AssetName         string `db:"asset_name"`
	AssetTypeID       int64  `db:"asset_type_id"`
	AssetTypeName     string `db:"asset_type_name"`
	AssetTypeOrdering int    `db:"asset_type_ordering"`
	AssetOrdering     int    `db:"asset_ordering"`
	Broker            string `db:"broker"`
	AssetIsCash       bool   `db:"asset_is_cash"`
	AssetIsActive     bool   `db:"asset_is_active"`
}

type autofillAssetRow struct {
	AssetID           int64           `db:"asset_id"`
	AssetName         string          `db:"asset_name"`
	AssetTypeID       int64           `db:"asset_type_id"`
	AssetTypeName     string          `db:"asset_type_name"`
	AssetTypeOrdering int             `db:"asset_type_ordering"`
	AssetOrdering     int             `db:"asset_ordering"`
	Broker            string          `db:"broker"`
	AssetIsCash       bool            `db:"asset_is_cash"`
	AssetIsActive     bool            `db:"asset_is_active"`
	AutoIncrement     float64         `db:"auto_increment"`
	PrevBoughtPrice   sql.NullFloat64 `db:"prev_bought_price"`
	PrevCurrentPrice  sql.NullFloat64 `db:"prev_current_price"`
	PrevRemarks       sql.NullString  `db:"prev_remarks"`
}

func NewRecordSnapshotRepository(db *sqlx.DB) *RecordSnapshotRepository {
	return &RecordSnapshotRepository{db: db}
}

func (r *RecordSnapshotRepository) ListSnapshotOptions(
	ctx context.Context,
) ([]dto.SnapshotOption, error) {
	headers, err := r.listSnapshotHeaders(ctx)
	if err != nil {
		return nil, err
	}

	options := make([]dto.SnapshotOption, 0, len(headers))
	for index, header := range headers {
		if !header.RecordDate.Valid {
			continue
		}

		options = append(options, dto.SnapshotOption{
			Offset: index,
			Label:  formatSnapshotDate(header.RecordDate.String),
		})
	}

	return options, nil
}

func (r *RecordSnapshotRepository) GetSnapshotByOffset(
	ctx context.Context,
	offset int,
) (*dto.Snapshot, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}

	headers, err := r.listSnapshotHeaders(ctx)
	if err != nil {
		return nil, err
	}
	if offset >= len(headers) {
		return nil, nil
	}
	header := headers[offset]

	itemRows := []snapshotItemRow{}
	if err := r.db.SelectContext(ctx, &itemRows, `
		SELECT
			a.id AS asset_id,
			a.name AS asset_name,
			at.id AS asset_type_id,
			at.name AS asset_type_name,
			at.ordering AS asset_type_ordering,
			a.ordering AS asset_ordering,
			a.broker AS broker,
			a.is_cash AS asset_is_cash,
			ri.bought_price AS bought_price,
			ri.current_price AS current_price,
			ri.remarks AS remarks
		FROM record_items ri
		INNER JOIN assets a ON a.id = ri.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		WHERE ri.snapshot_id = ?
		  AND ri.deleted_at IS NULL
		ORDER BY
			COALESCE(at.ordering, 0),
			COALESCE(at.name, ''),
			a.ordering,
			a.name
	`, header.ID); err != nil {
		return nil, fmt.Errorf("select snapshot items: %w", err)
	}

	items := make([]dto.SnapshotItem, 0, len(itemRows))
	for _, row := range itemRows {
		items = append(items, dto.SnapshotItem{
			AssetID:           row.AssetID,
			AssetName:         row.AssetName,
			AssetTypeID:       row.AssetTypeID.Int64,
			AssetTypeName:     row.AssetTypeName.String,
			AssetTypeOrdering: int(row.AssetTypeOrdering.Int64),
			AssetOrdering:     row.AssetOrdering,
			Broker:            row.Broker,
			IsCash:            row.AssetIsCash,
			BoughtPrice:       row.BoughtPrice,
			CurrentPrice:      row.CurrentPrice,
			Remarks:           row.Remarks,
		})
	}

	return &dto.Snapshot{
		ID:         header.ID,
		RecordDate: parseSnapshotDate(header.RecordDate.String),
		Items:      items,
	}, nil
}

func (r *RecordSnapshotRepository) GetEditableSnapshotByOffset(
	ctx context.Context,
	offset int,
) (*dto.EditableSnapshot, error) {
	if offset < 0 {
		return nil, fmt.Errorf("offset must be non-negative")
	}

	headers, err := r.listSnapshotHeaders(ctx)
	if err != nil {
		return nil, err
	}
	if offset >= len(headers) {
		return nil, nil
	}
	header := headers[offset]

	items, err := r.selectSnapshotItems(ctx, r.db, header.ID)
	if err != nil {
		return nil, err
	}

	availableAssets, err := r.listAvailableAssetsExcludingSnapshot(ctx, header.ID)
	if err != nil {
		return nil, err
	}

	return &dto.EditableSnapshot{
		ID:              header.ID,
		RecordDate:      parseSnapshotDate(header.RecordDate.String),
		Items:           items,
		AvailableAssets: availableAssets,
	}, nil
}

func (r *RecordSnapshotRepository) GetNewSnapshotDraft(
	ctx context.Context,
	recordDate time.Time,
) (*dto.EditableSnapshot, error) {
	headers, err := r.listSnapshotHeaders(ctx)
	if err != nil {
		return nil, err
	}

	var latestSnapshotID int64
	if len(headers) > 0 {
		latestSnapshotID = headers[0].ID
	}

	rows := []autofillAssetRow{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT
			a.id AS asset_id,
			a.name AS asset_name,
			at.id AS asset_type_id,
			at.name AS asset_type_name,
			at.ordering AS asset_type_ordering,
			a.ordering AS asset_ordering,
			a.broker AS broker,
			a.is_cash AS asset_is_cash,
			a.is_active AS asset_is_active,
			a.auto_increment AS auto_increment,
			ri.bought_price AS prev_bought_price,
			ri.current_price AS prev_current_price,
			ri.remarks AS prev_remarks
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		LEFT JOIN record_items ri ON ri.asset_id = a.id
		  AND ri.snapshot_id = ?
		  AND ri.deleted_at IS NULL
		WHERE a.deleted_at IS NULL
		  AND at.deleted_at IS NULL
		  AND at.is_active = TRUE
		  AND a.is_active = TRUE
		ORDER BY
			at.ordering,
			at.name,
			a.ordering,
			a.name
	`, latestSnapshotID); err != nil {
		return nil, fmt.Errorf("select new snapshot autofill assets: %w", err)
	}

	items := make([]dto.EditableSnapshotItem, 0, len(rows))
	excludedAssetIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		boughtPrice := 0.0
		currentPrice := 0.0
		remarks := ""
		if row.PrevBoughtPrice.Valid {
			boughtPrice = row.PrevBoughtPrice.Float64
		}
		if row.PrevCurrentPrice.Valid {
			currentPrice = row.PrevCurrentPrice.Float64
		}
		if row.PrevRemarks.Valid {
			remarks = row.PrevRemarks.String
		}
		if !row.AssetIsCash {
			boughtPrice += row.AutoIncrement
		}
		if row.AssetIsCash {
			boughtPrice = 0
		}

		items = append(items, dto.EditableSnapshotItem{
			AssetID:           row.AssetID,
			AssetName:         row.AssetName,
			AssetTypeID:       row.AssetTypeID,
			AssetTypeName:     row.AssetTypeName,
			AssetTypeOrdering: row.AssetTypeOrdering,
			AssetOrdering:     row.AssetOrdering,
			Broker:            row.Broker,
			IsCash:            row.AssetIsCash,
			IsActive:          row.AssetIsActive,
			BoughtPrice:       boughtPrice,
			CurrentPrice:      currentPrice,
			Remarks:           remarks,
		})
		excludedAssetIDs = append(excludedAssetIDs, row.AssetID)
	}

	availableAssets, err := r.listAvailableAssetsExcludingAssetIDs(ctx, excludedAssetIDs)
	if err != nil {
		return nil, err
	}

	return &dto.EditableSnapshot{
		ID:              0,
		RecordDate:      recordDate,
		Items:           items,
		AvailableAssets: availableAssets,
	}, nil
}

func (r *RecordSnapshotRepository) GetSnapshotOffsetByID(
	ctx context.Context,
	snapshotID int64,
) (int, error) {
	headers, err := r.listSnapshotHeaders(ctx)
	if err != nil {
		return 0, err
	}

	for offset, header := range headers {
		if header.ID == snapshotID {
			return offset, nil
		}
	}

	return 0, recorderr.ErrSnapshotNotFound
}

func (r *RecordSnapshotRepository) listAvailableAssetsExcludingSnapshot(
	ctx context.Context,
	snapshotID int64,
) ([]dto.EditableAssetOption, error) {
	rows := []availableAssetRow{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT
			a.id AS asset_id,
			a.name AS asset_name,
			at.id AS asset_type_id,
			at.name AS asset_type_name,
			at.ordering AS asset_type_ordering,
			a.ordering AS asset_ordering,
			a.broker AS broker,
			a.is_cash AS asset_is_cash,
			a.is_active AS asset_is_active
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		WHERE a.deleted_at IS NULL
		  AND at.deleted_at IS NULL
		  AND at.is_active = TRUE
		  AND a.is_active = TRUE
		  AND NOT EXISTS (
			  SELECT 1
			  FROM record_items ri
			  WHERE ri.snapshot_id = ?
			    AND ri.asset_id = a.id
			    AND ri.deleted_at IS NULL
		  )
		ORDER BY
			at.ordering,
			at.name,
			a.ordering,
			a.name
	`, snapshotID); err != nil {
		return nil, fmt.Errorf("select available assets: %w", err)
	}

	return mapAvailableAssetRows(rows), nil
}

func (r *RecordSnapshotRepository) listAvailableAssetsExcludingAssetIDs(
	ctx context.Context,
	excludedAssetIDs []int64,
) ([]dto.EditableAssetOption, error) {
	query := `
		SELECT
			a.id AS asset_id,
			a.name AS asset_name,
			at.id AS asset_type_id,
			at.name AS asset_type_name,
			at.ordering AS asset_type_ordering,
			a.ordering AS asset_ordering,
			a.broker AS broker,
			a.is_cash AS asset_is_cash,
			a.is_active AS asset_is_active
		FROM assets a
		INNER JOIN asset_types at ON at.id = a.asset_type_id
		WHERE a.deleted_at IS NULL
		  AND at.deleted_at IS NULL
		  AND at.is_active = TRUE
		  AND a.is_active = TRUE
	`
	args := []any{}
	if len(excludedAssetIDs) > 0 {
		notInQuery, notInArgs, err := sqlx.In(`AND a.id NOT IN (?)`, excludedAssetIDs)
		if err != nil {
			return nil, fmt.Errorf("build available assets exclusion query: %w", err)
		}
		query += "\n" + notInQuery
		args = append(args, notInArgs...)
	}
	query += `
		ORDER BY
			at.ordering,
			at.name,
			a.ordering,
			a.name
	`
	query = r.db.Rebind(query)

	rows := []availableAssetRow{}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("select available assets: %w", err)
	}

	return mapAvailableAssetRows(rows), nil
}

func (r *RecordSnapshotRepository) selectSnapshotItems(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	snapshotID int64,
) ([]dto.EditableSnapshotItem, error) {
	itemRows := []snapshotItemRow{}
	if err := sqlx.SelectContext(ctx, queryer, &itemRows, `
		SELECT
			a.id AS asset_id,
			a.name AS asset_name,
			at.id AS asset_type_id,
			at.name AS asset_type_name,
			at.ordering AS asset_type_ordering,
			a.ordering AS asset_ordering,
			a.broker AS broker,
			a.is_cash AS asset_is_cash,
			a.is_active AS asset_is_active,
			ri.bought_price AS bought_price,
			ri.current_price AS current_price,
			ri.remarks AS remarks
		FROM record_items ri
		INNER JOIN assets a ON a.id = ri.asset_id
		LEFT JOIN asset_types at ON at.id = a.asset_type_id
		WHERE ri.snapshot_id = ?
		  AND ri.deleted_at IS NULL
		ORDER BY
			COALESCE(at.ordering, 0),
			COALESCE(at.name, ''),
			a.ordering,
			a.name
	`, snapshotID); err != nil {
		return nil, fmt.Errorf("select snapshot items: %w", err)
	}

	items := make([]dto.EditableSnapshotItem, 0, len(itemRows))
	for _, row := range itemRows {
		items = append(items, dto.EditableSnapshotItem{
			AssetID:           row.AssetID,
			AssetName:         row.AssetName,
			AssetTypeID:       row.AssetTypeID.Int64,
			AssetTypeName:     row.AssetTypeName.String,
			AssetTypeOrdering: int(row.AssetTypeOrdering.Int64),
			AssetOrdering:     row.AssetOrdering,
			Broker:            row.Broker,
			IsCash:            row.AssetIsCash,
			IsActive:          row.AssetIsActive,
			BoughtPrice:       row.BoughtPrice,
			CurrentPrice:      row.CurrentPrice,
			Remarks:           row.Remarks,
		})
	}

	return items, nil
}

func mapAvailableAssetRows(rows []availableAssetRow) []dto.EditableAssetOption {
	availableAssets := make([]dto.EditableAssetOption, 0, len(rows))
	for _, row := range rows {
		availableAssets = append(availableAssets, dto.EditableAssetOption{
			AssetID:           row.AssetID,
			AssetName:         row.AssetName,
			AssetTypeID:       row.AssetTypeID,
			AssetTypeName:     row.AssetTypeName,
			AssetTypeOrdering: row.AssetTypeOrdering,
			AssetOrdering:     row.AssetOrdering,
			Broker:            row.Broker,
			IsCash:            row.AssetIsCash,
			IsActive:          row.AssetIsActive,
		})
	}

	return availableAssets
}

func (r *RecordSnapshotRepository) listSnapshotAssetIDs(
	ctx context.Context,
	queryer sqlx.QueryerContext,
	snapshotID int64,
) ([]int64, error) {
	assetIDs := []int64{}
	if err := sqlx.SelectContext(ctx, queryer, &assetIDs, `
		SELECT asset_id
		FROM record_items
		WHERE snapshot_id = ?
		  AND deleted_at IS NULL
	`, snapshotID); err != nil {
		return nil, fmt.Errorf("list snapshot asset ids: %w", err)
	}

	return assetIDs, nil
}

func (r *RecordSnapshotRepository) listSnapshotHeaders(
	ctx context.Context,
) ([]snapshotRow, error) {
	headers := []snapshotRow{}
	if err := r.db.SelectContext(ctx, &headers, `
		SELECT rs.id, CAST(rs.record_date AS TEXT) AS record_date
		FROM record_snapshots rs
		WHERE rs.deleted_at IS NULL
		  AND EXISTS (
			  SELECT 1
			  FROM record_items ri
			  WHERE ri.snapshot_id = rs.id
			    AND ri.deleted_at IS NULL
		  )
		ORDER BY rs.record_date DESC, rs.id DESC
	`); err != nil {
		return nil, fmt.Errorf("list snapshot headers: %w", err)
	}

	return headers, nil
}

func parseSnapshotDate(raw string) time.Time {
	recordDate, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}
	}

	return recordDate
}

func formatSnapshotDate(raw string) string {
	return parseSnapshotDate(raw).Format("02 Jan 2006")
}

func isSQLiteUniqueError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "UNIQUE constraint failed") ||
		strings.Contains(message, "constraint failed")
}
