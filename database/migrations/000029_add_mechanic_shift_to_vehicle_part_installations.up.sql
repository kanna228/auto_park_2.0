ALTER TABLE vehicle_part_installations
    ADD COLUMN IF NOT EXISTS mechanic_shift_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'vehicle_part_installations_mechanic_shift_id_fkey'
    ) THEN
        ALTER TABLE vehicle_part_installations
            ADD CONSTRAINT vehicle_part_installations_mechanic_shift_id_fkey
            FOREIGN KEY (mechanic_shift_id)
            REFERENCES mechanic_shifts(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_mechanic_shift_id
    ON vehicle_part_installations(mechanic_shift_id);
