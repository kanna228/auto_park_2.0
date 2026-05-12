DROP TRIGGER IF EXISTS trg_set_mechanic_shift_activity ON mechanic_shifts;

DROP FUNCTION IF EXISTS set_mechanic_shift_activity();

DROP FUNCTION IF EXISTS refresh_mechanic_shifts_activity();

DROP INDEX IF EXISTS idx_mechanic_shifts_user_date;
DROP INDEX IF EXISTS idx_mechanic_shifts_is_deleted;
DROP INDEX IF EXISTS idx_mechanic_shifts_is_active;
DROP INDEX IF EXISTS idx_mechanic_shifts_shift_date;
DROP INDEX IF EXISTS idx_mechanic_shifts_user_id;

DROP TABLE IF EXISTS mechanic_shifts;