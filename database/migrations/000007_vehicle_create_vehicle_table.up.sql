
-- 2️⃣ Создаем таблицу автомобилей
CREATE TABLE IF NOT EXISTS vehicles (
    id BIGSERIAL PRIMARY KEY,

    board_number VARCHAR(50) NOT NULL,                         -- Борт номер
    technical_passport_number VARCHAR(100) NOT NULL,           -- Номер техпаспорта
    state_number VARCHAR(20) NOT NULL UNIQUE,                  -- Гос номер
    vin VARCHAR(50) NOT NULL UNIQUE,                           -- VIN

    brand_model VARCHAR(150) NOT NULL,                         -- Марка, модель
    manufacture_year INT NOT NULL CHECK (manufacture_year >= 1900),

    received_date DATE NOT NULL,                               -- Дата получения

    empty_weight_kg NUMERIC(10,2),                             -- Масса без нагрузки
    max_weight_kg NUMERIC(10,2),                                -- Максимальная масса
    engine_volume_cc INT,                                      -- Объем двигателя

    insurance_policy_number VARCHAR(100),                      -- № полиса
    insurance_expiry_date DATE,                                -- Срок действия полиса

    mileage BIGINT DEFAULT 0 CHECK (mileage >= 0),             -- Пробег
    current_fuel NUMERIC(10,2) DEFAULT 0 CHECK (current_fuel >= 0),  -- Текущее топливо (литры)

    drivers_ids BIGINT[] DEFAULT '{}',                         -- массив id водителей

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3️⃣ Индексы для ускорения поиска
CREATE INDEX idx_vehicles_board_number 
    ON vehicles(board_number);

CREATE INDEX idx_vehicles_drivers_ids 
    ON vehicles USING GIN (drivers_ids);