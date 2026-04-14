BEGIN;

CREATE TABLE IF NOT EXISTS goals (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    target_amount REAL NOT NULL DEFAULT 0,
    target_date DATE,
    deleted_at DATETIME
);

COMMIT;
