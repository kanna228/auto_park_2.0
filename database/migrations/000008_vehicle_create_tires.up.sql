-- tire_places
CREATE TABLE IF NOT EXISTS tire_places (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- заполняем справочник мест шин
INSERT INTO tire_places (name)
VALUES
    ('передняя левая'),
    ('передняя правая'),
    ('задняя левая'),
    ('задняя правая'),
    ('запасная шина'),
    ('на складе')
ON CONFLICT (name) DO NOTHING;

-- tires
CREATE TABLE IF NOT EXISTS tires (
    id BIGSERIAL PRIMARY KEY,
    place_id BIGINT NOT NULL,
    vehicle_id BIGINT NOT NULL,
    tire VARCHAR(255) NOT NULL,
    mileage BIGINT NOT NULL DEFAULT 0 CHECK (mileage >= 0),
    max_usage BIGINT NOT NULL DEFAULT 0 CHECK (max_usage >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tires_place
        FOREIGN KEY (place_id)
        REFERENCES tire_places(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_tires_vehicle
        FOREIGN KEY (vehicle_id)
        REFERENCES vehicles(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tires_vehicle_id
    ON tires(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_tires_place_id
    ON tires(place_id);