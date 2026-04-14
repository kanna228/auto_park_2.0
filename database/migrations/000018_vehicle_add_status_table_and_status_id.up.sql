CREATE TABLE IF NOT EXISTS vehicle_status (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO vehicle_status (name)
VALUES
    ('В использовании'),
    ('На ТО'),
    ('Не используется'),
    ('На ремонте'),
    ('Списан')
ON CONFLICT (name) DO NOTHING;


ALTER TABLE vehicles
ADD COLUMN IF NOT EXISTS status_id BIGINT;

UPDATE vehicles
SET status_id = (
    SELECT id
    FROM vehicle_status
    WHERE name = 'Не используется'
    LIMIT 1
)
WHERE status_id IS NULL;

ALTER TABLE vehicles
ALTER COLUMN status_id SET NOT NULL;

ALTER TABLE vehicles
ADD CONSTRAINT fk_vehicles_status
FOREIGN KEY (status_id)
REFERENCES vehicle_status(id)
ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_vehicles_status_id
    ON vehicles(status_id);