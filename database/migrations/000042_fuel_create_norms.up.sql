CREATE TABLE IF NOT EXISTS fuel_norms (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL UNIQUE REFERENCES vehicles(id) ON DELETE CASCADE,
    norm_per_100km NUMERIC(6,2) NOT NULL,
    summer_norm NUMERIC(6,2) NOT NULL,
    winter_norm NUMERIC(6,2) NOT NULL,
    cold_air_norm NUMERIC(6,2) NOT NULL DEFAULT 0,
    warm_air_norm NUMERIC(6,2) NOT NULL DEFAULT 0,
    deviation NUMERIC(5,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fuel_norms_vehicle_id
    ON fuel_norms(vehicle_id);
