ALTER TABLE part_requests
    ADD COLUMN IF NOT EXISTS rejection_comment TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_part_requests_rejection_comment_not_null
    ON part_requests(id)
    WHERE rejection_comment IS NOT NULL;