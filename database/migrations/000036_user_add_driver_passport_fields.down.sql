DROP INDEX IF EXISTS idx_drivers_license_number;
DROP INDEX IF EXISTS idx_drivers_birth_date;

ALTER TABLE drivers
    DROP CONSTRAINT IF EXISTS chk_drivers_experience_years_non_negative,
    DROP COLUMN IF EXISTS experience_years,
    DROP COLUMN IF EXISTS license_category,
    DROP COLUMN IF EXISTS license_number,
    DROP COLUMN IF EXISTS birth_date;
