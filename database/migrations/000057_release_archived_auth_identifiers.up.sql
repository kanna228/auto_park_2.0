UPDATE users
SET email = CONCAT('__archived_user_', id, '@auto-park.local'),
    iin = CONCAT('ARCH', id),
    updated_at = NOW()
WHERE is_archived = TRUE
  AND (
      email <> CONCAT('__archived_user_', id, '@auto-park.local')
      OR iin <> CONCAT('ARCH', id)
  );

UPDATE drivers
SET mail = CASE
        WHEN mail IS NULL OR BTRIM(mail) = '' THEN mail
        ELSE CONCAT('__archived_driver_', id, '@auto-park.local')
    END,
    iin = CONCAT('X', LPAD(id::TEXT, 11, '0')),
    updated_at = NOW()
WHERE is_archived = TRUE
  AND (
      (mail IS NOT NULL AND BTRIM(mail) <> '' AND mail <> CONCAT('__archived_driver_', id, '@auto-park.local'))
      OR iin <> CONCAT('X', LPAD(id::TEXT, 11, '0'))
  );
