CREATE TABLE IF NOT EXISTS part_purchase_requests (
    id BIGSERIAL PRIMARY KEY,
    part_id BIGINT NOT NULL REFERENCES parts_catalog(id) ON DELETE RESTRICT,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'confirmed', 'cancelled')),
    source_part_request_id BIGINT NULL REFERENCES part_requests(id) ON DELETE SET NULL,
    comment TEXT NULL,
    created_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    confirmed_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    confirmed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_purchase_requests_part_id
    ON part_purchase_requests(part_id);

CREATE INDEX IF NOT EXISTS idx_part_purchase_requests_status
    ON part_purchase_requests(status);

CREATE INDEX IF NOT EXISTS idx_part_purchase_requests_source_part_request_id
    ON part_purchase_requests(source_part_request_id);

CREATE INDEX IF NOT EXISTS idx_part_purchase_requests_created_at
    ON part_purchase_requests(created_at DESC);
