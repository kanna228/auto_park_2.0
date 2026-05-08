ALTER TABLE part_requests
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_part_requests_is_deleted
    ON part_requests(is_deleted);

CREATE TABLE IF NOT EXISTS part_request_history (
    id BIGSERIAL PRIMARY KEY,
    part_request_id BIGINT NOT NULL REFERENCES part_requests(id) ON DELETE RESTRICT,
    status_id BIGINT NOT NULL REFERENCES part_request_statuses(id) ON DELETE RESTRICT,
    changed_by_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    comment TEXT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_part_request_history_request_id
    ON part_request_history(part_request_id);

CREATE INDEX IF NOT EXISTS idx_part_request_history_status_id
    ON part_request_history(status_id);

CREATE INDEX IF NOT EXISTS idx_part_request_history_changed_by_user_id
    ON part_request_history(changed_by_user_id);

CREATE INDEX IF NOT EXISTS idx_part_request_history_changed_at
    ON part_request_history(changed_at DESC);

INSERT INTO part_request_history (
    part_request_id,
    status_id,
    changed_by_user_id,
    comment,
    changed_at
)
SELECT
    pr.id,
    pr.status_id,
    pr.author_user_id,
    'Начальная запись истории заявки',
    pr.created_at
FROM part_requests pr
WHERE NOT EXISTS (
    SELECT 1
    FROM part_request_history h
    WHERE h.part_request_id = pr.id
);