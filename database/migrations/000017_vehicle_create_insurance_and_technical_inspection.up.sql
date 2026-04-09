CREATE TABLE IF NOT EXISTS insurance (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    file_path TEXT,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_insurance_vehicle
        FOREIGN KEY (vehicle_id)
        REFERENCES vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_insurance_vehicle_id
    ON insurance(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_insurance_is_active
    ON insurance(is_active);

CREATE INDEX IF NOT EXISTS idx_insurance_start_date
    ON insurance(start_date);

CREATE INDEX IF NOT EXISTS idx_insurance_end_date
    ON insurance(end_date);


CREATE TABLE IF NOT EXISTS technical_inspection (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    file_path TEXT,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_technical_inspection_vehicle
        FOREIGN KEY (vehicle_id)
        REFERENCES vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_technical_inspection_vehicle_id
    ON technical_inspection(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_technical_inspection_is_active
    ON technical_inspection(is_active);

CREATE INDEX IF NOT EXISTS idx_technical_inspection_start_date
    ON technical_inspection(start_date);

CREATE INDEX IF NOT EXISTS idx_technical_inspection_end_date
    ON technical_inspection(end_date);