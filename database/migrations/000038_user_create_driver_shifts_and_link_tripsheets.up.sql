CREATE TABLE IF NOT EXISTS driver_shifts (
    id BIGSERIAL PRIMARY KEY,
    driver_id BIGINT NOT NULL REFERENCES drivers(id) ON DELETE RESTRICT,
    shift_date DATE NOT NULL,
    time_from TIME NOT NULL,
    time_to TIME NULL,
    comment TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ NULL,
    deleted_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_driver_shifts_time_range CHECK (time_to IS NULL OR time_to > time_from)
);

CREATE INDEX IF NOT EXISTS idx_driver_shifts_driver_id
    ON driver_shifts(driver_id);

CREATE INDEX IF NOT EXISTS idx_driver_shifts_shift_date
    ON driver_shifts(shift_date);

CREATE INDEX IF NOT EXISTS idx_driver_shifts_is_active
    ON driver_shifts(is_active);

CREATE INDEX IF NOT EXISTS idx_driver_shifts_is_deleted
    ON driver_shifts(is_deleted);

CREATE INDEX IF NOT EXISTS idx_driver_shifts_driver_date
    ON driver_shifts(driver_id, shift_date DESC);

CREATE OR REPLACE FUNCTION refresh_driver_shifts_activity()
RETURNS void AS $$
BEGIN
    UPDATE driver_shifts
    SET is_active = FALSE,
        updated_at = NOW()
    WHERE is_deleted = FALSE
      AND is_active = TRUE
      AND time_to IS NOT NULL
      AND (shift_date + time_to) <= (NOW() AT TIME ZONE current_setting('TimeZone'));
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION set_driver_shift_activity()
RETURNS trigger AS $$
BEGIN
    IF NEW.is_deleted = TRUE THEN
        NEW.is_active := FALSE;
    ELSIF NEW.time_to IS NOT NULL
       AND (NEW.shift_date + NEW.time_to) <= (NOW() AT TIME ZONE current_setting('TimeZone')) THEN
        NEW.is_active := FALSE;
    ELSE
        NEW.is_active := COALESCE(NEW.is_active, TRUE);
    END IF;

    NEW.updated_at := NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_set_driver_shift_activity ON driver_shifts;

CREATE TRIGGER trg_set_driver_shift_activity
BEFORE INSERT OR UPDATE OF shift_date, time_from, time_to, is_active, is_deleted
ON driver_shifts
FOR EACH ROW
EXECUTE FUNCTION set_driver_shift_activity();

ALTER TABLE tripsheets
    ADD COLUMN IF NOT EXISTS driver_shift_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_tripsheets_driver_shift'
          AND table_name = 'tripsheets'
    ) THEN
        ALTER TABLE tripsheets
            ADD CONSTRAINT fk_tripsheets_driver_shift
            FOREIGN KEY (driver_shift_id)
            REFERENCES driver_shifts(id)
            ON DELETE SET NULL;
    END IF;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_tripsheets_driver_shift_id
    ON tripsheets(driver_shift_id);

-- Best-effort backfill for existing tripsheets if driver shifts already exist.
-- It only fills rows where driver_shift_id is NULL and there is a clear matching driver/date/time.
UPDATE tripsheets t
SET driver_shift_id = ds.id
FROM driver_shifts ds
WHERE t.driver_shift_id IS NULL
  AND t.driver_id IS NOT NULL
  AND ds.is_deleted = FALSE
  AND ds.driver_id = t.driver_id
  AND ds.shift_date = t.tripsheet_date
  AND (
        t.start_time IS NULL
        OR t.start_time::time >= ds.time_from
      )
  AND (
        ds.time_to IS NULL
        OR t.end_time IS NULL
        OR t.end_time::time <= ds.time_to
      );
