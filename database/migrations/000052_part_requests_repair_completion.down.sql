DROP INDEX IF EXISTS idx_part_requests_vehicle_part_installation_id;
DROP INDEX IF EXISTS idx_part_requests_repair_status;
DROP INDEX IF EXISTS idx_part_requests_mechanic_shift_id;
DROP INDEX IF EXISTS idx_part_requests_vehicle_id;

ALTER TABLE part_requests
    DROP CONSTRAINT IF EXISTS part_requests_repair_status_check;

ALTER TABLE part_requests
    DROP COLUMN IF EXISTS vehicle_part_installation_id,
    DROP COLUMN IF EXISTS completed_by_user_id,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS repair_status,
    DROP COLUMN IF EXISTS planned_replacement_at,
    DROP COLUMN IF EXISTS mechanic_shift_id,
    DROP COLUMN IF EXISTS vehicle_id;
