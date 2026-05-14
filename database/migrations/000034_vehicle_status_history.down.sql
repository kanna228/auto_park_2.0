DROP TRIGGER IF EXISTS trg_sync_vehicle_status_history ON vehicles;

DROP FUNCTION IF EXISTS sync_vehicle_status_history();

DROP INDEX IF EXISTS ux_vehicle_status_history_current;
DROP INDEX IF EXISTS idx_vehicle_status_history_end_date;
DROP INDEX IF EXISTS idx_vehicle_status_history_start_date;
DROP INDEX IF EXISTS idx_vehicle_status_history_status_id;
DROP INDEX IF EXISTS idx_vehicle_status_history_vehicle_id;

DROP TABLE IF EXISTS vehicle_status_history;
