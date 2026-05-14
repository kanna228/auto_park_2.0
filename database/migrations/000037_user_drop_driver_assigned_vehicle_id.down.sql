ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS assigned_vehicle_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'drivers_assigned_vehicle_id_fkey'
    ) THEN
        ALTER TABLE drivers
            ADD CONSTRAINT drivers_assigned_vehicle_id_fkey
            FOREIGN KEY (assigned_vehicle_id)
            REFERENCES vehicles(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_drivers_assigned_vehicle_id
    ON drivers(assigned_vehicle_id);
