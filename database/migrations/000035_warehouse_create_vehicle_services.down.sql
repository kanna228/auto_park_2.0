DROP INDEX IF EXISTS idx_vehicle_services_vehicle_date;
DROP INDEX IF EXISTS idx_vehicle_services_service_date;
DROP INDEX IF EXISTS idx_vehicle_services_vehicle_id;
DROP INDEX IF EXISTS idx_vehicle_services_part_id;
DROP INDEX IF EXISTS idx_vehicle_services_type_id;
DROP INDEX IF EXISTS idx_service_types_name;
DROP INDEX IF EXISTS idx_parts_collection_name;

DROP TABLE IF EXISTS vehicle_services;
DROP TABLE IF EXISTS service_types;
DROP TABLE IF EXISTS parts_collection;
