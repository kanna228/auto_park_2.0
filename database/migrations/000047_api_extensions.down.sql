DROP TABLE IF EXISTS waybill_route_points;

DROP INDEX IF EXISTS idx_drivers_status;

ALTER TABLE drivers
    DROP CONSTRAINT IF EXISTS chk_drivers_status_code,
    DROP COLUMN IF EXISTS status;

DROP INDEX IF EXISTS idx_vehicle_documents_valid_from;
DROP INDEX IF EXISTS idx_vehicle_documents_vehicle_id;
DROP TABLE IF EXISTS vehicle_documents;

DROP INDEX IF EXISTS idx_tires_installed_at;

ALTER TABLE tires
    DROP COLUMN IF EXISTS installed_at;
