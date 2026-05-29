DELETE FROM notification_types
WHERE code IN ('purchase_request_created', 'purchase_request_confirmed');

DELETE FROM part_request_statuses
WHERE code = 'issued'
  AND NOT EXISTS (
      SELECT 1 FROM part_requests
      WHERE status_id = part_request_statuses.id
  )
  AND NOT EXISTS (
      SELECT 1 FROM part_request_history
      WHERE status_id = part_request_statuses.id
  );

DROP INDEX IF EXISTS idx_parts_catalog_is_archived;

ALTER TABLE parts_catalog
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS is_archived;

DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_is_archived;

ALTER TABLE users
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS is_archived,
    DROP COLUMN IF EXISTS is_active;

DROP INDEX IF EXISTS idx_vehicles_active_id;
DROP INDEX IF EXISTS idx_vehicles_is_archived;

ALTER TABLE vehicles
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS is_archived;
