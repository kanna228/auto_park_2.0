ALTER TABLE parts_catalog
    ADD COLUMN IF NOT EXISTS min_stock_quantity BIGINT NOT NULL DEFAULT 0 CHECK (min_stock_quantity >= 0),
    ADD COLUMN IF NOT EXISTS unit TEXT NOT NULL DEFAULT 'шт';

CREATE TABLE IF NOT EXISTS part_stock_movements (
    id BIGSERIAL PRIMARY KEY,
    part_id BIGINT NOT NULL REFERENCES parts_catalog(id) ON DELETE RESTRICT,
    type TEXT NOT NULL CHECK (type IN ('arrival', 'issue', 'return', 'writeoff')),
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    vehicle_id BIGINT NULL REFERENCES vehicles(id) ON DELETE SET NULL,
    part_request_id BIGINT NULL REFERENCES part_requests(id) ON DELETE SET NULL,
    document_number TEXT NULL,
    actor_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_stock_movements_part_id
    ON part_stock_movements(part_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_part_stock_movements_vehicle_id
    ON part_stock_movements(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_part_stock_movements_part_request_id
    ON part_stock_movements(part_request_id);

CREATE TABLE IF NOT EXISTS part_arrivals (
    id BIGSERIAL PRIMARY KEY,
    document_number TEXT NOT NULL UNIQUE,
    arrival_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'accepted')),
    comment TEXT NULL,
    created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    accepted_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    accepted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_arrivals_arrival_date
    ON part_arrivals(arrival_date DESC);

CREATE INDEX IF NOT EXISTS idx_part_arrivals_status
    ON part_arrivals(status);

CREATE TABLE IF NOT EXISTS part_arrival_items (
    id BIGSERIAL PRIMARY KEY,
    arrival_id BIGINT NOT NULL REFERENCES part_arrivals(id) ON DELETE CASCADE,
    part_id BIGINT NOT NULL REFERENCES parts_catalog(id) ON DELETE RESTRICT,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    price NUMERIC(12,2) NULL CHECK (price IS NULL OR price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_arrival_items_arrival_id
    ON part_arrival_items(arrival_id);

CREATE INDEX IF NOT EXISTS idx_part_arrival_items_part_id
    ON part_arrival_items(part_id);
