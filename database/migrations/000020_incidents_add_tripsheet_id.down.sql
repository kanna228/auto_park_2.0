DROP INDEX IF EXISTS idx_incidents_tripsheet_id;

ALTER TABLE incidents
DROP CONSTRAINT IF EXISTS fk_incidents_tripsheet;

ALTER TABLE incidents
DROP COLUMN IF EXISTS tripsheet_id;