CREATE TABLE IF NOT EXISTS fuel_refills (
    id BIGSERIAL PRIMARY KEY,
    tripsheet_id BIGINT NOT NULL REFERENCES tripsheets(id) ON DELETE CASCADE,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    fuel_amount NUMERIC(10,2) NOT NULL CHECK (fuel_amount > 0),
    date DATE NOT NULL,
    time TIME NOT NULL,
    location VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fuel_refills_tripsheet_id ON fuel_refills(tripsheet_id);
CREATE INDEX IF NOT EXISTS idx_fuel_refills_vehicle_id ON fuel_refills(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fuel_refills_date ON fuel_refills(date);
