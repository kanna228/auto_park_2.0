ALTER TABLE vehicles
DROP CONSTRAINT IF EXISTS fk_vehicles_status;

DROP INDEX IF EXISTS idx_vehicles_status_id;

ALTER TABLE vehicles
DROP COLUMN IF EXISTS status_id;

DROP TABLE IF EXISTS vehicle_status;