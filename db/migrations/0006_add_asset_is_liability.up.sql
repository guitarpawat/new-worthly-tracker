BEGIN;

ALTER TABLE assets
    ADD COLUMN is_liability BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE assets
SET is_liability = TRUE
WHERE id IN (
    SELECT DISTINCT ri.asset_id
    FROM record_items ri
    INNER JOIN record_snapshots rs ON rs.id = ri.snapshot_id
    WHERE ri.deleted_at IS NULL
      AND rs.deleted_at IS NULL
      AND ri.current_price < 0
      AND rs.record_date = (
          SELECT MAX(rs2.record_date)
          FROM record_snapshots rs2
          WHERE rs2.deleted_at IS NULL
      )
);

COMMIT;
