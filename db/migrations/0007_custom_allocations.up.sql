BEGIN;

CREATE TABLE custom_allocation_charts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL,
    deleted_at DATETIME
);
CREATE UNIQUE INDEX custom_allocation_chart_name ON custom_allocation_charts(name_key) WHERE deleted_at IS NULL;

CREATE TABLE custom_allocation_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chart_id INTEGER NOT NULL REFERENCES custom_allocation_charts(id),
    asset_id INTEGER NOT NULL REFERENCES assets(id),
    category TEXT NOT NULL,
    category_key TEXT NOT NULL,
    percentage TEXT NOT NULL,
    deleted_at DATETIME
);
CREATE UNIQUE INDEX custom_allocation_entry_category
    ON custom_allocation_entries(chart_id, asset_id, category_key) WHERE deleted_at IS NULL;

COMMIT;
