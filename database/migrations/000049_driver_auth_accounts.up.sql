INSERT INTO roles (id, name)
VALUES (6, 'driver')
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name;

SELECT setval(
    pg_get_serial_sequence('roles', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM roles), 1),
    true
);

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS password TEXT NULL,
    ADD COLUMN IF NOT EXISTS role_id BIGINT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS session_token TEXT NULL,
    ADD COLUMN IF NOT EXISTS last_seen TIMESTAMPTZ NULL;

UPDATE drivers
SET role_id = (SELECT id FROM roles WHERE name = 'driver' LIMIT 1),
    updated_at = NOW()
WHERE role_id IS NULL;

-- Temporary password for drivers that existed before auth was added: admin123.
-- Newly created drivers receive a generated password by email.
UPDATE drivers
SET password = '$2a$12$93CmzSxMMK.jo6YfoXs5G.rneemNhlMu6DHuYPV3QtX6OQa3JKLme',
    updated_at = NOW()
WHERE password IS NULL
  AND mail IS NOT NULL
  AND BTRIM(mail) <> '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'ux_drivers_mail_lower'
    )
    AND NOT EXISTS (
        SELECT 1
        FROM (
            SELECT LOWER(mail) AS email_key
            FROM drivers
            WHERE mail IS NOT NULL AND BTRIM(mail) <> ''
            GROUP BY LOWER(mail)
            HAVING COUNT(*) > 1
        ) duplicates
    ) THEN
        EXECUTE $create_index$
            CREATE UNIQUE INDEX ux_drivers_mail_lower
            ON drivers (LOWER(mail))
            WHERE mail IS NOT NULL AND BTRIM(mail) <> ''
        $create_index$;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_drivers_role_id
    ON drivers(role_id);

CREATE INDEX IF NOT EXISTS idx_drivers_last_seen
    ON drivers(last_seen);
