ALTER TABLE incidents
ADD COLUMN IF NOT EXISTS tripsheet_id BIGINT NULL;

ALTER TABLE incidents
ADD CONSTRAINT fk_incidents_tripsheet
FOREIGN KEY (tripsheet_id)
REFERENCES tripsheets(id)
ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_incidents_tripsheet_id
    ON incidents (tripsheet_id);