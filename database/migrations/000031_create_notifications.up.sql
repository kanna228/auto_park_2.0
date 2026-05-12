CREATE TABLE IF NOT EXISTS notification_types (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO notification_types (id, code, name) VALUES
    (1, 'part_request_created', 'Создана заявка на деталь'),
    (2, 'part_request_approved', 'Заявка на деталь утверждена'),
    (3, 'part_request_rejected', 'Заявка на деталь отклонена')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();

SELECT setval(
    pg_get_serial_sequence('notification_types', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM notification_types), 1),
    true
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_type_id BIGINT NOT NULL REFERENCES notification_types(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_readed BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id
    ON notifications(user_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id, is_readed, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_type_id
    ON notifications(notification_type_id);

CREATE INDEX IF NOT EXISTS idx_notifications_created_at
    ON notifications(created_at DESC);
