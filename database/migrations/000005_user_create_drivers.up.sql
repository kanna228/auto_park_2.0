-- 00005_create_drivers_table.up.sql

CREATE TABLE IF NOT EXISTS drivers (
    id          BIGSERIAL PRIMARY KEY,
    iin         VARCHAR(12) NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    surname     TEXT NOT NULL,
    middlename  TEXT,
    phone       TEXT,
    mail        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Полезные индексы (поиск)
CREATE INDEX IF NOT EXISTS idx_drivers_iin ON drivers (iin);
CREATE INDEX IF NOT EXISTS idx_drivers_surname_name ON drivers (surname, name);
