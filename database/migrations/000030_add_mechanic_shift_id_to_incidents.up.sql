ALTER TABLE incidents
    ADD COLUMN IF NOT EXISTS mechanic_shift_id BIGINT NULL;

DO $$
BEGIN
    ALTER TABLE incidents
        ADD CONSTRAINT fk_incidents_mechanic_shift
        FOREIGN KEY (mechanic_shift_id)
        REFERENCES mechanic_shifts(id)
        ON DELETE RESTRICT;
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_incidents_mechanic_shift_id
    ON incidents (mechanic_shift_id);
