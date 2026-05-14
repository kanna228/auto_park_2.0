CREATE TABLE IF NOT EXISTS parts_collection (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service_types (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vehicle_services (
    id BIGSERIAL PRIMARY KEY,
    type_id BIGINT NOT NULL REFERENCES service_types(id) ON DELETE RESTRICT,
    part_id BIGINT NOT NULL REFERENCES parts_collection(id) ON DELETE RESTRICT,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    service_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parts_collection_name
    ON parts_collection(name);

CREATE INDEX IF NOT EXISTS idx_service_types_name
    ON service_types(name);

CREATE INDEX IF NOT EXISTS idx_vehicle_services_type_id
    ON vehicle_services(type_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_services_part_id
    ON vehicle_services(part_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_services_vehicle_id
    ON vehicle_services(vehicle_id);

CREATE INDEX IF NOT EXISTS idx_vehicle_services_service_date
    ON vehicle_services(service_date DESC);

CREATE INDEX IF NOT EXISTS idx_vehicle_services_vehicle_date
    ON vehicle_services(vehicle_id, service_date DESC);

INSERT INTO parts_collection (name, description)
VALUES
    ('Передний бампер', 'Передняя наружная часть кузова автомобиля'),
    ('Задний бампер', 'Задняя наружная часть кузова автомобиля'),
    ('Переднее левое крыло', 'Левая передняя часть кузова над колесом'),
    ('Переднее правое крыло', 'Правая передняя часть кузова над колесом'),
    ('Заднее левое крыло', 'Левая задняя часть кузова над колесом'),
    ('Заднее правое крыло', 'Правая задняя часть кузова над колесом'),
    ('Передняя левая дверь', 'Левая передняя дверь автомобиля'),
    ('Передняя правая дверь', 'Правая передняя дверь автомобиля'),
    ('Задняя левая дверь', 'Левая задняя дверь автомобиля'),
    ('Задняя правая дверь', 'Правая задняя дверь автомобиля'),
    ('Капот', 'Передняя крышка моторного отсека'),
    ('Крыша', 'Верхняя часть кузова автомобиля'),
    ('Крышка багажника', 'Задняя крышка багажного отделения'),
    ('Лобовое стекло', 'Переднее стекло автомобиля'),
    ('Заднее стекло', 'Заднее стекло автомобиля'),
    ('Боковые стекла', 'Комплект боковых стекол'),
    ('Фары', 'Передние световые элементы'),
    ('Задние фонари', 'Задние световые элементы'),
    ('Боковые зеркала', 'Наружные зеркала заднего вида'),
    ('Весь кузов', 'Обслуживание всего кузова автомобиля')
ON CONFLICT (name) DO NOTHING;

INSERT INTO service_types (name, description)
VALUES
    ('Полировка', 'Восстановление блеска и удаление мелких дефектов поверхности'),
    ('Тонировка', 'Нанесение тонировочной пленки на стекла'),
    ('Нанесение защитной пленки', 'Оклейка части автомобиля защитной пленкой'),
    ('Антигравийная защита', 'Защита кузовных элементов от сколов и повреждений'),
    ('Химчистка', 'Очистка салона или отдельных элементов'),
    ('Детейлинг', 'Комплексное косметическое обслуживание автомобиля'),
    ('Покраска', 'Окрашивание элемента кузова'),
    ('Рихтовка', 'Восстановление геометрии кузовного элемента'),
    ('Мойка двигателя', 'Очистка моторного отсека'),
    ('Обработка антикором', 'Антикоррозийная обработка кузова или элементов')
ON CONFLICT (name) DO NOTHING;
