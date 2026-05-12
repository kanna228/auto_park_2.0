DELETE FROM notifications
WHERE notification_type_id IN (
    SELECT id
    FROM notification_types
    WHERE code IN (
        'vehicle_part_replacement_7_days',
        'vehicle_part_replacement_today'
    )
);

DELETE FROM notification_types
WHERE code IN (
    'vehicle_part_replacement_7_days',
    'vehicle_part_replacement_today'
);

DROP INDEX IF EXISTS ux_notifications_user_dedup_key;

ALTER TABLE notifications
    DROP COLUMN IF EXISTS dedup_key;
