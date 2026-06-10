-- Move demo operational records onto the current calendar so analytics are
-- populated by real database rows after a fresh migration run.

WITH bounds AS (
    SELECT
        date_trunc('week', CURRENT_DATE)::date AS week_start,
        date_trunc('month', CURRENT_DATE)::date AS month_start
),
numbered AS (
    SELECT
        t.id,
        row_number() OVER (ORDER BY t.tripsheet_number) AS rn
    FROM tripsheets t
    WHERE t.tripsheet_number LIKE 'WL-2026-%'
),
mapped AS (
    SELECT
        n.id,
        CASE
            WHEN n.rn <= 8 THEN b.week_start + ((n.rn - 1) % 5)::int
            WHEN n.rn <= 12 THEN b.week_start - 7 + ((n.rn - 9) % 5)::int
            ELSE (b.month_start - ((((n.rn - 13) % 5) + 1)::int || ' months')::interval)::date + ((n.rn - 13) % 20)::int
        END AS target_date
    FROM numbered n
    CROSS JOIN bounds b
)
UPDATE tripsheets t
SET tripsheet_date = m.target_date,
    start_time = m.target_date + time '08:00',
    end_time = CASE WHEN t.end_time IS NULL THEN NULL ELSE m.target_date + time '17:00' END,
    updated_at = NOW()
FROM mapped m
WHERE t.id = m.id;

WITH bounds AS (
    SELECT
        date_trunc('week', CURRENT_DATE)::date AS week_start,
        date_trunc('month', CURRENT_DATE)::date AS month_start
),
numbered AS (
    SELECT
        fr.id,
        row_number() OVER (ORDER BY fr.id) AS rn
    FROM fuel_refills fr
    JOIN tripsheets t ON t.id = fr.tripsheet_id
    WHERE t.tripsheet_number LIKE 'WL-2026-%'
),
mapped AS (
    SELECT
        n.id,
        CASE
            WHEN n.rn <= 8 THEN b.week_start + ((n.rn - 1) % 5)::int
            WHEN n.rn <= 12 THEN b.week_start - 7 + ((n.rn - 9) % 5)::int
            ELSE (b.month_start - ((((n.rn - 13) % 5) + 1)::int || ' months')::interval)::date + ((n.rn - 13) % 20)::int
        END AS target_date
    FROM numbered n
    CROSS JOIN bounds b
)
UPDATE fuel_refills fr
SET date = m.target_date,
    price = CASE WHEN fr.price > 0 THEN fr.price ELSE fr.fuel_amount * 265 END,
    created_at = (m.target_date + fr.time)::timestamptz,
    updated_at = NOW()
FROM mapped m
WHERE fr.id = m.id;

WITH bounds AS (
    SELECT
        date_trunc('week', CURRENT_DATE)::date AS week_start,
        date_trunc('month', CURRENT_DATE)::date AS month_start
),
numbered AS (
    SELECT
        vi.id,
        pc.price AS catalog_price,
        row_number() OVER (ORDER BY vi.id) AS rn
    FROM vehicle_part_installations vi
    JOIN parts_catalog pc ON pc.id = vi.part_id
    LEFT JOIN vehicles v ON v.id = vi.vehicle_id
    WHERE pc.part_id LIKE 'DEMO-%'
       OR v.board_number BETWEEN 'B-101' AND 'B-120'
),
mapped AS (
    SELECT
        n.id,
        COALESCE(NULLIF(n.catalog_price, 0), 10000) AS unit_price,
        CASE
            WHEN n.rn <= 8 THEN b.week_start + ((n.rn - 1) % 5)::int
            WHEN n.rn <= 12 THEN b.week_start - 7 + ((n.rn - 9) % 5)::int
            ELSE (b.month_start - ((((n.rn - 13) % 5) + 1)::int || ' months')::interval)::date + ((n.rn - 13) % 20)::int
        END AS target_date
    FROM numbered n
    CROSS JOIN bounds b
)
UPDATE vehicle_part_installations vi
SET installed_at = m.target_date,
    unit_price = m.unit_price,
    total_price = m.unit_price * vi.quantity,
    created_at = (m.target_date + time '11:00')::timestamptz,
    updated_at = NOW()
FROM mapped m
WHERE vi.id = m.id;

WITH bounds AS (
    SELECT
        date_trunc('week', CURRENT_DATE)::date AS week_start,
        date_trunc('month', CURRENT_DATE)::date AS month_start
),
numbered AS (
    SELECT
        psm.id,
        row_number() OVER (ORDER BY psm.id) AS rn
    FROM part_stock_movements psm
    JOIN parts_catalog pc ON pc.id = psm.part_id
    WHERE pc.part_id LIKE 'DEMO-%'
      AND psm.type = 'issue'
),
mapped AS (
    SELECT
        n.id,
        CASE
            WHEN n.rn <= 8 THEN b.week_start + ((n.rn - 1) % 5)::int
            WHEN n.rn <= 12 THEN b.week_start - 7 + ((n.rn - 9) % 5)::int
            ELSE (b.month_start - ((((n.rn - 13) % 5) + 1)::int || ' months')::interval)::date + ((n.rn - 13) % 20)::int
        END AS target_date
    FROM numbered n
    CROSS JOIN bounds b
)
UPDATE part_stock_movements psm
SET created_at = (m.target_date + time '12:00')::timestamptz
FROM mapped m
WHERE psm.id = m.id;

WITH bounds AS (
    SELECT date_trunc('week', CURRENT_DATE)::date AS week_start
),
numbered AS (
    SELECT
        i.id,
        row_number() OVER (ORDER BY i.id) AS rn
    FROM incidents i
    JOIN vehicles v ON v.id = i.vehicle_id
    WHERE v.board_number BETWEEN 'B-101' AND 'B-120'
),
mapped AS (
    SELECT
        n.id,
        b.week_start + ((n.rn - 1) % 5)::int AS target_date
    FROM numbered n
    CROSS JOIN bounds b
)
UPDATE incidents i
SET incident_date = m.target_date,
    created_at = (m.target_date + i.incident_time)::timestamp,
    updated_at = NOW()
FROM mapped m
WHERE i.id = m.id;

WITH bounds AS (
    SELECT date_trunc('week', CURRENT_DATE)::date AS week_start
),
numbered AS (
    SELECT
        me.id,
        row_number() OVER (ORDER BY me.id) AS rn
    FROM maintenance_executions me
    WHERE me.board BETWEEN 'B-101' AND 'B-120'
),
mapped AS (
    SELECT
        n.id,
        b.week_start + ((n.rn - 1) % 5)::int AS target_date
    FROM numbered n
    CROSS JOIN bounds b
)
UPDATE maintenance_executions me
SET created_at = (m.target_date + time '09:00')::timestamptz,
    updated_at = NOW()
FROM mapped m
WHERE me.id = m.id;
