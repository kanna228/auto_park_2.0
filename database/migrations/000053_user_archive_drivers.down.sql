DROP INDEX IF EXISTS idx_drivers_active_surname_name;
DROP INDEX IF EXISTS idx_drivers_is_archived;

ALTER TABLE drivers
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS is_archived;
