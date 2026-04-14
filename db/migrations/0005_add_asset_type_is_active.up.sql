BEGIN;

ALTER TABLE asset_types
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE asset_types
SET is_active = FALSE
WHERE deleted_at IS NOT NULL;

COMMIT;
