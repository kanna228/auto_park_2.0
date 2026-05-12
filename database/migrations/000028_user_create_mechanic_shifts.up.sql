CREATE TABLE IF NOT EXISTS mechanic_shifts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
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

    CONSTRAINT chk_mechanic_shifts_time_range CHECK (time_to IS NULL OR time_to > time_from)
);

CREATE INDEX IF NOT EXISTS idx_mechanic_shifts_user_id
    ON mechanic_shifts(user_id);

CREATE INDEX IF NOT EXISTS idx_mechanic_shifts_shift_date
    ON mechanic_shifts(shift_date);

CREATE INDEX IF NOT EXISTS idx_mechanic_shifts_is_active
    ON mechanic_shifts(is_active);

CREATE INDEX IF NOT EXISTS idx_mechanic_shifts_is_deleted
    ON mechanic_shifts(is_deleted);

CREATE INDEX IF NOT EXISTS idx_mechanic_shifts_user_date
    ON mechanic_shifts(user_id, shift_date DESC);

-- Функция массово закрывает все смены, у которых уже прошло время окончания.
CREATE OR REPLACE FUNCTION refresh_mechanic_shifts_activity()
RETURNS void AS $$
BEGIN
    UPDATE mechanic_shifts
    SET is_active = FALSE,
        updated_at = NOW()
    WHERE is_deleted = FALSE
      AND is_active = TRUE
      AND time_to IS NOT NULL
      AND (shift_date::timestamp + time_to) <= (NOW() AT TIME ZONE current_setting('TimeZone'));
END;
$$ LANGUAGE plpgsql;

-- Функция автоматически выставляет is_active при INSERT/UPDATE конкретной строки.
CREATE OR REPLACE FUNCTION set_mechanic_shift_activity()
RETURNS trigger AS $$
BEGIN
    IF NEW.is_deleted = TRUE THEN
        NEW.is_active := FALSE;
    ELSIF NEW.time_to IS NOT NULL
       AND (NEW.shift_date::timestamp + NEW.time_to) <= (NOW() AT TIME ZONE current_setting('TimeZone')) THEN
        NEW.is_active := FALSE;
    ELSE
        NEW.is_active := TRUE;
    END IF;

    NEW.updated_at := NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_set_mechanic_shift_activity ON mechanic_shifts;

CREATE TRIGGER trg_set_mechanic_shift_activity
BEFORE INSERT OR UPDATE OF shift_date, time_from, time_to, is_deleted
ON mechanic_shifts
FOR EACH ROW
EXECUTE FUNCTION set_mechanic_shift_activity();