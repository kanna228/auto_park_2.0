CREATE TABLE IF NOT EXISTS part_request_statuses (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO part_request_statuses (id, code, name) VALUES
    (1, 'new', 'Новая'),
    (2, 'rejected', 'Отклонена'),
    (3, 'approved', 'Утверждена')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name;

SELECT setval(
    pg_get_serial_sequence('part_request_statuses', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM part_request_statuses), 1),
    true
);

CREATE TABLE IF NOT EXISTS part_requests (
    id BIGSERIAL PRIMARY KEY,
    part_id BIGINT NOT NULL REFERENCES parts_catalog(id) ON DELETE RESTRICT,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    mechanic_comment TEXT NOT NULL,
    status_id BIGINT NOT NULL DEFAULT 1 REFERENCES part_request_statuses(id) ON DELETE RESTRICT,
    author_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_requests_part_id ON part_requests(part_id);
CREATE INDEX IF NOT EXISTS idx_part_requests_status_id ON part_requests(status_id);
CREATE INDEX IF NOT EXISTS idx_part_requests_author_user_id ON part_requests(author_user_id);
CREATE INDEX IF NOT EXISTS idx_part_requests_created_at ON part_requests(created_at DESC);
