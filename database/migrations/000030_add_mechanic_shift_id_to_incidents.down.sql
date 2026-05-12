ALTER TABLE incidents
    DROP CONSTRAINT IF EXISTS fk_incidents_mechanic_shift;

DROP INDEX IF EXISTS idx_incidents_mechanic_shift_id;

ALTER TABLE incidents
    DROP COLUMN IF EXISTS mechanic_shift_id;
