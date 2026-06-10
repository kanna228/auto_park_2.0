-- Restore demo operational dates close to the original May 2026 demo window.

WITH numbered AS (
    SELECT
        t.id,
        row_number() OVER (ORDER BY t.tripsheet_number) AS rn
    FROM tripsheets t
    WHERE t.tripsheet_number LIKE 'WL-2026-%'
),
mapped AS (
    SELECT id, DATE '2026-05-05' + ((rn - 1) % 21)::int AS target_date
    FROM numbered
)
UPDATE tripsheets t
SET tripsheet_date = m.target_date,
    start_time = m.target_date + time '08:00',
    end_time = CASE WHEN t.end_time IS NULL THEN NULL ELSE m.target_date + time '17:00' END,
    updated_at = NOW()
FROM mapped m
WHERE t.id = m.id;

WITH numbered AS (
    SELECT
        fr.id,
        row_number() OVER (ORDER BY fr.id) AS rn
    FROM fuel_refills fr
    JOIN tripsheets t ON t.id = fr.tripsheet_id
    WHERE t.tripsheet_number LIKE 'WL-2026-%'
),
mapped AS (
    SELECT id, DATE '2026-05-05' + ((rn - 1) % 21)::int AS target_date
    FROM numbered
)
UPDATE fuel_refills fr
SET date = m.target_date,
    created_at = (m.target_date + fr.time)::timestamptz,
    updated_at = NOW()
FROM mapped m
WHERE fr.id = m.id;

WITH numbered AS (
    SELECT
        vi.id,
        row_number() OVER (ORDER BY vi.id) AS rn
    FROM vehicle_part_installations vi
    JOIN parts_catalog pc ON pc.id = vi.part_id
    LEFT JOIN vehicles v ON v.id = vi.vehicle_id
    WHERE pc.part_id LIKE 'DEMO-%'
       OR v.board_number BETWEEN 'B-101' AND 'B-120'
),
mapped AS (
    SELECT id, DATE '2026-05-05' + ((rn - 1) % 21)::int AS target_date
    FROM numbered
)
UPDATE vehicle_part_installations vi
SET installed_at = m.target_date,
    created_at = (m.target_date + time '11:00')::timestamptz,
    updated_at = NOW()
FROM mapped m
WHERE vi.id = m.id;

WITH numbered AS (
    SELECT
        psm.id,
        row_number() OVER (ORDER BY psm.id) AS rn
    FROM part_stock_movements psm
    JOIN parts_catalog pc ON pc.id = psm.part_id
    WHERE pc.part_id LIKE 'DEMO-%'
      AND psm.type = 'issue'
),
mapped AS (
    SELECT id, DATE '2026-05-05' + ((rn - 1) % 21)::int AS target_date
    FROM numbered
)
UPDATE part_stock_movements psm
SET created_at = (m.target_date + time '12:00')::timestamptz
FROM mapped m
WHERE psm.id = m.id;

WITH numbered AS (
    SELECT
        i.id,
        row_number() OVER (ORDER BY i.id) AS rn
    FROM incidents i
    JOIN vehicles v ON v.id = i.vehicle_id
    WHERE v.board_number BETWEEN 'B-101' AND 'B-120'
),
mapped AS (
    SELECT id, DATE '2026-05-08' + ((rn - 1) % 15)::int AS target_date
    FROM numbered
)
UPDATE incidents i
SET incident_date = m.target_date,
    created_at = (m.target_date + i.incident_time)::timestamp,
    updated_at = NOW()
FROM mapped m
WHERE i.id = m.id;

WITH numbered AS (
    SELECT
        me.id,
        row_number() OVER (ORDER BY me.id) AS rn
    FROM maintenance_executions me
    WHERE me.board BETWEEN 'B-101' AND 'B-120'
),
mapped AS (
    SELECT id, DATE '2026-05-06' + ((rn - 1) % 15)::int AS target_date
    FROM numbered
)
UPDATE maintenance_executions me
SET created_at = (m.target_date + time '09:00')::timestamptz,
    updated_at = NOW()
FROM mapped m
WHERE me.id = m.id;
