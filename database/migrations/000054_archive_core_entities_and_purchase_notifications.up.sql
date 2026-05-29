ALTER TABLE vehicles
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_vehicles_is_archived
    ON vehicles(is_archived);

CREATE INDEX IF NOT EXISTS idx_vehicles_active_id
    ON vehicles(id)
    WHERE is_archived = FALSE;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_users_is_archived
    ON users(is_archived);

CREATE INDEX IF NOT EXISTS idx_users_is_active
    ON users(is_active);

ALTER TABLE parts_catalog
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_parts_catalog_is_archived
    ON parts_catalog(is_archived);

INSERT INTO notification_types (code, name) VALUES
    ('purchase_request_created', 'Создана заявка на закупку'),
    ('purchase_request_confirmed', 'Заявка на закупку подтверждена')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

INSERT INTO part_request_statuses (code, name) VALUES
    ('issued', 'Выдано')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();
