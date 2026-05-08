DROP INDEX IF EXISTS idx_part_request_history_changed_at;
DROP INDEX IF EXISTS idx_part_request_history_changed_by_user_id;
DROP INDEX IF EXISTS idx_part_request_history_status_id;
DROP INDEX IF EXISTS idx_part_request_history_request_id;

DROP TABLE IF EXISTS part_request_history;

DROP INDEX IF EXISTS idx_part_requests_is_deleted;

ALTER TABLE part_requests
    DROP COLUMN IF EXISTS deleted_by_user_id,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS is_deleted;