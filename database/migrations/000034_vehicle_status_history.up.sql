CREATE TABLE IF NOT EXISTS vehicle_status_history (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    status_id BIGINT NOT NULL REFERENCES vehicle_status(id) ON DELETE RESTRICT,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_vehicle_status_history_dates
        CHECK (end_date IS NULL OR end_date > start_date)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_status_history_vehicle_id
    ON vehicle_status_history(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_status_history_status_id
    ON vehicle_status_history(status_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_status_history_start_date
    ON vehicle_status_history(start_date DESC);

CREATE INDEX IF NOT EXISTS idx_vehicle_status_history_end_date
    ON vehicle_status_history(end_date DESC);

CREATE UNIQUE INDEX IF NOT EXISTS ux_vehicle_status_history_current
    ON vehicle_status_history(vehicle_id)
    WHERE end_date IS NULL;

INSERT INTO vehicle_status_history (
    vehicle_id,
    status_id,
    start_date,
    end_date,
    created_at,
    updated_at
)
SELECT
    v.id,
    v.status_id,
    COALESCE(v.created_at, NOW())::timestamptz,
    NULL,
    NOW(),
    NOW()
FROM vehicles v
WHERE v.status_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM vehicle_status_history h
      WHERE h.vehicle_id = v.id
        AND h.end_date IS NULL
  );

CREATE OR REPLACE FUNCTION sync_vehicle_status_history()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO vehicle_status_history (
            vehicle_id,
            status_id,
            start_date,
            end_date,
            created_at,
            updated_at
        )
        VALUES (
            NEW.id,
            NEW.status_id,
            COALESCE(NEW.created_at, NOW())::timestamptz,
            NULL,
            NOW(),
            NOW()
        )
        ON CONFLICT DO NOTHING;

        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' AND OLD.status_id IS DISTINCT FROM NEW.status_id THEN
        UPDATE vehicle_status_history
        SET end_date = NOW(),
            updated_at = NOW()
        WHERE vehicle_id = NEW.id
          AND end_date IS NULL;

        INSERT INTO vehicle_status_history (
            vehicle_id,
            status_id,
            start_date,
            end_date,
            created_at,
            updated_at
        )
        VALUES (
            NEW.id,
            NEW.status_id,
            NOW(),
            NULL,
            NOW(),
            NOW()
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sync_vehicle_status_history ON vehicles;

CREATE TRIGGER trg_sync_vehicle_status_history
AFTER INSERT OR UPDATE OF status_id
ON vehicles
FOR EACH ROW
EXECUTE FUNCTION sync_vehicle_status_history();
