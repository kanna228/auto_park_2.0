ALTER TABLE parts_catalog
    ADD COLUMN IF NOT EXISTS is_consumable BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS vehicle_part_installations (
    id BIGSERIAL PRIMARY KEY,
    part_id BIGINT NOT NULL REFERENCES parts_catalog(id) ON DELETE RESTRICT,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    installed_at DATE NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    installed_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_part_id
    ON vehicle_part_installations(part_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_vehicle_id
    ON vehicle_part_installations(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_installed_by_user_id
    ON vehicle_part_installations(installed_by_user_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_installed_at
    ON vehicle_part_installations(installed_at DESC);

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_active
    ON vehicle_part_installations(vehicle_id, part_id, is_active);
