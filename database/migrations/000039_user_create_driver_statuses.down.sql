DROP INDEX IF EXISTS idx_drivers_status_id;

ALTER TABLE drivers
    DROP CONSTRAINT IF EXISTS drivers_status_id_fkey,
    ALTER COLUMN status_id DROP DEFAULT,
    DROP COLUMN IF EXISTS status_id;

DROP TABLE IF EXISTS driver_statuses;
