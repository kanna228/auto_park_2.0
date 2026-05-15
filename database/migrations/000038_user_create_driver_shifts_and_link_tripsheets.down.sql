DROP INDEX IF EXISTS idx_tripsheets_driver_shift_id;

ALTER TABLE tripsheets
    DROP CONSTRAINT IF EXISTS fk_tripsheets_driver_shift;

ALTER TABLE tripsheets
    DROP COLUMN IF EXISTS driver_shift_id;

DROP TRIGGER IF EXISTS trg_set_driver_shift_activity ON driver_shifts;

DROP FUNCTION IF EXISTS set_driver_shift_activity();

DROP FUNCTION IF EXISTS refresh_driver_shifts_activity();

DROP INDEX IF EXISTS idx_driver_shifts_driver_date;
DROP INDEX IF EXISTS idx_driver_shifts_is_deleted;
DROP INDEX IF EXISTS idx_driver_shifts_is_active;
DROP INDEX IF EXISTS idx_driver_shifts_shift_date;
DROP INDEX IF EXISTS idx_driver_shifts_driver_id;

DROP TABLE IF EXISTS driver_shifts;
