ALTER TABLE vehicle_part_installations
    ADD COLUMN IF NOT EXISTS planned_replacement_at DATE NULL;

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_planned_replacement_at
    ON vehicle_part_installations(planned_replacement_at);