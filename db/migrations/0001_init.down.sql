BEGIN;

DROP INDEX IF EXISTS idx_record_snapshots_record_date;
DROP INDEX IF EXISTS idx_record_items_snapshot_asset;
DROP TABLE IF EXISTS record_items;
DROP TABLE IF EXISTS record_snapshots;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS asset_types;

COMMIT;
