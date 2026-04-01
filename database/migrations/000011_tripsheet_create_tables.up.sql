-- no-op: using public schema

-- =========================
-- 1. Статусы путевых листов
-- =========================
CREATE TABLE IF NOT EXISTS tripsheet_statuses (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- =========================
-- 2. Маршрутный лист
-- =========================
CREATE TABLE IF NOT EXISTS tripsheets (
    id BIGSERIAL PRIMARY KEY,

    tripsheet_number VARCHAR(100) NOT NULL,
    tripsheet_date DATE NOT NULL,

    vehicle_brand VARCHAR(255),
    vehicle_plate_number VARCHAR(50) NOT NULL,

    driver_last_name VARCHAR(255),
    driver_first_name VARCHAR(255),
    driver_middle_name VARCHAR(255),
    driver_id BIGINT,

    start_time TIMESTAMP,
    end_time TIMESTAMP,

    mileage_start INT NOT NULL DEFAULT 0,
    mileage_end INT NOT NULL DEFAULT 0,

    fuel_start INT NOT NULL DEFAULT 0,
    fuel_issued INT NOT NULL DEFAULT 0,

    fuel_consumption_theoretical INT NOT NULL DEFAULT 0,
    fuel_consumption_actual INT NOT NULL DEFAULT 0,

    status_id BIGINT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tripsheets_status
        FOREIGN KEY (status_id)
        REFERENCES tripsheet_statuses(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_tripsheets_mileage_start_non_negative
        CHECK (mileage_start >= 0),

    CONSTRAINT chk_tripsheets_mileage_end_non_negative
        CHECK (mileage_end >= 0),

    CONSTRAINT chk_tripsheets_fuel_start_non_negative
        CHECK (fuel_start >= 0),

    CONSTRAINT chk_tripsheets_fuel_issued_non_negative
        CHECK (fuel_issued >= 0),

    CONSTRAINT chk_tripsheets_fuel_consumption_theoretical_non_negative
        CHECK (fuel_consumption_theoretical >= 0),

    CONSTRAINT chk_tripsheets_fuel_consumption_actual_non_negative
        CHECK (fuel_consumption_actual >= 0),

    CONSTRAINT chk_tripsheets_mileage_end_gte_start
        CHECK (mileage_end >= mileage_start)
);

-- Можно запретить дублирование номера путевого листа в пределах одной даты
CREATE UNIQUE INDEX IF NOT EXISTS ux_tripsheets_number_date
    ON tripsheets (tripsheet_number, tripsheet_date);

CREATE INDEX IF NOT EXISTS idx_tripsheets_driver_id
    ON tripsheets (driver_id);

CREATE INDEX IF NOT EXISTS idx_tripsheets_status_id
    ON tripsheets (status_id);

CREATE INDEX IF NOT EXISTS idx_tripsheets_vehicle_plate_number
    ON tripsheets (vehicle_plate_number);

CREATE INDEX IF NOT EXISTS idx_tripsheets_tripsheet_date
    ON tripsheets (tripsheet_date);

-- =========================
-- 3. Поездки внутри путевого листа
-- =========================
CREATE TABLE IF NOT EXISTS tripsheet_trips (
    id BIGSERIAL PRIMARY KEY,

    tripsheet_id BIGINT NOT NULL,
    route_description VARCHAR(1000) NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    distance_passed INT NOT NULL DEFAULT 0,
    status_id BIGINT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tripsheet_trips_tripsheet
        FOREIGN KEY (tripsheet_id)
        REFERENCES tripsheets(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tripsheet_trips_status
        FOREIGN KEY (status_id)
        REFERENCES tripsheet_statuses(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_tripsheet_trips_distance_non_negative
        CHECK (distance_passed >= 0)
);

CREATE INDEX IF NOT EXISTS idx_tripsheet_trips_tripsheet_id
    ON tripsheet_trips (tripsheet_id);

CREATE INDEX IF NOT EXISTS idx_tripsheet_trips_status_id
    ON tripsheet_trips (status_id);

-- =========================
-- Seed статусов
-- =========================
INSERT INTO tripsheet_statuses (name)
VALUES
    ('Создано'),
    ('В процессе'),
    ('В ожидании'),
    ('Окончен'),
    ('Отменен')
ON CONFLICT (name) DO NOTHING;