DROP INDEX IF EXISTS idx_vehicle_part_installations_mechanic_shift_id;

ALTER TABLE vehicle_part_installations
    DROP CONSTRAINT IF EXISTS vehicle_part_installations_mechanic_shift_id_fkey;

ALTER TABLE vehicle_part_installations
    DROP COLUMN IF EXISTS mechanic_shift_id;
