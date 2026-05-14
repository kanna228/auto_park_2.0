DROP INDEX IF EXISTS idx_drivers_assigned_vehicle_id;

ALTER TABLE drivers
    DROP CONSTRAINT IF EXISTS drivers_assigned_vehicle_id_fkey,
    DROP COLUMN IF EXISTS assigned_vehicle_id;
