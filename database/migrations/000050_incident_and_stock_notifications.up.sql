INSERT INTO notification_types (code, name) VALUES
    ('incident_created', 'Создан инцидент для осмотра механиком'),
    ('part_request_stock_shortage', 'Нехватка деталей по заявке')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = NOW();
