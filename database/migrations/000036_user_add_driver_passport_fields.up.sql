ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS birth_date DATE NULL,
    ADD COLUMN IF NOT EXISTS license_number TEXT NULL,
    ADD COLUMN IF NOT EXISTS license_category TEXT NULL,
    ADD COLUMN IF NOT EXISTS experience_years INT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_drivers_experience_years_non_negative'
    ) THEN
        ALTER TABLE drivers
            ADD CONSTRAINT chk_drivers_experience_years_non_negative
            CHECK (experience_years IS NULL OR experience_years >= 0);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_drivers_birth_date
    ON drivers(birth_date);

CREATE INDEX IF NOT EXISTS idx_drivers_license_number
    ON drivers(license_number);
