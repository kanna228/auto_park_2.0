DELETE FROM notifications
WHERE notification_type_id IN (
    SELECT id
    FROM notification_types
    WHERE code IN ('incident_created', 'part_request_stock_shortage')
);

DELETE FROM notification_types
WHERE code IN ('incident_created', 'part_request_stock_shortage');
