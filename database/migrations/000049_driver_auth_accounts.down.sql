DROP INDEX IF EXISTS idx_drivers_last_seen;
DROP INDEX IF EXISTS idx_drivers_role_id;
DROP INDEX IF EXISTS ux_drivers_mail_lower;

ALTER TABLE drivers
    DROP COLUMN IF EXISTS last_seen,
    DROP COLUMN IF EXISTS session_token,
    DROP COLUMN IF EXISTS role_id,
    DROP COLUMN IF EXISTS password;

DELETE FROM roles
WHERE id = 6
  AND name = 'driver';
