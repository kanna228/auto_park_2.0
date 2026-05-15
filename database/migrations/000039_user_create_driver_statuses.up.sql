CREATE TABLE IF NOT EXISTS driver_statuses (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO driver_statuses (id, code, name, description, created_at, updated_at)
VALUES
    (1, 'available', 'Доступен', 'Водитель доступен и может быть назначен на рейс', NOW(), NOW()),
    (2, 'on_trip', 'На выезде', 'Водитель находится на выезде или выполняет путевой лист', NOW(), NOW()),
    (3, 'unavailable', 'Недоступен', 'Водитель временно недоступен для назначения', NOW(), NOW()),
    (4, 'vacation', 'В отпуске', 'Водитель находится в отпуске', NOW(), NOW()),
    (5, 'sick_leave', 'На больничном', 'Водитель находится на больничном', NOW(), NOW())
ON CONFLICT (id) DO UPDATE
SET code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

SELECT setval(
    pg_get_serial_sequence('driver_statuses', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM driver_statuses), 1),
    true
);

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS status_id BIGINT;

UPDATE drivers
SET status_id = 1
WHERE status_id IS NULL;

ALTER TABLE drivers
    ALTER COLUMN status_id SET DEFAULT 1,
    ALTER COLUMN status_id SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'drivers_status_id_fkey'
    ) THEN
        ALTER TABLE drivers
            ADD CONSTRAINT drivers_status_id_fkey
            FOREIGN KEY (status_id)
            REFERENCES driver_statuses(id)
            ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_drivers_status_id
    ON drivers(status_id);
