PRAGMA foreign_keys = OFF;

BEGIN;

CREATE TABLE assets_new (
    id INTEGER PRIMARY KEY,
    asset_type_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    broker TEXT NOT NULL DEFAULT '',
    is_cash BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    auto_increment BOOLEAN NOT NULL DEFAULT FALSE,
    increment_amount REAL NOT NULL DEFAULT 0,
    ordering INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME,
    FOREIGN KEY (asset_type_id) REFERENCES asset_types(id)
);

INSERT INTO assets_new (
    id,
    asset_type_id,
    name,
    broker,
    is_cash,
    is_active,
    auto_increment,
    increment_amount,
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
    CASE
        WHEN auto_increment <> 0 THEN TRUE
        ELSE FALSE
    END AS auto_increment,
    auto_increment AS increment_amount,
    ordering,
    deleted_at
FROM assets;

DROP TABLE assets;

ALTER TABLE assets_new RENAME TO assets;

COMMIT;

PRAGMA foreign_keys = ON;
