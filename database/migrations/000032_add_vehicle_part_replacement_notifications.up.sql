ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS dedup_key TEXT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_notifications_user_dedup_key
    ON notifications(user_id, dedup_key)
    WHERE dedup_key IS NOT NULL;

INSERT INTO notification_types (code, name) VALUES
    ('vehicle_part_replacement_7_days', 'Напоминание о замене детали за 7 дней'),
    ('vehicle_part_replacement_today', 'Напоминание о замене детали сегодня')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();
