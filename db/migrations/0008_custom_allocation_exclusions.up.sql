BEGIN;
CREATE TABLE custom_allocation_exclusions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chart_id INTEGER NOT NULL REFERENCES custom_allocation_charts(id),
    asset_id INTEGER NOT NULL REFERENCES assets(id),
    deleted_at DATETIME
);
CREATE UNIQUE INDEX custom_allocation_excluded_asset
    ON custom_allocation_exclusions(chart_id, asset_id) WHERE deleted_at IS NULL;
COMMIT;
