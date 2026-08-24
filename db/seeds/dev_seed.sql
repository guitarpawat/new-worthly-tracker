INSERT INTO asset_types (id, name, ordering, is_active) VALUES
    (1, 'Cash', 2, TRUE),
    (2, 'Investment', 1, TRUE),
    (3, 'Credit Card', 3, TRUE),
    (4, 'Savings', 4, TRUE),
    (6, 'Property', 5, TRUE),
    (7, 'Retirement', 6, TRUE),
    (8, 'Installment Loan', 7, TRUE),
    (5, 'Legacy', 8, FALSE);

INSERT INTO assets (id, asset_type_id, name, broker, is_cash, is_liability, is_active, auto_increment, ordering) VALUES
    (1, 1, 'Emergency Fund', 'SCB', TRUE, FALSE, TRUE, 0, 1),
    (2, 2, 'SET50 ETF', 'KKP', FALSE, FALSE, TRUE, 6500, 1),
    (3, 2, 'US Index Fund', 'IBKR', FALSE, FALSE, TRUE, 5200, 2),
    (4, 3, 'Visa Platinum', 'KBank', TRUE, TRUE, TRUE, 0, 1),
    (5, 2, 'Gold Fund', 'KKP', FALSE, FALSE, TRUE, 3200, 3),
    (6, 1, 'Daily Wallet', 'SCB', TRUE, FALSE, TRUE, 0, 2),
    (7, 4, 'Travel Fund', 'TTB', TRUE, FALSE, TRUE, 0, 1),
    (8, 2, 'Old Mutual Fund', 'IBKR', FALSE, FALSE, FALSE, 0, 4),
    (9, 3, 'MasterCard Titanium', 'KTC', TRUE, TRUE, FALSE, 0, 2),
    (10, 5, 'Legacy Bond', 'Local Bank', FALSE, FALSE, TRUE, 0, 1),
    (11, 6, 'Condo Unit', 'Demo Property', FALSE, FALSE, TRUE, 0, 1),
    (12, 7, 'Provident Fund', 'Demo Retirement', FALSE, FALSE, TRUE, 4000, 1),
    (13, 8, 'Auto Loan', 'Demo Lender', TRUE, TRUE, TRUE, 0, 1),
    (14, 4, 'Education Fund', 'Demo Bank', TRUE, FALSE, TRUE, 0, 2),
    (15, 2, 'Reserve', 'Demo Broker', FALSE, FALSE, TRUE, 2000, 5),
    (16, 4, 'Reserve', 'Demo Bank', TRUE, FALSE, TRUE, 0, 3);

WITH RECURSIVE months(snapshot_id, record_date) AS (
    VALUES (1, date('2024-05-12'))
    UNION ALL
    SELECT snapshot_id + 1, date(record_date, '+1 month')
    FROM months
    WHERE snapshot_id < 28
)
INSERT INTO record_snapshots (id, record_date)
SELECT snapshot_id, record_date
FROM months;

WITH RECURSIVE months(snapshot_id, record_date) AS (
    VALUES (1, date('2024-05-12'))
    UNION ALL
    SELECT snapshot_id + 1, date(record_date, '+1 month')
    FROM months
    WHERE snapshot_id < 28
)
INSERT INTO record_items (id, snapshot_id, asset_id, bought_price, current_price, remarks)
SELECT
    (snapshot_id - 1) * 20 + 1 AS id,
    snapshot_id,
    1 AS asset_id,
    0 AS bought_price,
    80000 + (snapshot_id - 1) * 2500 AS current_price,
    'Cash reserve'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 2 AS id,
    snapshot_id,
    2 AS asset_id,
    120000 + (snapshot_id - 1) * 6500 AS bought_price,
    126000 + (snapshot_id - 1) * 7100 AS current_price,
    'Thai core equity'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 3 AS id,
    snapshot_id,
    3 AS asset_id,
    90000 + (snapshot_id - 1) * 5200 AS bought_price,
    94000 + (snapshot_id - 1) * 5600 AS current_price,
    'Global DCA'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 4 AS id,
    snapshot_id,
    4 AS asset_id,
    0 AS bought_price,
    -11800 - (snapshot_id - 1) * 520 AS current_price,
    'Outstanding balance'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 5 AS id,
    snapshot_id,
    5 AS asset_id,
    45000 + (snapshot_id - 1) * 3200 AS bought_price,
    47000 + (snapshot_id - 1) * 3550 AS current_price,
    'Inflation hedge'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 6 AS id,
    snapshot_id,
    6 AS asset_id,
    0 AS bought_price,
    18000 + (snapshot_id - 1) * 700 AS current_price,
    'Pocket cash'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 7 AS id,
    snapshot_id,
    7 AS asset_id,
    0 AS bought_price,
    25000 + (snapshot_id - 1) * 1150 AS current_price,
    'Trip budget'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 8 AS id,
    snapshot_id,
    8 AS asset_id,
    70000 + (snapshot_id - 1) * 1800 AS bought_price,
    72000 + (snapshot_id - 1) * 1500 AS current_price,
    'Historical holding'
FROM months
WHERE snapshot_id <= 10
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 9 AS id,
    snapshot_id,
    9 AS asset_id,
    0 AS bought_price,
    -8500 + (snapshot_id - 1) * 900 AS current_price,
    'Paid-off card'
FROM months
WHERE snapshot_id <= 6
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 10 AS id,
    snapshot_id,
    11 AS asset_id,
    1850000 AS bought_price,
    1900000 + (snapshot_id - 7) * 7500 AS current_price,
    'Illustrative property value'
FROM months
WHERE snapshot_id >= 7
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 11 AS id,
    snapshot_id,
    12 AS asset_id,
    60000 + (snapshot_id - 1) * 4000 AS bought_price,
    65000 + (snapshot_id - 1) * 4500 AS current_price,
    'Employer retirement plan'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 12 AS id,
    snapshot_id,
    13 AS asset_id,
    0 AS bought_price,
    -420000 + (snapshot_id - 1) * 12500 AS current_price,
    'Declining installment balance'
FROM months
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 13 AS id,
    snapshot_id,
    14 AS asset_id,
    0 AS bought_price,
    30000 + (snapshot_id - 13) * 3000 AS current_price,
    'Medium-term savings'
FROM months
WHERE snapshot_id >= 13
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 14 AS id,
    snapshot_id,
    15 AS asset_id,
    20000 + (snapshot_id - 18) * 2000 AS bought_price,
    21500 + (snapshot_id - 18) * 2200 AS current_price,
    'Same-name allocation example'
FROM months
WHERE snapshot_id >= 18
UNION ALL
SELECT
    (snapshot_id - 1) * 20 + 15 AS id,
    snapshot_id,
    16 AS asset_id,
    0 AS bought_price,
    15000 + (snapshot_id - 18) * 2500 AS current_price,
    'Same-name cash example'
FROM months
WHERE snapshot_id >= 18;

INSERT INTO goals (id, name, target_amount, target_date) VALUES
    (1, 'First Milestone', 1000000, NULL),
    (2, 'Growth Target', 4000000, '2028-12-31'),
    (3, 'Long-Term Independence', 6000000, '2032-12-31');
