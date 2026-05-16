CREATE TABLE IF NOT EXISTS maintenance_schedules (
    id BIGSERIAL PRIMARY KEY,
    is_draft BOOLEAN NOT NULL DEFAULT TRUE,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    consecutive_count INT NOT NULL DEFAULT 1 CHECK (consecutive_count > 0),
    consecutive_unit VARCHAR(16) NOT NULL,
    every_count INT NOT NULL DEFAULT 1 CHECK (every_count > 0),
    every_unit VARCHAR(16) NOT NULL,
    time_from VARCHAR(5) NOT NULL,
    time_to VARCHAR(5) NOT NULL,
    duration_value INT NOT NULL CHECK (duration_value > 0),
    duration_unit VARCHAR(16) NOT NULL,
    limit_boards_by_time BOOLEAN NOT NULL DEFAULT FALSE,
    categories JSONB NOT NULL DEFAULT '[]'::jsonb,
    boards JSONB NOT NULL DEFAULT '[]'::jsonb,
    mechanics JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_is_draft
    ON maintenance_schedules(is_draft);

CREATE INDEX IF NOT EXISTS idx_maintenance_schedules_date_from
    ON maintenance_schedules(date_from DESC);
