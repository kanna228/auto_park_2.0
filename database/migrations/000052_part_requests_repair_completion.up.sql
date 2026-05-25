ALTER TABLE part_requests
    ADD COLUMN IF NOT EXISTS vehicle_id BIGINT NULL REFERENCES vehicles(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS mechanic_shift_id BIGINT NULL REFERENCES mechanic_shifts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS planned_replacement_at DATE NULL,
    ADD COLUMN IF NOT EXISTS repair_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS completed_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS vehicle_part_installation_id BIGINT NULL REFERENCES vehicle_part_installations(id) ON DELETE SET NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'part_requests_repair_status_check'
    ) THEN
        ALTER TABLE part_requests
            ADD CONSTRAINT part_requests_repair_status_check
            CHECK (repair_status IN ('pending', 'in_progress', 'completed'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_part_requests_vehicle_id
    ON part_requests(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_part_requests_mechanic_shift_id
    ON part_requests(mechanic_shift_id);

CREATE INDEX IF NOT EXISTS idx_part_requests_repair_status
    ON part_requests(repair_status);

CREATE INDEX IF NOT EXISTS idx_part_requests_vehicle_part_installation_id
    ON part_requests(vehicle_part_installation_id);
