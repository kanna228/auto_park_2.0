DROP INDEX IF EXISTS idx_part_requests_rejection_comment_not_null;

ALTER TABLE part_requests
    DROP COLUMN IF EXISTS rejection_comment;