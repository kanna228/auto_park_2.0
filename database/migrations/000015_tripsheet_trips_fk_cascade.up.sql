ALTER TABLE tripsheet_trips
DROP CONSTRAINT IF EXISTS tripsheet_trips_tripsheet_id_fkey;

ALTER TABLE tripsheet_trips
ADD CONSTRAINT tripsheet_trips_tripsheet_id_fkey
FOREIGN KEY (tripsheet_id)
REFERENCES tripsheets(id)
ON DELETE CASCADE;