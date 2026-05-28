ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_drivers_is_archived
    ON drivers(is_archived);

CREATE INDEX IF NOT EXISTS idx_drivers_active_surname_name
    ON drivers(surname, name, id)
    WHERE is_archived = FALSE;
