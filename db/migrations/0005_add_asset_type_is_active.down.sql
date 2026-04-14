PRAGMA foreign_keys = OFF;

BEGIN;

CREATE TABLE asset_types_new (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    ordering INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME
);

INSERT INTO asset_types_new (
    id,
    name,
    ordering,
    deleted_at
)
SELECT
    id,
    name,
    ordering,
    deleted_at
FROM asset_types;

DROP TABLE asset_types;

ALTER TABLE asset_types_new RENAME TO asset_types;

COMMIT;

PRAGMA foreign_keys = ON;
