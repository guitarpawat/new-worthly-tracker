INSERT INTO asset_types (id, name, ordering, is_active) VALUES
    (1, 'Cash', 2, TRUE),
    (2, 'Investment', 1, TRUE),
    (3, 'Credit Card', 3, TRUE),
    (4, 'Savings', 4, TRUE),
    (5, 'Legacy', 5, FALSE);

INSERT INTO assets (id, asset_type_id, name, broker, is_cash, is_active, auto_increment, ordering) VALUES
    (1, 1, 'Emergency Fund', 'SCB', TRUE, TRUE, 0, 1),
    (2, 2, 'SET50 ETF', 'KKP', FALSE, TRUE, 6500, 1),
    (3, 2, 'US Index Fund', 'IBKR', FALSE, TRUE, 5200, 2),
    (4, 3, 'Visa Platinum', 'KBank', TRUE, TRUE, 0, 1),
    (5, 2, 'Gold Fund', 'KKP', FALSE, TRUE, 3200, 3),
    (6, 1, 'Daily Wallet', 'SCB', TRUE, TRUE, 0, 2),
    (7, 4, 'Travel Fund', 'TTB', TRUE, TRUE, 0, 1),
    (8, 2, 'Old Mutual Fund', 'IBKR', FALSE, FALSE, 0, 4),
    (9, 3, 'MasterCard Titanium', 'KTC', TRUE, FALSE, 0, 2),
    (10, 5, 'Legacy Bond', 'Local Bank', FALSE, TRUE, 0, 1);

WITH RECURSIVE months(snapshot_id, record_date) AS (
    VALUES (1, date('2024-05-12'))
    UNION ALL
    SELECT snapshot_id + 1, date(record_date, '+1 month')
    FROM months
    WHERE snapshot_id < 24
)
INSERT INTO record_snapshots (id, record_date)
SELECT snapshot_id, record_date
FROM months;

WITH RECURSIVE months(snapshot_id, record_date) AS (
    VALUES (1, date('2024-05-12'))
    UNION ALL
    SELECT snapshot_id + 1, date(record_date, '+1 month')
    FROM months
    WHERE snapshot_id < 24
)
INSERT INTO record_items (id, snapshot_id, asset_id, bought_price, current_price, remarks)
SELECT
    (snapshot_id - 1) * 7 + 1 AS id,
    snapshot_id,
    1 AS asset_id,
    0 AS bought_price,
    80000 + (snapshot_id - 1) * 2500 AS current_price,
    'Cash reserve'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 7 + 2 AS id,
    snapshot_id,
    2 AS asset_id,
    120000 + (snapshot_id - 1) * 6500 AS bought_price,
    126000 + (snapshot_id - 1) * 7100 AS current_price,
    'Thai core equity'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 7 + 3 AS id,
    snapshot_id,
    3 AS asset_id,
    90000 + (snapshot_id - 1) * 5200 AS bought_price,
    94000 + (snapshot_id - 1) * 5600 AS current_price,
    'Global DCA'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 7 + 4 AS id,
    snapshot_id,
    4 AS asset_id,
    0 AS bought_price,
    -11800 - (snapshot_id - 1) * 520 AS current_price,
    'Outstanding balance'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 7 + 5 AS id,
    snapshot_id,
    5 AS asset_id,
    45000 + (snapshot_id - 1) * 3200 AS bought_price,
    47000 + (snapshot_id - 1) * 3550 AS current_price,
    'Inflation hedge'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 7 + 6 AS id,
    snapshot_id,
    6 AS asset_id,
    0 AS bought_price,
    18000 + (snapshot_id - 1) * 700 AS current_price,
    'Pocket cash'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 7 + 7 AS id,
    snapshot_id,
    7 AS asset_id,
    0 AS bought_price,
    25000 + (snapshot_id - 1) * 1150 AS current_price,
    'Trip budget'
FROM months;
