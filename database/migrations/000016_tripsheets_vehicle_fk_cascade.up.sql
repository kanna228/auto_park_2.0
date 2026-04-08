ALTER TABLE tripsheets
DROP CONSTRAINT IF EXISTS tripsheets_vehicle_id_fkey;

ALTER TABLE tripsheets
ADD CONSTRAINT tripsheets_vehicle_id_fkey
FOREIGN KEY (vehicle_id)
REFERENCES vehicles(id)
ON DELETE CASCADE;