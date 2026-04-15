CREATE TABLE IF NOT EXISTS incident_types (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO incident_types (name)
VALUES
    ('ДТП'),
    ('Поломка'),
    ('Повреждение')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS incidents (
    id BIGSERIAL PRIMARY KEY,
    incident_type_id BIGINT NOT NULL,
    vehicle_id BIGINT NOT NULL,
    driver_id BIGINT NOT NULL,
    mechanic_id BIGINT NOT NULL,
    incident_date DATE NOT NULL,
    incident_time TIME NOT NULL,
    location VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_incidents_incident_type
        FOREIGN KEY (incident_type_id)
        REFERENCES incident_types(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_incidents_vehicle
        FOREIGN KEY (vehicle_id)
        REFERENCES vehicles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_incidents_driver
        FOREIGN KEY (driver_id)
        REFERENCES drivers(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_incidents_mechanic
        FOREIGN KEY (mechanic_id)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_incidents_incident_type_id
    ON incidents (incident_type_id);

CREATE INDEX IF NOT EXISTS idx_incidents_vehicle_id
    ON incidents (vehicle_id);

CREATE INDEX IF NOT EXISTS idx_incidents_driver_id
    ON incidents (driver_id);

CREATE INDEX IF NOT EXISTS idx_incidents_mechanic_id
    ON incidents (mechanic_id);

CREATE INDEX IF NOT EXISTS idx_incidents_incident_date
    ON incidents (incident_date);
