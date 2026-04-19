BEGIN;

PRAGMA foreign_keys = OFF;

CREATE TABLE assets_revert_is_liability (
    id INTEGER PRIMARY KEY,
    asset_type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    broker TEXT NOT NULL DEFAULT '',
    is_cash BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    auto_increment REAL NOT NULL DEFAULT 0,
    ordering INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME,
    FOREIGN KEY (asset_type_id) REFERENCES asset_types(id)
);

INSERT INTO assets_revert_is_liability (
    id,
    asset_type_id,
    name,
    broker,
    is_cash,
    is_active,
    auto_increment,
    ordering,
    deleted_at
)
SELECT
    id,
    asset_type_id,
    name,
    broker,
    is_cash,
    is_active,
    auto_increment,
    ordering,
    deleted_at
FROM assets;

DROP TABLE assets;
ALTER TABLE assets_revert_is_liability RENAME TO assets;

PRAGMA foreign_keys = ON;

COMMIT;
