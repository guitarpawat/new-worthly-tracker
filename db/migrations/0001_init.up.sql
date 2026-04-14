BEGIN;

CREATE TABLE IF NOT EXISTS asset_types (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    ordering INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS assets (
    id INTEGER PRIMARY KEY,
    asset_type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    broker TEXT NOT NULL DEFAULT '',
    is_cash BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    auto_increment BOOLEAN NOT NULL DEFAULT FALSE,
    ordering INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME,
    FOREIGN KEY (asset_type_id) REFERENCES asset_types(id)
);

CREATE TABLE IF NOT EXISTS record_snapshots (
    id INTEGER PRIMARY KEY,
    record_date DATE NOT NULL,
    deleted_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_record_snapshots_record_date
    ON record_snapshots(record_date)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS record_items (
    id INTEGER PRIMARY KEY,
    snapshot_id INTEGER NOT NULL,
    asset_id INTEGER NOT NULL,
    bought_price REAL NOT NULL DEFAULT 0,
    current_price REAL NOT NULL DEFAULT 0,
    remarks TEXT NOT NULL DEFAULT '',
    deleted_at DATETIME,
    FOREIGN KEY (snapshot_id) REFERENCES record_snapshots(id),
    FOREIGN KEY (asset_id) REFERENCES assets(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_record_items_snapshot_asset
    ON record_items(snapshot_id, asset_id)
    WHERE deleted_at IS NULL;

COMMIT;
