ALTER TABLE tires
    ADD COLUMN IF NOT EXISTS installed_at DATE NOT NULL DEFAULT CURRENT_DATE;

CREATE INDEX IF NOT EXISTS idx_tires_installed_at
    ON tires(installed_at DESC);

CREATE TABLE IF NOT EXISTS vehicle_documents (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('insurance', 'tachograph')),
    number TEXT NOT NULL,
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vehicle_documents_vehicle_id
    ON vehicle_documents(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_documents_valid_from
    ON vehicle_documents(vehicle_id, valid_from DESC, id DESC);

INSERT INTO driver_statuses (code, name, description, created_at, updated_at)
VALUES ('inactive', 'Inactive', 'Driver is inactive and cannot be assigned to a trip', NOW(), NOW())
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = NOW();

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'available';

UPDATE drivers d
SET status = CASE
    WHEN ds.code = 'unavailable' THEN 'inactive'
    WHEN ds.code IN ('available', 'on_trip', 'inactive') THEN ds.code
    ELSE 'available'
END
FROM driver_statuses ds
WHERE ds.id = d.status_id;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_drivers_status_code'
    ) THEN
        ALTER TABLE drivers
            ADD CONSTRAINT chk_drivers_status_code
            CHECK (status IN ('available', 'on_trip', 'inactive'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_drivers_status
    ON drivers(status);

CREATE TABLE IF NOT EXISTS waybill_route_points (
    id BIGSERIAL PRIMARY KEY,
    waybill_id BIGINT NOT NULL REFERENCES tripsheets(id) ON DELETE CASCADE,
    seq_number INT NOT NULL CHECK (seq_number > 0),
    destination TEXT NOT NULL,
    arrival_time TIME NULL,
    hospitalization_time TIME NULL,
    lpu_arrival_time TIME NULL,
    release_time TIME NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_waybill_route_points_waybill_id
    ON waybill_route_points(waybill_id, seq_number ASC, id ASC);
