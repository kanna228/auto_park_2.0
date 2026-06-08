-- Realistic demo data for presentation and local testing.
-- The seed is intentionally capped at 20 rows per business table.

INSERT INTO users (
    email, first_name, last_name, middle_name, iin, phone, password, role_id, last_seen
)
SELECT
    v.email,
    v.first_name,
    v.last_name,
    v.middle_name,
    v.iin,
    v.phone,
    '$2a$12$93CmzSxMMK.jo6YfoXs5G.rneemNhlMu6DHuYPV3QtX6OQa3JKLme',
    r.id,
    v.last_seen::timestamptz
FROM (VALUES
    ('manager@autopark.demo', 'Айдар', 'Сулейменов', 'Ерланович', '830114300101', '+77011234501', 'manager', '2026-05-25 08:45:00+05'),
    ('dispatcher@autopark.demo', 'Марина', 'Ким', 'Викторовна', '900327400102', '+77021234502', 'garage_dispatcher', '2026-05-25 09:10:00+05'),
    ('mechanic1@autopark.demo', 'Руслан', 'Абдрахманов', 'Серикович', '870602300103', '+77031234503', 'duty_mechanic', '2026-05-25 07:55:00+05'),
    ('mechanic2@autopark.demo', 'Николай', 'Петров', 'Андреевич', '820918300104', '+77041234504', 'duty_mechanic', '2026-05-24 18:30:00+05'),
    ('mechanic3@autopark.demo', 'Ермек', 'Тлеубаев', 'Муратович', '790405300105', '+77051234505', 'duty_mechanic', '2026-05-25 08:05:00+05'),
    ('mechanic4@autopark.demo', 'Данияр', 'Касенов', 'Бакытович', '880711300106', '+77061234506', 'duty_mechanic', '2026-05-23 16:45:00+05'),
    ('warehouse@autopark.demo', 'Ольга', 'Романова', 'Сергеевна', '860222400107', '+77071234507', 'warehouse_manager', '2026-05-25 08:20:00+05'),
    ('warehouse2@autopark.demo', 'Асель', 'Нурпеисова', 'Талгатовна', '920815400108', '+77081234508', 'warehouse_manager', '2026-05-24 17:05:00+05'),
    ('dispatcher2@autopark.demo', 'Тимур', 'Ахметов', 'Нурланович', '910509300109', '+77091234509', 'garage_dispatcher', '2026-05-25 06:50:00+05'),
    ('manager2@autopark.demo', 'Елена', 'Васильева', 'Игоревна', '840120400110', '+77101234510', 'manager', '2026-05-22 14:15:00+05')
) AS v(email, first_name, last_name, middle_name, iin, phone, role_name, last_seen)
JOIN roles r ON r.name = v.role_name
ON CONFLICT (email) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    middle_name = EXCLUDED.middle_name,
    phone = EXCLUDED.phone,
    role_id = EXCLUDED.role_id,
    last_seen = EXCLUDED.last_seen,
    updated_at = NOW();

INSERT INTO drivers (
    iin, name, surname, middlename, phone, mail, photo_path,
    birth_date, license_number, license_category, experience_years, status_id, status
)
SELECT
    v.iin,
    v.name,
    v.surname,
    v.middlename,
    v.phone,
    v.mail,
    v.photo_path,
    v.birth_date::date,
    v.license_number,
    v.license_category,
    v.experience_years,
    ds.id,
    v.status
FROM (VALUES
    ('850101300001', 'Арман', 'Ибраев', 'Канатович', '+77010001001', 'a.ibraev@autopark.demo', '/storage/drivers/arman-ibraev.jpg', '1985-01-01', 'KZ-ALM-118234', 'B,C', 17, 'available'),
    ('870214300002', 'Сергей', 'Мельников', 'Павлович', '+77010001002', 's.melnikov@autopark.demo', '/storage/drivers/sergey-melnikov.jpg', '1987-02-14', 'KZ-ALM-118235', 'B,C,D', 15, 'on_trip'),
    ('900305300003', 'Бауыржан', 'Омаров', 'Ермекович', '+77010001003', 'b.omarov@autopark.demo', '/storage/drivers/baurzhan-omarov.jpg', '1990-03-05', 'KZ-ALM-118236', 'B,C', 11, 'available'),
    ('820416300004', 'Алексей', 'Сидоров', 'Иванович', '+77010001004', 'a.sidorov@autopark.demo', '/storage/drivers/alexey-sidorov.jpg', '1982-04-16', 'KZ-AST-214551', 'B,C,E', 20, 'available'),
    ('760527300005', 'Мурат', 'Нуртаев', 'Сагатович', '+77010001005', 'm.nurtaev@autopark.demo', '/storage/drivers/murat-nurtaev.jpg', '1976-05-27', 'KZ-AST-214552', 'B,C,D', 24, 'inactive'),
    ('930608300006', 'Павел', 'Ковалев', 'Олегович', '+77010001006', 'p.kovalev@autopark.demo', '/storage/drivers/pavel-kovalev.jpg', '1993-06-08', 'KZ-KRG-334120', 'B,C', 8, 'available'),
    ('880719300007', 'Ерлан', 'Смагулов', 'Рахатович', '+77010001007', 'e.smagulov@autopark.demo', '/storage/drivers/erlan-smagulov.jpg', '1988-07-19', 'KZ-KRG-334121', 'B,C', 13, 'on_trip'),
    ('810830300008', 'Дмитрий', 'Новиков', 'Александрович', '+77010001008', 'd.novikov@autopark.demo', '/storage/drivers/dmitry-novikov.jpg', '1981-08-30', 'KZ-PVL-410902', 'B,C,E', 22, 'available'),
    ('951011300009', 'Асхат', 'Жумабаев', 'Талгатович', '+77010001009', 'a.zhumabaev@autopark.demo', '/storage/drivers/askhat-zhumabaev.jpg', '1995-10-11', 'KZ-PVL-410903', 'B', 6, 'available'),
    ('790922300010', 'Владимир', 'Кузнецов', 'Николаевич', '+77010001010', 'v.kuznetsov@autopark.demo', '/storage/drivers/vladimir-kuznetsov.jpg', '1979-09-22', 'KZ-SHY-509001', 'B,C,D,E', 25, 'available'),
    ('910103300011', 'Мади', 'Турсунов', 'Айбекович', '+77010001011', 'm.tursunov@autopark.demo', '/storage/drivers/madi-tursunov.jpg', '1991-01-03', 'KZ-SHY-509002', 'B,C', 10, 'on_trip'),
    ('841224300012', 'Игорь', 'Лебедев', 'Валерьевич', '+77010001012', 'i.lebedev@autopark.demo', '/storage/drivers/igor-lebedev.jpg', '1984-12-24', 'KZ-TLD-612004', 'B,C', 18, 'available'),
    ('960415300013', 'Нуржан', 'Рахимов', 'Серикович', '+77010001013', 'n.rakhimov@autopark.demo', '/storage/drivers/nurzhan-rakhimov.jpg', '1996-04-15', 'KZ-TLD-612005', 'B', 5, 'available'),
    ('830706300014', 'Роман', 'Федоров', 'Михайлович', '+77010001014', 'r.fedorov@autopark.demo', '/storage/drivers/roman-fedorov.jpg', '1983-07-06', 'KZ-ALM-118237', 'B,C,D', 19, 'available'),
    ('890217300015', 'Алишер', 'Каримов', 'Русланович', '+77010001015', 'a.karimov@autopark.demo', '/storage/drivers/alisher-karimov.jpg', '1989-02-17', 'KZ-ALM-118238', 'B,C', 12, 'inactive'),
    ('920328300016', 'Виктор', 'Соколов', 'Петрович', '+77010001016', 'v.sokolov@autopark.demo', '/storage/drivers/viktor-sokolov.jpg', '1992-03-28', 'KZ-AST-214553', 'B,C', 9, 'available'),
    ('780509300017', 'Кайрат', 'Есенов', 'Маратович', '+77010001017', 'k.esenov@autopark.demo', '/storage/drivers/kairat-esenov.jpg', '1978-05-09', 'KZ-AKT-730118', 'B,C,E', 26, 'available'),
    ('940620300018', 'Антон', 'Григорьев', 'Денисович', '+77010001018', 'a.grigorev@autopark.demo', '/storage/drivers/anton-grigorev.jpg', '1994-06-20', 'KZ-AKT-730119', 'B', 7, 'available'),
    ('800731300019', 'Самат', 'Бекенов', 'Оразович', '+77010001019', 's.bekenov@autopark.demo', '/storage/drivers/samat-bekenov.jpg', '1980-07-31', 'KZ-URA-881245', 'B,C,D', 23, 'on_trip'),
    ('971112300020', 'Руслан', 'Галиев', 'Маратович', '+77010001020', 'r.galiev@autopark.demo', '/storage/drivers/ruslan-galiev.jpg', '1997-11-12', 'KZ-URA-881246', 'B,C', 4, 'available')
) AS v(iin, name, surname, middlename, phone, mail, photo_path, birth_date, license_number, license_category, experience_years, status)
JOIN driver_statuses ds ON ds.code = v.status
ON CONFLICT (iin) DO UPDATE SET
    name = EXCLUDED.name,
    surname = EXCLUDED.surname,
    middlename = EXCLUDED.middlename,
    phone = EXCLUDED.phone,
    mail = EXCLUDED.mail,
    photo_path = EXCLUDED.photo_path,
    birth_date = EXCLUDED.birth_date,
    license_number = EXCLUDED.license_number,
    license_category = EXCLUDED.license_category,
    experience_years = EXCLUDED.experience_years,
    status_id = EXCLUDED.status_id,
    status = EXCLUDED.status,
    updated_at = NOW();

INSERT INTO vehicles (
    board_number, technical_passport_number, state_number, vin, brand_model,
    manufacture_year, received_date, empty_weight_kg, max_weight_kg, engine_volume_cc,
    insurance_policy_number, insurance_expiry_date, mileage, current_fuel, photo_path, status_id
)
SELECT
    v.board_number,
    v.technical_passport_number,
    v.state_number,
    v.vin,
    v.brand_model,
    v.manufacture_year,
    v.received_date::date,
    v.empty_weight_kg,
    v.max_weight_kg,
    v.engine_volume_cc,
    v.insurance_policy_number,
    v.insurance_expiry_date::date,
    v.mileage,
    v.current_fuel,
    v.photo_path,
    vs.id
FROM (VALUES
    ('B-101', 'TP-ALM-240101', '101ABC01', 'XTA210990K0000001', 'Toyota Hiace 2.8D', 2021, '2021-04-12', 2140.00, 3300.00, 2755, 'INS-AP-2026-101', '2026-12-31', 84210, 52.50, '/storage/vehicles/b-101.jpg', 'В использовании'),
    ('B-102', 'TP-ALM-240102', '102ABC01', 'XTA210990K0000002', 'Hyundai H350 Cargo', 2020, '2020-10-05', 2480.00, 3500.00, 2497, 'INS-AP-2026-102', '2026-11-20', 116430, 61.00, '/storage/vehicles/b-102.jpg', 'В использовании'),
    ('B-103', 'TP-ALM-240103', '103ABC01', 'XTA210990K0000003', 'GAZelle Next A32R32', 2019, '2019-08-19', 2360.00, 3500.00, 2800, 'INS-AP-2026-103', '2026-10-14', 148900, 43.75, '/storage/vehicles/b-103.jpg', 'На ТО'),
    ('B-104', 'TP-ALM-240104', '104ABC01', 'XTA210990K0000004', 'Mercedes-Benz Sprinter 316 CDI', 2022, '2022-02-03', 2425.00, 3500.00, 2143, 'INS-AP-2026-104', '2027-01-18', 63980, 70.20, '/storage/vehicles/b-104.jpg', 'В использовании'),
    ('B-105', 'TP-ALM-240105', '105ABC01', 'XTA210990K0000005', 'Ford Transit L3H2', 2021, '2021-09-15', 2320.00, 3500.00, 2198, 'INS-AP-2026-105', '2026-09-30', 95870, 55.00, '/storage/vehicles/b-105.jpg', 'На ремонте'),
    ('B-106', 'TP-AST-240106', '106ABC01', 'XTA210990K0000006', 'Isuzu NPR 75', 2018, '2018-06-25', 4100.00, 7500.00, 5193, 'INS-AP-2026-106', '2026-08-22', 202450, 92.40, '/storage/vehicles/b-106.jpg', 'В использовании'),
    ('B-107', 'TP-AST-240107', '107ABC01', 'XTA210990K0000007', 'Hino 300 XZU', 2020, '2020-05-11', 3860.00, 7500.00, 4009, 'INS-AP-2026-107', '2026-12-05', 132740, 88.10, '/storage/vehicles/b-107.jpg', 'В использовании'),
    ('B-108', 'TP-AST-240108', '108ABC01', 'XTA210990K0000008', 'Kia K2700', 2019, '2019-12-17', 1850.00, 2800.00, 2665, 'INS-AP-2026-108', '2026-07-19', 120300, 38.60, '/storage/vehicles/b-108.jpg', 'В использовании'),
    ('B-109', 'TP-KRG-240109', '109ABC01', 'XTA210990K0000009', 'Toyota Land Cruiser Prado 150', 2022, '2022-07-21', 2185.00, 2990.00, 2755, 'INS-AP-2026-109', '2027-02-10', 51420, 74.30, '/storage/vehicles/b-109.jpg', 'В использовании'),
    ('B-110', 'TP-KRG-240110', '110ABC01', 'XTA210990K0000010', 'Mitsubishi L200', 2021, '2021-11-02', 1935.00, 2850.00, 2442, 'INS-AP-2026-110', '2026-10-28', 87650, 49.10, '/storage/vehicles/b-110.jpg', 'В использовании'),
    ('B-111', 'TP-PVL-240111', '111ABC01', 'XTA210990K0000011', 'Volkswagen Crafter 35', 2020, '2020-03-30', 2530.00, 3500.00, 1968, 'INS-AP-2026-111', '2026-09-18', 139880, 64.00, '/storage/vehicles/b-111.jpg', 'На ТО'),
    ('B-112', 'TP-PVL-240112', '112ABC01', 'XTA210990K0000012', 'Renault Master L2H2', 2019, '2019-04-09', 2260.00, 3500.00, 2299, 'INS-AP-2026-112', '2026-06-30', 154210, 58.20, '/storage/vehicles/b-112.jpg', 'В использовании'),
    ('B-113', 'TP-SHY-240113', '113ABC01', 'XTA210990K0000013', 'MAN TGL 8.180', 2018, '2018-09-12', 4980.00, 8600.00, 4580, 'INS-AP-2026-113', '2026-11-11', 221340, 115.00, '/storage/vehicles/b-113.jpg', 'В использовании'),
    ('B-114', 'TP-SHY-240114', '114ABC01', 'XTA210990K0000014', 'KamAZ 4308', 2017, '2017-05-24', 5850.00, 11500.00, 6700, 'INS-AP-2026-114', '2026-05-29', 264900, 130.50, '/storage/vehicles/b-114.jpg', 'На ремонте'),
    ('B-115', 'TP-TLD-240115', '115ABC01', 'XTA210990K0000015', 'Lada Largus Van', 2021, '2021-01-18', 1370.00, 2010.00, 1596, 'INS-AP-2026-115', '2026-12-24', 79230, 31.20, '/storage/vehicles/b-115.jpg', 'В использовании'),
    ('B-116', 'TP-TLD-240116', '116ABC01', 'XTA210990K0000016', 'Skoda Octavia', 2022, '2022-06-06', 1395.00, 1940.00, 1395, 'INS-AP-2026-116', '2027-03-04', 45780, 28.40, '/storage/vehicles/b-116.jpg', 'В использовании'),
    ('B-117', 'TP-AKT-240117', '117ABC01', 'XTA210990K0000017', 'UAZ Profi', 2020, '2020-08-13', 1990.00, 3500.00, 2693, 'INS-AP-2026-117', '2026-08-31', 110540, 46.90, '/storage/vehicles/b-117.jpg', 'Не используется'),
    ('B-118', 'TP-AKT-240118', '118ABC01', 'XTA210990K0000018', 'Hyundai Staria', 2023, '2023-04-20', 2210.00, 3030.00, 2199, 'INS-AP-2026-118', '2027-04-20', 21980, 67.00, '/storage/vehicles/b-118.jpg', 'В использовании'),
    ('B-119', 'TP-URA-240119', '119ABC01', 'XTA210990K0000019', 'Chevrolet Cobalt', 2021, '2021-12-01', 1140.00, 1570.00, 1485, 'INS-AP-2026-119', '2026-10-02', 68320, 24.70, '/storage/vehicles/b-119.jpg', 'В использовании'),
    ('B-120', 'TP-URA-240120', '120ABC01', 'XTA210990K0000020', 'Toyota Camry 70', 2022, '2022-09-09', 1570.00, 2030.00, 2487, 'INS-AP-2026-120', '2027-01-12', 39210, 42.80, '/storage/vehicles/b-120.jpg', 'В использовании')
) AS v(board_number, technical_passport_number, state_number, vin, brand_model, manufacture_year, received_date, empty_weight_kg, max_weight_kg, engine_volume_cc, insurance_policy_number, insurance_expiry_date, mileage, current_fuel, photo_path, status_name)
JOIN vehicle_status vs ON vs.name = v.status_name
ON CONFLICT (state_number) DO UPDATE SET
    board_number = EXCLUDED.board_number,
    technical_passport_number = EXCLUDED.technical_passport_number,
    brand_model = EXCLUDED.brand_model,
    manufacture_year = EXCLUDED.manufacture_year,
    received_date = EXCLUDED.received_date,
    empty_weight_kg = EXCLUDED.empty_weight_kg,
    max_weight_kg = EXCLUDED.max_weight_kg,
    engine_volume_cc = EXCLUDED.engine_volume_cc,
    insurance_policy_number = EXCLUDED.insurance_policy_number,
    insurance_expiry_date = EXCLUDED.insurance_expiry_date,
    mileage = EXCLUDED.mileage,
    current_fuel = EXCLUDED.current_fuel,
    photo_path = EXCLUDED.photo_path,
    status_id = EXCLUDED.status_id,
    updated_at = NOW();

WITH assignments(board_number, driver_iins) AS (
    VALUES
        ('B-101', ARRAY['850101300001', '870214300002']),
        ('B-102', ARRAY['900305300003']),
        ('B-103', ARRAY['820416300004']),
        ('B-104', ARRAY['760527300005', '930608300006']),
        ('B-105', ARRAY['880719300007']),
        ('B-106', ARRAY['810830300008']),
        ('B-107', ARRAY['951011300009', '790922300010']),
        ('B-108', ARRAY['910103300011']),
        ('B-109', ARRAY['841224300012']),
        ('B-110', ARRAY['960415300013']),
        ('B-111', ARRAY['830706300014']),
        ('B-112', ARRAY['890217300015']),
        ('B-113', ARRAY['920328300016']),
        ('B-114', ARRAY['780509300017']),
        ('B-115', ARRAY['940620300018']),
        ('B-116', ARRAY['800731300019']),
        ('B-117', ARRAY['971112300020']),
        ('B-118', ARRAY['850101300001']),
        ('B-119', ARRAY['900305300003']),
        ('B-120', ARRAY['841224300012'])
)
UPDATE vehicles v
SET drivers_ids = (
        SELECT ARRAY_AGG(d.id ORDER BY d.id)
        FROM drivers d
        WHERE d.iin = ANY(a.driver_iins)
    ),
    updated_at = NOW()
FROM assignments a
WHERE v.board_number = a.board_number;

INSERT INTO mechanic_shifts (user_id, shift_date, time_from, time_to, comment)
SELECT u.id, v.shift_date::date, v.time_from::time, v.time_to::time, v.comment
FROM (VALUES
    ('mechanic1@autopark.demo', '2026-05-21', '08:00', '20:00', 'Плановая дневная смена зоны ТО N1'),
    ('mechanic2@autopark.demo', '2026-05-21', '20:00', '23:59', 'Вечерняя смена диагностики'),
    ('mechanic3@autopark.demo', '2026-05-22', '08:00', '20:00', 'Дневная смена ремонта ходовой'),
    ('mechanic4@autopark.demo', '2026-05-22', '09:00', '18:00', 'Смена по выпуску транспорта на линию'),
    ('mechanic1@autopark.demo', '2026-05-23', '08:00', '20:00', 'Плановая дневная смена склада шин'),
    ('mechanic2@autopark.demo', '2026-05-23', '20:00', '23:59', 'Дежурство по аварийным заявкам'),
    ('mechanic3@autopark.demo', '2026-05-24', '08:00', '20:00', 'Проверка тормозных систем'),
    ('mechanic4@autopark.demo', '2026-05-24', '09:00', '18:00', 'Подготовка машин к понедельнику'),
    ('mechanic1@autopark.demo', '2026-05-25', '07:00', NULL, 'Активная смена линии выпуска'),
    ('mechanic3@autopark.demo', '2026-05-25', '08:00', NULL, 'Активная смена по ремонту')
) AS v(email, shift_date, time_from, time_to, comment)
JOIN users u ON u.email = v.email
WHERE NOT EXISTS (
    SELECT 1 FROM mechanic_shifts ms
    WHERE ms.user_id = u.id AND ms.shift_date = v.shift_date::date AND ms.comment = v.comment
);

INSERT INTO driver_shifts (driver_id, shift_date, time_from, time_to, comment)
SELECT d.id, v.shift_date::date, v.time_from::time, v.time_to::time, v.comment
FROM (VALUES
    ('850101300001', '2026-05-06', '08:00', '17:30', 'Городской маршрут: склады и центральный офис'),
    ('870214300002', '2026-05-07', '07:30', '18:00', 'Доставка документов и оборудования по Алматы'),
    ('900305300003', '2026-05-08', '08:00', '17:00', 'Плановый рейс по сервисным точкам'),
    ('820416300004', '2026-05-09', '09:00', '18:30', 'Межрайонный рейс с возвратом в парк'),
    ('760527300005', '2026-05-10', '08:00', '16:45', 'Резервный выезд по заявкам диспетчера'),
    ('930608300006', '2026-05-11', '07:45', '18:15', 'Поставка расходников на филиалы'),
    ('880719300007', '2026-05-12', '08:15', '17:40', 'Маршрут с заездом на АЗС и склад'),
    ('810830300008', '2026-05-13', '08:00', '19:00', 'Грузовой рейс до промышленной зоны'),
    ('951011300009', '2026-05-14', '09:00', '18:00', 'Курьерский маршрут по административным адресам'),
    ('790922300010', '2026-05-15', '07:00', '17:00', 'Дальний рейс в пригородную зону'),
    ('910103300011', '2026-05-16', '08:30', '18:30', 'Плановая перевозка оборудования'),
    ('841224300012', '2026-05-17', '08:00', '17:15', 'Маршрут по медицинским учреждениям'),
    ('960415300013', '2026-05-18', '09:00', '18:00', 'Подмена на легковом автомобиле'),
    ('830706300014', '2026-05-19', '07:30', '17:30', 'Снабжение объектов северного направления'),
    ('890217300015', '2026-05-20', '08:00', '16:30', 'Короткий маршрут до сервисной базы'),
    ('920328300016', '2026-05-21', '08:00', '18:00', 'Грузовой рейс с погрузкой на складе'),
    ('780509300017', '2026-05-22', '07:30', '17:45', 'Маршрут по производственным площадкам'),
    ('940620300018', '2026-05-23', '09:00', '18:20', 'Доставка сотрудников и документов'),
    ('800731300019', '2026-05-24', '08:00', '19:00', 'Рабочий выезд с несколькими точками'),
    ('971112300020', '2026-05-25', '07:30', NULL, 'Активный рейс текущего дня')
) AS v(iin, shift_date, time_from, time_to, comment)
JOIN drivers d ON d.iin = v.iin
WHERE NOT EXISTS (
    SELECT 1 FROM driver_shifts ds
    WHERE ds.driver_id = d.id AND ds.shift_date = v.shift_date::date AND ds.comment = v.comment
);

INSERT INTO insurance (vehicle_id, name, start_date, end_date, file_path, is_active)
SELECT veh.id, v.name, v.start_date::date, v.end_date::date, v.file_path, TRUE
FROM (VALUES
    ('B-101', 'Полис ОГПО ТС B-101 / 2026', '2026-01-01', '2026-12-31', '/storage/docs/insurance/b-101-2026.pdf'),
    ('B-102', 'Полис ОГПО ТС B-102 / 2026', '2025-11-20', '2026-11-20', '/storage/docs/insurance/b-102-2026.pdf'),
    ('B-103', 'Полис ОГПО ТС B-103 / 2026', '2025-10-14', '2026-10-14', '/storage/docs/insurance/b-103-2026.pdf'),
    ('B-104', 'Полис ОГПО ТС B-104 / 2026', '2026-01-18', '2027-01-18', '/storage/docs/insurance/b-104-2027.pdf'),
    ('B-105', 'Полис ОГПО ТС B-105 / 2026', '2025-09-30', '2026-09-30', '/storage/docs/insurance/b-105-2026.pdf'),
    ('B-106', 'Полис ОГПО ТС B-106 / 2026', '2025-08-22', '2026-08-22', '/storage/docs/insurance/b-106-2026.pdf'),
    ('B-107', 'Полис ОГПО ТС B-107 / 2026', '2025-12-05', '2026-12-05', '/storage/docs/insurance/b-107-2026.pdf'),
    ('B-108', 'Полис ОГПО ТС B-108 / 2026', '2025-07-19', '2026-07-19', '/storage/docs/insurance/b-108-2026.pdf'),
    ('B-109', 'Полис ОГПО ТС B-109 / 2026', '2026-02-10', '2027-02-10', '/storage/docs/insurance/b-109-2027.pdf'),
    ('B-110', 'Полис ОГПО ТС B-110 / 2026', '2025-10-28', '2026-10-28', '/storage/docs/insurance/b-110-2026.pdf'),
    ('B-111', 'Полис ОГПО ТС B-111 / 2026', '2025-09-18', '2026-09-18', '/storage/docs/insurance/b-111-2026.pdf'),
    ('B-112', 'Полис ОГПО ТС B-112 / 2026', '2025-06-30', '2026-06-30', '/storage/docs/insurance/b-112-2026.pdf'),
    ('B-113', 'Полис ОГПО ТС B-113 / 2026', '2025-11-11', '2026-11-11', '/storage/docs/insurance/b-113-2026.pdf'),
    ('B-114', 'Полис ОГПО ТС B-114 / 2026', '2025-05-29', '2026-05-29', '/storage/docs/insurance/b-114-2026.pdf'),
    ('B-115', 'Полис ОГПО ТС B-115 / 2026', '2025-12-24', '2026-12-24', '/storage/docs/insurance/b-115-2026.pdf'),
    ('B-116', 'Полис ОГПО ТС B-116 / 2026', '2026-03-04', '2027-03-04', '/storage/docs/insurance/b-116-2027.pdf'),
    ('B-117', 'Полис ОГПО ТС B-117 / 2026', '2025-08-31', '2026-08-31', '/storage/docs/insurance/b-117-2026.pdf'),
    ('B-118', 'Полис ОГПО ТС B-118 / 2026', '2026-04-20', '2027-04-20', '/storage/docs/insurance/b-118-2027.pdf'),
    ('B-119', 'Полис ОГПО ТС B-119 / 2026', '2025-10-02', '2026-10-02', '/storage/docs/insurance/b-119-2026.pdf'),
    ('B-120', 'Полис ОГПО ТС B-120 / 2026', '2026-01-12', '2027-01-12', '/storage/docs/insurance/b-120-2027.pdf')
) AS v(board_number, name, start_date, end_date, file_path)
JOIN vehicles veh ON veh.board_number = v.board_number
WHERE NOT EXISTS (
    SELECT 1 FROM insurance i WHERE i.vehicle_id = veh.id AND i.name = v.name
);

INSERT INTO technical_inspection (vehicle_id, name, start_date, end_date, file_path, is_active)
SELECT veh.id, v.name, v.start_date::date, v.end_date::date, v.file_path, TRUE
FROM (VALUES
    ('B-101', 'Техосмотр B-101 / 2026', '2026-02-01', '2027-02-01', '/storage/docs/inspection/b-101-2026.pdf'),
    ('B-102', 'Техосмотр B-102 / 2026', '2026-02-03', '2027-02-03', '/storage/docs/inspection/b-102-2026.pdf'),
    ('B-103', 'Техосмотр B-103 / 2026', '2026-02-05', '2027-02-05', '/storage/docs/inspection/b-103-2026.pdf'),
    ('B-104', 'Техосмотр B-104 / 2026', '2026-02-07', '2027-02-07', '/storage/docs/inspection/b-104-2026.pdf'),
    ('B-105', 'Техосмотр B-105 / 2026', '2026-02-09', '2027-02-09', '/storage/docs/inspection/b-105-2026.pdf'),
    ('B-106', 'Техосмотр B-106 / 2026', '2026-02-11', '2027-02-11', '/storage/docs/inspection/b-106-2026.pdf'),
    ('B-107', 'Техосмотр B-107 / 2026', '2026-02-13', '2027-02-13', '/storage/docs/inspection/b-107-2026.pdf'),
    ('B-108', 'Техосмотр B-108 / 2026', '2026-02-15', '2027-02-15', '/storage/docs/inspection/b-108-2026.pdf'),
    ('B-109', 'Техосмотр B-109 / 2026', '2026-02-17', '2027-02-17', '/storage/docs/inspection/b-109-2026.pdf'),
    ('B-110', 'Техосмотр B-110 / 2026', '2026-02-19', '2027-02-19', '/storage/docs/inspection/b-110-2026.pdf'),
    ('B-111', 'Техосмотр B-111 / 2026', '2026-02-21', '2027-02-21', '/storage/docs/inspection/b-111-2026.pdf'),
    ('B-112', 'Техосмотр B-112 / 2026', '2026-02-23', '2027-02-23', '/storage/docs/inspection/b-112-2026.pdf'),
    ('B-113', 'Техосмотр B-113 / 2026', '2026-02-25', '2027-02-25', '/storage/docs/inspection/b-113-2026.pdf'),
    ('B-114', 'Техосмотр B-114 / 2026', '2026-02-27', '2027-02-27', '/storage/docs/inspection/b-114-2026.pdf'),
    ('B-115', 'Техосмотр B-115 / 2026', '2026-03-01', '2027-03-01', '/storage/docs/inspection/b-115-2026.pdf'),
    ('B-116', 'Техосмотр B-116 / 2026', '2026-03-03', '2027-03-03', '/storage/docs/inspection/b-116-2026.pdf'),
    ('B-117', 'Техосмотр B-117 / 2026', '2026-03-05', '2027-03-05', '/storage/docs/inspection/b-117-2026.pdf'),
    ('B-118', 'Техосмотр B-118 / 2026', '2026-03-07', '2027-03-07', '/storage/docs/inspection/b-118-2026.pdf'),
    ('B-119', 'Техосмотр B-119 / 2026', '2026-03-09', '2027-03-09', '/storage/docs/inspection/b-119-2026.pdf'),
    ('B-120', 'Техосмотр B-120 / 2026', '2026-03-11', '2027-03-11', '/storage/docs/inspection/b-120-2026.pdf')
) AS v(board_number, name, start_date, end_date, file_path)
JOIN vehicles veh ON veh.board_number = v.board_number
WHERE NOT EXISTS (
    SELECT 1 FROM technical_inspection ti WHERE ti.vehicle_id = veh.id AND ti.name = v.name
);

INSERT INTO vehicle_documents (vehicle_id, type, number, valid_from, valid_to)
SELECT veh.id, v.type, v.number, v.valid_from::date, v.valid_to::date
FROM (VALUES
    ('B-101', 'insurance', 'DOC-INS-B101-2026', '2026-01-01', '2026-12-31'),
    ('B-102', 'insurance', 'DOC-INS-B102-2026', '2025-11-20', '2026-11-20'),
    ('B-103', 'insurance', 'DOC-INS-B103-2026', '2025-10-14', '2026-10-14'),
    ('B-104', 'insurance', 'DOC-INS-B104-2027', '2026-01-18', '2027-01-18'),
    ('B-105', 'insurance', 'DOC-INS-B105-2026', '2025-09-30', '2026-09-30'),
    ('B-106', 'insurance', 'DOC-INS-B106-2026', '2025-08-22', '2026-08-22'),
    ('B-107', 'insurance', 'DOC-INS-B107-2026', '2025-12-05', '2026-12-05'),
    ('B-108', 'insurance', 'DOC-INS-B108-2026', '2025-07-19', '2026-07-19'),
    ('B-109', 'insurance', 'DOC-INS-B109-2027', '2026-02-10', '2027-02-10'),
    ('B-110', 'insurance', 'DOC-INS-B110-2026', '2025-10-28', '2026-10-28'),
    ('B-111', 'tachograph', 'DOC-TCH-B111-2026', '2026-01-15', '2027-01-15'),
    ('B-112', 'tachograph', 'DOC-TCH-B112-2026', '2026-01-16', '2027-01-16'),
    ('B-113', 'tachograph', 'DOC-TCH-B113-2026', '2026-01-17', '2027-01-17'),
    ('B-114', 'tachograph', 'DOC-TCH-B114-2026', '2026-01-18', '2027-01-18'),
    ('B-115', 'tachograph', 'DOC-TCH-B115-2026', '2026-01-19', '2027-01-19'),
    ('B-116', 'tachograph', 'DOC-TCH-B116-2026', '2026-01-20', '2027-01-20'),
    ('B-117', 'tachograph', 'DOC-TCH-B117-2026', '2026-01-21', '2027-01-21'),
    ('B-118', 'tachograph', 'DOC-TCH-B118-2026', '2026-01-22', '2027-01-22'),
    ('B-119', 'tachograph', 'DOC-TCH-B119-2026', '2026-01-23', '2027-01-23'),
    ('B-120', 'tachograph', 'DOC-TCH-B120-2026', '2026-01-24', '2027-01-24')
) AS v(board_number, type, number, valid_from, valid_to)
JOIN vehicles veh ON veh.board_number = v.board_number
WHERE NOT EXISTS (
    SELECT 1 FROM vehicle_documents vd WHERE vd.vehicle_id = veh.id AND vd.number = v.number
);

INSERT INTO fuel_norms (
    vehicle_id, norm_per_100km, summer_norm, winter_norm, cold_air_norm, warm_air_norm, deviation
)
SELECT veh.id, v.norm_per_100km, v.summer_norm, v.winter_norm, v.cold_air_norm, v.warm_air_norm, v.deviation
FROM (VALUES
    ('B-101', 11.80, 11.20, 12.60, 0.80, 0.40, 5.00),
    ('B-102', 12.90, 12.30, 13.80, 0.90, 0.50, 5.00),
    ('B-103', 14.20, 13.60, 15.20, 1.00, 0.60, 6.00),
    ('B-104', 10.80, 10.20, 11.60, 0.80, 0.40, 5.00),
    ('B-105', 11.40, 10.80, 12.20, 0.80, 0.40, 5.00),
    ('B-106', 18.50, 17.80, 19.90, 1.40, 0.70, 7.00),
    ('B-107', 17.80, 17.10, 19.10, 1.30, 0.70, 7.00),
    ('B-108', 10.20, 9.70, 11.00, 0.80, 0.30, 5.00),
    ('B-109', 12.10, 11.50, 13.00, 0.90, 0.40, 5.00),
    ('B-110', 9.80, 9.30, 10.60, 0.80, 0.30, 5.00),
    ('B-111', 12.70, 12.00, 13.70, 1.00, 0.50, 6.00),
    ('B-112', 11.90, 11.30, 12.80, 0.90, 0.40, 5.00),
    ('B-113', 19.80, 19.00, 21.20, 1.40, 0.80, 7.00),
    ('B-114', 24.50, 23.40, 26.20, 1.70, 0.90, 8.00),
    ('B-115', 8.70, 8.20, 9.30, 0.60, 0.30, 5.00),
    ('B-116', 7.60, 7.20, 8.20, 0.60, 0.20, 5.00),
    ('B-117', 13.50, 12.80, 14.60, 1.10, 0.50, 6.00),
    ('B-118', 9.90, 9.40, 10.70, 0.80, 0.30, 5.00),
    ('B-119', 7.40, 7.00, 8.00, 0.60, 0.20, 5.00),
    ('B-120', 8.90, 8.40, 9.60, 0.70, 0.30, 5.00)
) AS v(board_number, norm_per_100km, summer_norm, winter_norm, cold_air_norm, warm_air_norm, deviation)
JOIN vehicles veh ON veh.board_number = v.board_number
ON CONFLICT (vehicle_id) DO UPDATE SET
    norm_per_100km = EXCLUDED.norm_per_100km,
    summer_norm = EXCLUDED.summer_norm,
    winter_norm = EXCLUDED.winter_norm,
    cold_air_norm = EXCLUDED.cold_air_norm,
    warm_air_norm = EXCLUDED.warm_air_norm,
    deviation = EXCLUDED.deviation,
    updated_at = NOW();

INSERT INTO tires (place_id, vehicle_id, tire, mileage, max_usage, installed_at)
SELECT tp.id, veh.id, v.tire, v.mileage, v.max_usage, v.installed_at::date
FROM (VALUES
    ('B-101', 'передняя левая', 'Michelin Agilis 215/65 R16 DEMO-001', 18200, 65000, '2025-10-15'),
    ('B-101', 'передняя правая', 'Michelin Agilis 215/65 R16 DEMO-002', 18200, 65000, '2025-10-15'),
    ('B-101', 'задняя левая', 'Michelin Agilis 215/65 R16 DEMO-003', 17600, 65000, '2025-10-15'),
    ('B-101', 'задняя правая', 'Michelin Agilis 215/65 R16 DEMO-004', 17600, 65000, '2025-10-15'),
    ('B-102', 'передняя левая', 'Continental VanContact 235/65 R16 DEMO-005', 24500, 70000, '2025-09-20'),
    ('B-102', 'передняя правая', 'Continental VanContact 235/65 R16 DEMO-006', 24500, 70000, '2025-09-20'),
    ('B-102', 'задняя левая', 'Continental VanContact 235/65 R16 DEMO-007', 23800, 70000, '2025-09-20'),
    ('B-102', 'задняя правая', 'Continental VanContact 235/65 R16 DEMO-008', 23800, 70000, '2025-09-20'),
    ('B-106', 'передняя левая', 'Bridgestone Duravis 215/75 R17.5 DEMO-009', 31200, 85000, '2025-08-11'),
    ('B-106', 'передняя правая', 'Bridgestone Duravis 215/75 R17.5 DEMO-010', 31200, 85000, '2025-08-11'),
    ('B-106', 'задняя левая', 'Bridgestone Duravis 215/75 R17.5 DEMO-011', 30500, 85000, '2025-08-11'),
    ('B-106', 'задняя правая', 'Bridgestone Duravis 215/75 R17.5 DEMO-012', 30500, 85000, '2025-08-11'),
    ('B-109', 'передняя левая', 'Nokian Hakka SUV 265/60 R18 DEMO-013', 12100, 60000, '2025-11-02'),
    ('B-109', 'передняя правая', 'Nokian Hakka SUV 265/60 R18 DEMO-014', 12100, 60000, '2025-11-02'),
    ('B-109', 'задняя левая', 'Nokian Hakka SUV 265/60 R18 DEMO-015', 11800, 60000, '2025-11-02'),
    ('B-109', 'задняя правая', 'Nokian Hakka SUV 265/60 R18 DEMO-016', 11800, 60000, '2025-11-02'),
    ('B-120', 'передняя левая', 'Pirelli Cinturato 235/45 R18 DEMO-017', 8400, 55000, '2026-01-12'),
    ('B-120', 'передняя правая', 'Pirelli Cinturato 235/45 R18 DEMO-018', 8400, 55000, '2026-01-12'),
    ('B-120', 'задняя левая', 'Pirelli Cinturato 235/45 R18 DEMO-019', 8200, 55000, '2026-01-12'),
    ('B-120', 'задняя правая', 'Pirelli Cinturato 235/45 R18 DEMO-020', 8200, 55000, '2026-01-12')
) AS v(board_number, place_name, tire, mileage, max_usage, installed_at)
JOIN vehicles veh ON veh.board_number = v.board_number
JOIN tire_places tp ON tp.name = v.place_name
WHERE NOT EXISTS (SELECT 1 FROM tires t WHERE t.tire = v.tire);

INSERT INTO tripsheets (
    tripsheet_number, tripsheet_date, vehicle_id, vehicle_brand, vehicle_plate_number,
    driver_last_name, driver_first_name, driver_middle_name, driver_id, driver_shift_id,
    start_time, end_time, mileage_start, mileage_end, fuel_start, fuel_issued,
    fuel_consumption_theoretical, fuel_consumption_actual, status_id
)
SELECT
    v.number,
    v.trip_date::date,
    veh.id,
    veh.brand_model,
    veh.state_number,
    d.surname,
    d.name,
    d.middlename,
    d.id,
    ds.id,
    (v.trip_date::date + v.start_time::time),
    CASE WHEN v.end_time IS NULL THEN NULL ELSE (v.trip_date::date + v.end_time::time) END,
    v.mileage_start,
    v.mileage_end,
    v.fuel_start,
    v.fuel_issued,
    v.fuel_theoretical,
    v.fuel_actual,
    ts.id
FROM (VALUES
    ('WL-2026-0001', '2026-05-06', 'B-101', '850101300001', '08:15', '17:10', 83820, 84062, 48, 30, 28, 29, 'Окончен'),
    ('WL-2026-0002', '2026-05-07', 'B-102', '870214300002', '07:45', '17:55', 115910, 116188, 62, 42, 36, 38, 'Окончен'),
    ('WL-2026-0003', '2026-05-08', 'B-103', '900305300003', '08:10', '16:50', 148120, 148305, 40, 25, 26, 27, 'Окончен'),
    ('WL-2026-0004', '2026-05-09', 'B-104', '820416300004', '09:10', '18:20', 63580, 63748, 65, 20, 18, 19, 'Окончен'),
    ('WL-2026-0005', '2026-05-10', 'B-105', '760527300005', '08:20', '16:25', 95310, 95430, 51, 18, 14, 15, 'Окончен'),
    ('WL-2026-0006', '2026-05-11', 'B-106', '930608300006', '08:00', '18:05', 201870, 202130, 94, 60, 48, 50, 'Окончен'),
    ('WL-2026-0007', '2026-05-12', 'B-107', '880719300007', '08:30', '17:30', 132300, 132498, 83, 36, 35, 36, 'Окончен'),
    ('WL-2026-0008', '2026-05-13', 'B-108', '810830300008', '08:10', '18:45', 119770, 119998, 35, 30, 23, 24, 'Окончен'),
    ('WL-2026-0009', '2026-05-14', 'B-109', '951011300009', '09:10', '17:55', 51040, 51204, 70, 22, 20, 21, 'Окончен'),
    ('WL-2026-0010', '2026-05-15', 'B-110', '790922300010', '07:15', '16:45', 87310, 87522, 47, 25, 21, 22, 'Окончен'),
    ('WL-2026-0011', '2026-05-16', 'B-111', '910103300011', '08:45', '18:10', 139320, 139556, 60, 34, 30, 31, 'Окончен'),
    ('WL-2026-0012', '2026-05-17', 'B-112', '841224300012', '08:05', '17:05', 153820, 154025, 54, 31, 24, 25, 'Окончен'),
    ('WL-2026-0013', '2026-05-18', 'B-113', '960415300013', '09:05', '17:50', 220910, 221102, 110, 44, 38, 39, 'Окончен'),
    ('WL-2026-0014', '2026-05-19', 'B-114', '830706300014', '07:45', '17:15', 264540, 264705, 128, 40, 41, 43, 'Окончен'),
    ('WL-2026-0015', '2026-05-20', 'B-115', '890217300015', '08:15', '16:10', 78910, 79025, 28, 15, 10, 11, 'Окончен'),
    ('WL-2026-0016', '2026-05-21', 'B-116', '920328300016', '08:10', '17:55', 45480, 45658, 26, 18, 14, 14, 'Окончен'),
    ('WL-2026-0017', '2026-05-22', 'B-117', '780509300017', '07:40', '17:35', 110120, 110330, 44, 35, 28, 29, 'Окончен'),
    ('WL-2026-0018', '2026-05-23', 'B-118', '940620300018', '09:05', '18:05', 21740, 21872, 62, 16, 13, 13, 'Окончен'),
    ('WL-2026-0019', '2026-05-24', 'B-119', '800731300019', '08:20', '18:50', 68120, 68268, 23, 15, 11, 12, 'Окончен'),
    ('WL-2026-0020', '2026-05-25', 'B-120', '971112300020', '07:45', NULL, 39120, 39190, 40, 12, 6, 6, 'В процессе')
) AS v(number, trip_date, board_number, driver_iin, start_time, end_time, mileage_start, mileage_end, fuel_start, fuel_issued, fuel_theoretical, fuel_actual, status_name)
JOIN vehicles veh ON veh.board_number = v.board_number
JOIN drivers d ON d.iin = v.driver_iin
LEFT JOIN LATERAL (
    SELECT id
    FROM driver_shifts ds
    WHERE ds.driver_id = d.id
      AND ds.shift_date = v.trip_date::date
    ORDER BY ds.id DESC
    LIMIT 1
) ds ON TRUE
JOIN tripsheet_statuses ts ON ts.name = v.status_name
ON CONFLICT (tripsheet_number, tripsheet_date) DO UPDATE SET
    vehicle_id = EXCLUDED.vehicle_id,
    vehicle_brand = EXCLUDED.vehicle_brand,
    vehicle_plate_number = EXCLUDED.vehicle_plate_number,
    driver_last_name = EXCLUDED.driver_last_name,
    driver_first_name = EXCLUDED.driver_first_name,
    driver_middle_name = EXCLUDED.driver_middle_name,
    driver_id = EXCLUDED.driver_id,
    driver_shift_id = EXCLUDED.driver_shift_id,
    start_time = EXCLUDED.start_time,
    end_time = EXCLUDED.end_time,
    mileage_start = EXCLUDED.mileage_start,
    mileage_end = EXCLUDED.mileage_end,
    fuel_start = EXCLUDED.fuel_start,
    fuel_issued = EXCLUDED.fuel_issued,
    fuel_consumption_theoretical = EXCLUDED.fuel_consumption_theoretical,
    fuel_consumption_actual = EXCLUDED.fuel_consumption_actual,
    status_id = EXCLUDED.status_id,
    updated_at = NOW();

INSERT INTO tripsheet_trips (
    tripsheet_id, route_description, start_time, end_time, distance_passed, status_id
)
SELECT
    t.id,
    v.route_description,
    (t.tripsheet_date + v.start_time::time),
    CASE WHEN v.end_time IS NULL THEN NULL ELSE (t.tripsheet_date + v.end_time::time) END,
    v.distance_passed,
    ts.id
FROM (VALUES
    ('WL-2026-0001', 'Парк - склад Алтын Орда - центральный офис - парк', '08:30', '16:55', 242, 'Окончен'),
    ('WL-2026-0002', 'Парк - БЦ Нурлы Тау - медцентр - склад запчастей - парк', '08:00', '17:40', 278, 'Окончен'),
    ('WL-2026-0003', 'Парк - сервисная зона Майлина - филиал Ауэзов - парк', '08:25', '16:35', 185, 'Окончен'),
    ('WL-2026-0004', 'Парк - логистический терминал DAMU - парк', '09:25', '18:00', 168, 'Окончен'),
    ('WL-2026-0005', 'Парк - районная база - АЗС - парк', '08:35', '16:05', 120, 'Окончен'),
    ('WL-2026-0006', 'Парк - промзона Алатау - склад N2 - парк', '08:15', '17:50', 260, 'Окончен'),
    ('WL-2026-0007', 'Парк - пункт выдачи Жетысу - склад шин - парк', '08:45', '17:20', 198, 'Окончен'),
    ('WL-2026-0008', 'Парк - промышленный парк - центральный склад - парк', '08:25', '18:25', 228, 'Окончен'),
    ('WL-2026-0009', 'Парк - акимат района - архив - парк', '09:20', '17:35', 164, 'Окончен'),
    ('WL-2026-0010', 'Парк - пригородный объект - склад - парк', '07:30', '16:25', 212, 'Окончен'),
    ('WL-2026-0011', 'Парк - технопарк - филиал N3 - парк', '09:00', '17:55', 236, 'Окончен'),
    ('WL-2026-0012', 'Парк - городская больница - поликлиника - парк', '08:20', '16:50', 205, 'Окончен'),
    ('WL-2026-0013', 'Парк - склад крупногабаритных материалов - парк', '09:20', '17:35', 192, 'Окончен'),
    ('WL-2026-0014', 'Парк - ремонтная база - мойка - парк', '08:00', '16:55', 165, 'Окончен'),
    ('WL-2026-0015', 'Парк - сервисная база - центральный офис - парк', '08:30', '15:55', 115, 'Окончен'),
    ('WL-2026-0016', 'Парк - склад документов - клиентский центр - парк', '08:25', '17:40', 178, 'Окончен'),
    ('WL-2026-0017', 'Парк - производственная площадка Восток - парк', '07:55', '17:20', 210, 'Окончен'),
    ('WL-2026-0018', 'Парк - офис HR - филиал логистики - парк', '09:20', '17:50', 132, 'Окончен'),
    ('WL-2026-0019', 'Парк - склад N4 - точка выдачи - парк', '08:35', '18:30', 148, 'Окончен'),
    ('WL-2026-0020', 'Парк - административный центр - склад - маршрут продолжается', '08:00', NULL, 70, 'В процессе')
) AS v(tripsheet_number, route_description, start_time, end_time, distance_passed, status_name)
JOIN tripsheets t ON t.tripsheet_number = v.tripsheet_number
JOIN tripsheet_statuses ts ON ts.name = v.status_name
WHERE NOT EXISTS (
    SELECT 1 FROM tripsheet_trips tt
    WHERE tt.tripsheet_id = t.id AND tt.route_description = v.route_description
);

INSERT INTO waybill_route_points (
    waybill_id, seq_number, destination, arrival_time, hospitalization_time, lpu_arrival_time, release_time
)
SELECT t.id, v.seq_number, v.destination, v.arrival_time::time, v.hospitalization_time::time, v.lpu_arrival_time::time, v.release_time::time
FROM (VALUES
    ('WL-2026-0001', 1, 'Склад Алтын Орда, ворота 4', '09:10', NULL, NULL, '09:40'),
    ('WL-2026-0002', 1, 'БЦ Нурлы Тау, блок 3Б', '08:45', NULL, NULL, '09:20'),
    ('WL-2026-0003', 1, 'Сервисная зона Майлина, пост 2', '09:05', NULL, NULL, '09:50'),
    ('WL-2026-0004', 1, 'Логистический терминал DAMU', '10:20', NULL, NULL, '11:10'),
    ('WL-2026-0005', 1, 'Районная база, склад ГСМ', '09:00', NULL, NULL, '09:35'),
    ('WL-2026-0006', 1, 'Промзона Алатау, корпус 7', '09:15', NULL, NULL, '10:15'),
    ('WL-2026-0007', 1, 'Пункт выдачи Жетысу', '09:30', NULL, NULL, '10:05'),
    ('WL-2026-0008', 1, 'Промышленный парк, КПП 1', '09:25', NULL, NULL, '10:30'),
    ('WL-2026-0009', 1, 'Акимат района, канцелярия', '10:00', NULL, NULL, '10:25'),
    ('WL-2026-0010', 1, 'Пригородный объект, склад N2', '09:15', NULL, NULL, '10:20'),
    ('WL-2026-0011', 1, 'Технопарк, блок С', '09:45', NULL, NULL, '10:30'),
    ('WL-2026-0012', 1, 'Городская больница, приемный блок', '09:05', NULL, '09:15', '09:55'),
    ('WL-2026-0013', 1, 'Склад крупногабаритных материалов', '10:10', NULL, NULL, '11:00'),
    ('WL-2026-0014', 1, 'Ремонтная база, зона диагностики', '08:45', NULL, NULL, '09:35'),
    ('WL-2026-0015', 1, 'Сервисная база, отдел эксплуатации', '09:00', NULL, NULL, '09:30'),
    ('WL-2026-0016', 1, 'Склад документов, архив', '09:10', NULL, NULL, '09:50'),
    ('WL-2026-0017', 1, 'Производственная площадка Восток', '09:05', NULL, NULL, '10:15'),
    ('WL-2026-0018', 1, 'Офис HR, проспект Абая', '10:00', NULL, NULL, '10:25'),
    ('WL-2026-0019', 1, 'Склад N4, зона отгрузки', '09:20', NULL, NULL, '10:05'),
    ('WL-2026-0020', 1, 'Административный центр, главный вход', '08:40', NULL, NULL, NULL)
) AS v(tripsheet_number, seq_number, destination, arrival_time, hospitalization_time, lpu_arrival_time, release_time)
JOIN tripsheets t ON t.tripsheet_number = v.tripsheet_number
WHERE NOT EXISTS (
    SELECT 1 FROM waybill_route_points rp
    WHERE rp.waybill_id = t.id AND rp.seq_number = v.seq_number AND rp.destination = v.destination
);

INSERT INTO fuel_refills (tripsheet_id, vehicle_id, fuel_amount, date, time, location, price)
SELECT t.id, veh.id, v.fuel_amount, v.date::date, v.time::time, v.location, v.price
FROM (VALUES
    ('WL-2026-0001', 'B-101', 30.00, '2026-05-06', '12:20', 'АЗС Sinooil, ул. Толе би', 7650.00),
    ('WL-2026-0002', 'B-102', 42.00, '2026-05-07', '13:10', 'АЗС Helios, пр. Сейфуллина', 10710.00),
    ('WL-2026-0003', 'B-103', 25.00, '2026-05-08', '11:45', 'АЗС Qazaq Oil, ул. Майлина', 6375.00),
    ('WL-2026-0004', 'B-104', 20.00, '2026-05-09', '14:05', 'АЗС Sinooil, трасса А3', 5100.00),
    ('WL-2026-0005', 'B-105', 18.00, '2026-05-10', '10:30', 'АЗС Helios, Райымбека', 4590.00),
    ('WL-2026-0006', 'B-106', 60.00, '2026-05-11', '15:15', 'АЗС Compass, промзона', 15300.00),
    ('WL-2026-0007', 'B-107', 36.00, '2026-05-12', '12:50', 'АЗС Qazaq Oil, Жетысу', 9180.00),
    ('WL-2026-0008', 'B-108', 30.00, '2026-05-13', '13:35', 'АЗС Helios, Северное кольцо', 7650.00),
    ('WL-2026-0009', 'B-109', 22.00, '2026-05-14', '11:20', 'АЗС Sinooil, Аль-Фараби', 5610.00),
    ('WL-2026-0010', 'B-110', 25.00, '2026-05-15', '09:55', 'АЗС Royal Petrol, Каскелен', 6375.00),
    ('WL-2026-0011', 'B-111', 34.00, '2026-05-16', '14:30', 'АЗС Helios, Ташкентская', 8670.00),
    ('WL-2026-0012', 'B-112', 31.00, '2026-05-17', '12:05', 'АЗС Qazaq Oil, Абая', 7905.00),
    ('WL-2026-0013', 'B-113', 44.00, '2026-05-18', '13:40', 'АЗС Compass, складская зона', 11220.00),
    ('WL-2026-0014', 'B-114', 40.00, '2026-05-19', '10:50', 'АЗС Helios, Восточная объездная', 10200.00),
    ('WL-2026-0015', 'B-115', 15.00, '2026-05-20', '11:10', 'АЗС Sinooil, Сатпаева', 3825.00),
    ('WL-2026-0016', 'B-116', 18.00, '2026-05-21', '13:00', 'АЗС Qazaq Oil, Достык', 4590.00),
    ('WL-2026-0017', 'B-117', 35.00, '2026-05-22', '12:30', 'АЗС Helios, Восток', 8925.00),
    ('WL-2026-0018', 'B-118', 16.00, '2026-05-23', '15:05', 'АЗС Sinooil, Назарбаева', 4080.00),
    ('WL-2026-0019', 'B-119', 15.00, '2026-05-24', '16:20', 'АЗС Royal Petrol, Жибек Жолы', 3825.00),
    ('WL-2026-0020', 'B-120', 12.00, '2026-05-25', '10:25', 'АЗС Qazaq Oil, Аль-Фараби', 3060.00)
) AS v(tripsheet_number, board_number, fuel_amount, date, time, location, price)
JOIN tripsheets t ON t.tripsheet_number = v.tripsheet_number
JOIN vehicles veh ON veh.board_number = v.board_number
WHERE NOT EXISTS (
    SELECT 1 FROM fuel_refills fr
    WHERE fr.tripsheet_id = t.id AND fr.date = v.date::date AND fr.time = v.time::time
);

INSERT INTO parts_catalog (
    part_id, name, quantity, category, dimensions, manufacturer,
    is_consumable, price, min_stock_quantity, unit
)
SELECT *
FROM (VALUES
    ('DEMO-BRK-PAD-FR', 'Колодки тормозные передние Toyota/Hyundai', 34, 'Тормозная система', 'комплект 4 шт', 'Brembo', FALSE, 28500.00, 8, 'к-т'),
    ('DEMO-BRK-DISC-FR', 'Диски тормозные передние универсальные', 18, 'Тормозная система', 'D=300 мм', 'TRW', FALSE, 42000.00, 6, 'к-т'),
    ('DEMO-OIL-5W30', 'Моторное масло 5W-30 дизель', 120, 'Масла и жидкости', 'канистра 4 л', 'Shell', TRUE, 9800.00, 24, 'шт'),
    ('DEMO-FILTER-OIL', 'Фильтр масляный коммерческий транспорт', 56, 'Фильтры', 'M20x1.5', 'MANN-FILTER', TRUE, 4200.00, 16, 'шт'),
    ('DEMO-FILTER-AIR', 'Фильтр воздушный фургон/пикап', 42, 'Фильтры', 'панельный', 'Bosch', TRUE, 6800.00, 12, 'шт'),
    ('DEMO-FILTER-FUEL', 'Фильтр топливный дизель', 37, 'Фильтры', 'тонкая очистка', 'Mahle', TRUE, 7600.00, 10, 'шт'),
    ('DEMO-WIPER-650', 'Щетка стеклоочистителя 650 мм', 45, 'Кузов и стекла', '650 мм', 'Denso', TRUE, 3900.00, 12, 'шт'),
    ('DEMO-BAT-90AH', 'Аккумулятор 90 Ач обратная полярность', 12, 'Электрика', '90 Ач', 'Varta', FALSE, 64500.00, 4, 'шт'),
    ('DEMO-LAMP-H7', 'Лампа H7 12V 55W', 80, 'Электрика', 'H7', 'Philips', TRUE, 2500.00, 20, 'шт'),
    ('DEMO-TIRE-215-65R16', 'Шина летняя 215/65 R16C', 28, 'Шины', '215/65 R16C', 'Michelin', FALSE, 54500.00, 8, 'шт'),
    ('DEMO-TIRE-235-65R16', 'Шина летняя 235/65 R16C', 24, 'Шины', '235/65 R16C', 'Continental', FALSE, 57500.00, 8, 'шт'),
    ('DEMO-COOLANT-G12', 'Антифриз G12 красный', 65, 'Масла и жидкости', 'канистра 5 л', 'Felix', TRUE, 6200.00, 15, 'шт'),
    ('DEMO-BELT-GEN', 'Ремень генератора коммерческий транспорт', 21, 'Двигатель', '6PK', 'Gates', FALSE, 11800.00, 6, 'шт'),
    ('DEMO-PUMP-WATER', 'Помпа системы охлаждения', 9, 'Двигатель', 'алюминий', 'GMB', FALSE, 36800.00, 3, 'шт'),
    ('DEMO-SHOCK-FR', 'Амортизатор передний усиленный', 16, 'Подвеска', 'стойка', 'KYB', FALSE, 33500.00, 4, 'шт'),
    ('DEMO-SHOCK-RR', 'Амортизатор задний усиленный', 18, 'Подвеска', 'стойка', 'Sachs', FALSE, 29800.00, 4, 'шт'),
    ('DEMO-PADS-DRUM', 'Колодки барабанные задние', 14, 'Тормозная система', 'комплект', 'Ferodo', FALSE, 24500.00, 4, 'к-т'),
    ('DEMO-ADBLUE-10L', 'AdBlue раствор мочевины', 90, 'Масла и жидкости', 'канистра 10 л', 'TotalEnergies', TRUE, 5200.00, 20, 'шт'),
    ('DEMO-FUSE-KIT', 'Набор автомобильных предохранителей', 40, 'Электрика', 'мини/стандарт', 'Bosch', TRUE, 1800.00, 12, 'к-т'),
    ('DEMO-MIRROR-L', 'Зеркало боковое левое с подогревом', 7, 'Кузов и стекла', 'левое', 'TYC', FALSE, 46500.00, 2, 'шт')
) AS v(part_id, name, quantity, category, dimensions, manufacturer, is_consumable, price, min_stock_quantity, unit)
ON CONFLICT (part_id) DO UPDATE SET
    name = EXCLUDED.name,
    quantity = EXCLUDED.quantity,
    category = EXCLUDED.category,
    dimensions = EXCLUDED.dimensions,
    manufacturer = EXCLUDED.manufacturer,
    is_consumable = EXCLUDED.is_consumable,
    price = EXCLUDED.price,
    min_stock_quantity = EXCLUDED.min_stock_quantity,
    unit = EXCLUDED.unit,
    updated_at = NOW();

INSERT INTO part_arrivals (
    document_number, arrival_date, status, comment, created_by_user_id, accepted_by_user_id, accepted_at
)
SELECT v.document_number, v.arrival_date::date, 'accepted', v.comment, creator.id, accepter.id, v.accepted_at::timestamptz
FROM (VALUES
    ('ARR-DEMO-2026-001', '2026-05-02', 'Поставка фильтров и масел от ТОО AutoSupply', 'warehouse@autopark.demo', 'warehouse2@autopark.demo', '2026-05-02 15:30:00+05'),
    ('ARR-DEMO-2026-002', '2026-05-08', 'Поставка тормозных компонентов и шин', 'warehouse2@autopark.demo', 'warehouse@autopark.demo', '2026-05-08 16:10:00+05'),
    ('ARR-DEMO-2026-003', '2026-05-15', 'Поставка электрики и расходных материалов', 'warehouse@autopark.demo', 'warehouse2@autopark.demo', '2026-05-15 12:40:00+05'),
    ('ARR-DEMO-2026-004', '2026-05-22', 'Допоставка подвески и кузовных элементов', 'warehouse2@autopark.demo', 'warehouse@autopark.demo', '2026-05-22 17:05:00+05')
) AS v(document_number, arrival_date, comment, created_by_email, accepted_by_email, accepted_at)
JOIN users creator ON creator.email = v.created_by_email
JOIN users accepter ON accepter.email = v.accepted_by_email
ON CONFLICT (document_number) DO UPDATE SET
    arrival_date = EXCLUDED.arrival_date,
    status = EXCLUDED.status,
    comment = EXCLUDED.comment,
    accepted_by_user_id = EXCLUDED.accepted_by_user_id,
    accepted_at = EXCLUDED.accepted_at,
    updated_at = NOW();

INSERT INTO part_arrival_items (arrival_id, part_id, quantity, price)
SELECT pa.id, pc.id, v.quantity, v.price
FROM (VALUES
    ('ARR-DEMO-2026-001', 'DEMO-OIL-5W30', 40, 9600.00),
    ('ARR-DEMO-2026-001', 'DEMO-FILTER-OIL', 30, 3900.00),
    ('ARR-DEMO-2026-001', 'DEMO-FILTER-AIR', 25, 6400.00),
    ('ARR-DEMO-2026-001', 'DEMO-FILTER-FUEL', 22, 7200.00),
    ('ARR-DEMO-2026-001', 'DEMO-COOLANT-G12', 35, 5900.00),
    ('ARR-DEMO-2026-002', 'DEMO-BRK-PAD-FR', 14, 27000.00),
    ('ARR-DEMO-2026-002', 'DEMO-BRK-DISC-FR', 8, 39500.00),
    ('ARR-DEMO-2026-002', 'DEMO-PADS-DRUM', 6, 23000.00),
    ('ARR-DEMO-2026-002', 'DEMO-TIRE-215-65R16', 12, 52000.00),
    ('ARR-DEMO-2026-002', 'DEMO-TIRE-235-65R16', 10, 54800.00),
    ('ARR-DEMO-2026-003', 'DEMO-LAMP-H7', 40, 2300.00),
    ('ARR-DEMO-2026-003', 'DEMO-BAT-90AH', 6, 61000.00),
    ('ARR-DEMO-2026-003', 'DEMO-FUSE-KIT', 18, 1600.00),
    ('ARR-DEMO-2026-003', 'DEMO-WIPER-650', 20, 3600.00),
    ('ARR-DEMO-2026-003', 'DEMO-ADBLUE-10L', 45, 4900.00),
    ('ARR-DEMO-2026-004', 'DEMO-SHOCK-FR', 8, 31800.00),
    ('ARR-DEMO-2026-004', 'DEMO-SHOCK-RR', 8, 28200.00),
    ('ARR-DEMO-2026-004', 'DEMO-BELT-GEN', 10, 10900.00),
    ('ARR-DEMO-2026-004', 'DEMO-PUMP-WATER', 4, 34500.00),
    ('ARR-DEMO-2026-004', 'DEMO-MIRROR-L', 3, 43000.00)
) AS v(document_number, part_code, quantity, price)
JOIN part_arrivals pa ON pa.document_number = v.document_number
JOIN parts_catalog pc ON pc.part_id = v.part_code
WHERE NOT EXISTS (
    SELECT 1 FROM part_arrival_items item
    WHERE item.arrival_id = pa.id AND item.part_id = pc.id
);

INSERT INTO part_stock_movements (
    part_id, type, quantity, vehicle_id, part_request_id, document_number, actor_user_id, created_at
)
SELECT pc.id, 'arrival', v.quantity, NULL, NULL, v.document_number, u.id, v.created_at::timestamptz
FROM (VALUES
    ('DEMO-OIL-5W30', 40, 'ARR-DEMO-2026-001', 'warehouse@autopark.demo', '2026-05-02 15:35:00+05'),
    ('DEMO-FILTER-OIL', 30, 'ARR-DEMO-2026-001', 'warehouse@autopark.demo', '2026-05-02 15:36:00+05'),
    ('DEMO-FILTER-AIR', 25, 'ARR-DEMO-2026-001', 'warehouse@autopark.demo', '2026-05-02 15:37:00+05'),
    ('DEMO-FILTER-FUEL', 22, 'ARR-DEMO-2026-001', 'warehouse@autopark.demo', '2026-05-02 15:38:00+05'),
    ('DEMO-COOLANT-G12', 35, 'ARR-DEMO-2026-001', 'warehouse@autopark.demo', '2026-05-02 15:39:00+05'),
    ('DEMO-BRK-PAD-FR', 14, 'ARR-DEMO-2026-002', 'warehouse2@autopark.demo', '2026-05-08 16:15:00+05'),
    ('DEMO-BRK-DISC-FR', 8, 'ARR-DEMO-2026-002', 'warehouse2@autopark.demo', '2026-05-08 16:16:00+05'),
    ('DEMO-PADS-DRUM', 6, 'ARR-DEMO-2026-002', 'warehouse2@autopark.demo', '2026-05-08 16:17:00+05'),
    ('DEMO-TIRE-215-65R16', 12, 'ARR-DEMO-2026-002', 'warehouse2@autopark.demo', '2026-05-08 16:18:00+05'),
    ('DEMO-TIRE-235-65R16', 10, 'ARR-DEMO-2026-002', 'warehouse2@autopark.demo', '2026-05-08 16:19:00+05'),
    ('DEMO-LAMP-H7', 40, 'ARR-DEMO-2026-003', 'warehouse@autopark.demo', '2026-05-15 12:45:00+05'),
    ('DEMO-BAT-90AH', 6, 'ARR-DEMO-2026-003', 'warehouse@autopark.demo', '2026-05-15 12:46:00+05'),
    ('DEMO-FUSE-KIT', 18, 'ARR-DEMO-2026-003', 'warehouse@autopark.demo', '2026-05-15 12:47:00+05'),
    ('DEMO-WIPER-650', 20, 'ARR-DEMO-2026-003', 'warehouse@autopark.demo', '2026-05-15 12:48:00+05'),
    ('DEMO-ADBLUE-10L', 45, 'ARR-DEMO-2026-003', 'warehouse@autopark.demo', '2026-05-15 12:49:00+05'),
    ('DEMO-SHOCK-FR', 8, 'ARR-DEMO-2026-004', 'warehouse2@autopark.demo', '2026-05-22 17:10:00+05'),
    ('DEMO-SHOCK-RR', 8, 'ARR-DEMO-2026-004', 'warehouse2@autopark.demo', '2026-05-22 17:11:00+05'),
    ('DEMO-BELT-GEN', 10, 'ARR-DEMO-2026-004', 'warehouse2@autopark.demo', '2026-05-22 17:12:00+05'),
    ('DEMO-PUMP-WATER', 4, 'ARR-DEMO-2026-004', 'warehouse2@autopark.demo', '2026-05-22 17:13:00+05'),
    ('DEMO-MIRROR-L', 3, 'ARR-DEMO-2026-004', 'warehouse2@autopark.demo', '2026-05-22 17:14:00+05')
) AS v(part_code, quantity, document_number, actor_email, created_at)
JOIN parts_catalog pc ON pc.part_id = v.part_code
JOIN users u ON u.email = v.actor_email
WHERE NOT EXISTS (
    SELECT 1 FROM part_stock_movements psm
    WHERE psm.part_id = pc.id AND psm.document_number = v.document_number AND psm.type = 'arrival'
);

INSERT INTO part_requests (
    part_id, quantity, mechanic_comment, status_id, author_user_id, rejection_comment, created_at, updated_at
)
SELECT
    pc.id,
    v.quantity,
    v.comment,
    prs.id,
    u.id,
    v.rejection_comment,
    v.created_at::timestamptz,
    v.created_at::timestamptz
FROM (VALUES
    ('DEMO-BRK-PAD-FR', 1, 'Плановая замена передних тормозных колодок B-101', 'approved', 'mechanic1@autopark.demo', NULL, '2026-05-06 10:20:00+05'),
    ('DEMO-FILTER-OIL', 2, 'ТО B-102: замена масляного фильтра', 'issued', 'mechanic2@autopark.demo', NULL, '2026-05-07 11:15:00+05'),
    ('DEMO-OIL-5W30', 3, 'ТО B-102: моторное масло по регламенту', 'issued', 'mechanic2@autopark.demo', NULL, '2026-05-07 11:17:00+05'),
    ('DEMO-FILTER-AIR', 1, 'B-103: загрязнен воздушный фильтр после пыльного рейса', 'approved', 'mechanic3@autopark.demo', NULL, '2026-05-08 12:05:00+05'),
    ('DEMO-LAMP-H7', 2, 'B-104: перегорели лампы ближнего света', 'issued', 'mechanic4@autopark.demo', NULL, '2026-05-09 15:35:00+05'),
    ('DEMO-BAT-90AH', 1, 'B-105: низкий пусковой ток аккумулятора', 'new', 'mechanic1@autopark.demo', NULL, '2026-05-10 09:55:00+05'),
    ('DEMO-COOLANT-G12', 2, 'B-106: долив антифриза после проверки системы', 'approved', 'mechanic2@autopark.demo', NULL, '2026-05-11 16:25:00+05'),
    ('DEMO-WIPER-650', 2, 'B-107: износ щеток стеклоочистителя', 'issued', 'mechanic3@autopark.demo', NULL, '2026-05-12 10:40:00+05'),
    ('DEMO-TIRE-215-65R16', 2, 'B-108: замена двух шин из-за остатка протектора', 'approved', 'mechanic4@autopark.demo', NULL, '2026-05-13 14:10:00+05'),
    ('DEMO-FILTER-FUEL', 1, 'B-109: плановая замена топливного фильтра', 'issued', 'mechanic1@autopark.demo', NULL, '2026-05-14 11:45:00+05'),
    ('DEMO-BELT-GEN', 1, 'B-110: трещины на ремне генератора', 'approved', 'mechanic2@autopark.demo', NULL, '2026-05-15 12:30:00+05'),
    ('DEMO-SHOCK-FR', 2, 'B-111: стук передней подвески на неровностях', 'new', 'mechanic3@autopark.demo', NULL, '2026-05-16 15:20:00+05'),
    ('DEMO-ADBLUE-10L', 4, 'B-113: пополнение AdBlue перед дальним рейсом', 'issued', 'mechanic4@autopark.demo', NULL, '2026-05-18 10:05:00+05'),
    ('DEMO-PUMP-WATER', 1, 'B-114: перегрев под нагрузкой, требуется помпа', 'approved', 'mechanic1@autopark.demo', NULL, '2026-05-19 16:10:00+05'),
    ('DEMO-MIRROR-L', 1, 'B-115: повреждено левое зеркало после парковки', 'rejected', 'mechanic2@autopark.demo', 'Нужен акт осмотра и фотофиксация повреждения', '2026-05-20 09:35:00+05'),
    ('DEMO-FUSE-KIT', 1, 'B-116: пополнить аварийный набор предохранителей', 'issued', 'mechanic3@autopark.demo', NULL, '2026-05-21 13:50:00+05'),
    ('DEMO-SHOCK-RR', 2, 'B-117: требуется замена задних амортизаторов', 'approved', 'mechanic4@autopark.demo', NULL, '2026-05-22 11:25:00+05'),
    ('DEMO-BRK-DISC-FR', 1, 'B-118: биение руля при торможении', 'new', 'mechanic1@autopark.demo', NULL, '2026-05-23 12:15:00+05'),
    ('DEMO-PADS-DRUM', 1, 'B-119: профилактика задних барабанных тормозов', 'approved', 'mechanic2@autopark.demo', NULL, '2026-05-24 15:45:00+05'),
    ('DEMO-TIRE-235-65R16', 2, 'B-120: подготовка комплекта шин к летнему сезону', 'new', 'mechanic3@autopark.demo', NULL, '2026-05-25 09:30:00+05')
) AS v(part_code, quantity, comment, status_code, author_email, rejection_comment, created_at)
JOIN parts_catalog pc ON pc.part_id = v.part_code
JOIN part_request_statuses prs ON prs.code = v.status_code
JOIN users u ON u.email = v.author_email
WHERE NOT EXISTS (
    SELECT 1 FROM part_requests pr
    WHERE pr.part_id = pc.id AND pr.author_user_id = u.id AND pr.mechanic_comment = v.comment
);

INSERT INTO part_request_history (part_request_id, status_id, changed_by_user_id, comment, changed_at)
SELECT pr.id, pr.status_id, pr.author_user_id, 'Статус заявки зафиксирован при демонстрационном наполнении', pr.created_at
FROM part_requests pr
JOIN parts_catalog pc ON pc.id = pr.part_id
WHERE pc.part_id LIKE 'DEMO-%'
  AND NOT EXISTS (
      SELECT 1 FROM part_request_history h
      WHERE h.part_request_id = pr.id
        AND h.comment = 'Статус заявки зафиксирован при демонстрационном наполнении'
  );

INSERT INTO vehicle_part_installations (
    part_id, vehicle_id, installed_at, quantity, installed_by_user_id, is_active,
    planned_replacement_at, mechanic_shift_id, unit_price, total_price
)
SELECT
    pc.id,
    veh.id,
    v.installed_at::date,
    v.quantity,
    u.id,
    v.is_active,
    v.planned_replacement_at::date,
    ms.id,
    pc.price,
    pc.price * v.quantity
FROM (VALUES
    ('DEMO-BRK-PAD-FR', 'B-101', '2026-05-06', 1, 'mechanic1@autopark.demo', TRUE, '2026-11-06'),
    ('DEMO-FILTER-OIL', 'B-102', '2026-05-07', 2, 'mechanic2@autopark.demo', TRUE, '2026-08-07'),
    ('DEMO-OIL-5W30', 'B-102', '2026-05-07', 3, 'mechanic2@autopark.demo', FALSE, '2026-08-07'),
    ('DEMO-FILTER-AIR', 'B-103', '2026-05-08', 1, 'mechanic3@autopark.demo', TRUE, '2026-08-08'),
    ('DEMO-LAMP-H7', 'B-104', '2026-05-09', 2, 'mechanic4@autopark.demo', TRUE, '2027-05-09'),
    ('DEMO-COOLANT-G12', 'B-106', '2026-05-11', 2, 'mechanic2@autopark.demo', FALSE, '2026-11-11'),
    ('DEMO-WIPER-650', 'B-107', '2026-05-12', 2, 'mechanic3@autopark.demo', TRUE, '2026-11-12'),
    ('DEMO-FILTER-FUEL', 'B-109', '2026-05-14', 1, 'mechanic1@autopark.demo', TRUE, '2026-08-14'),
    ('DEMO-BELT-GEN', 'B-110', '2026-05-15', 1, 'mechanic2@autopark.demo', TRUE, '2027-05-15'),
    ('DEMO-ADBLUE-10L', 'B-113', '2026-05-18', 4, 'mechanic4@autopark.demo', FALSE, '2026-07-18'),
    ('DEMO-PUMP-WATER', 'B-114', '2026-05-19', 1, 'mechanic1@autopark.demo', TRUE, '2027-05-19'),
    ('DEMO-FUSE-KIT', 'B-116', '2026-05-21', 1, 'mechanic3@autopark.demo', TRUE, '2027-05-21'),
    ('DEMO-SHOCK-RR', 'B-117', '2026-05-22', 2, 'mechanic4@autopark.demo', TRUE, '2027-05-22'),
    ('DEMO-BRK-DISC-FR', 'B-118', '2026-05-23', 1, 'mechanic1@autopark.demo', TRUE, '2027-05-23'),
    ('DEMO-PADS-DRUM', 'B-119', '2026-05-24', 1, 'mechanic2@autopark.demo', TRUE, '2027-05-24')
) AS v(part_code, board_number, installed_at, quantity, mechanic_email, is_active, planned_replacement_at)
JOIN parts_catalog pc ON pc.part_id = v.part_code
JOIN vehicles veh ON veh.board_number = v.board_number
JOIN users u ON u.email = v.mechanic_email
LEFT JOIN LATERAL (
    SELECT id
    FROM mechanic_shifts ms
    WHERE ms.user_id = u.id
      AND ms.shift_date = v.installed_at::date
    ORDER BY ms.id DESC
    LIMIT 1
) ms ON TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM vehicle_part_installations vi
    WHERE vi.part_id = pc.id AND vi.vehicle_id = veh.id AND vi.installed_at = v.installed_at::date
);

INSERT INTO incidents (
    incident_type_id, vehicle_id, driver_id, mechanic_id, incident_date, incident_time,
    location, description, tripsheet_id, mechanic_shift_id
)
SELECT it.id, veh.id, d.id, mech.id, v.incident_date::date, v.incident_time::time,
       v.location, v.description, t.id, ms.id
FROM (VALUES
    ('Поломка', 'B-103', '900305300003', 'mechanic3@autopark.demo', '2026-05-08', '15:20', 'Сервисная зона Майлина', 'На маршруте загорелась индикация по давлению масла, автомобиль отправлен на диагностику', 'WL-2026-0003'),
    ('Повреждение', 'B-105', '760527300005', 'mechanic1@autopark.demo', '2026-05-10', '12:10', 'Парковка районной базы', 'Царапина на правом борту после маневра на тесной парковке', 'WL-2026-0005'),
    ('Поломка', 'B-111', '910103300011', 'mechanic3@autopark.demo', '2026-05-16', '16:40', 'Технопарк, блок С', 'Посторонний стук передней подвески при движении по неровной дороге', 'WL-2026-0011'),
    ('Повреждение', 'B-115', '890217300015', 'mechanic2@autopark.demo', '2026-05-20', '10:05', 'Сервисная база', 'Поврежден корпус левого зеркала, требуется акт осмотра', 'WL-2026-0015'),
    ('ДТП', 'B-114', '830706300014', 'mechanic1@autopark.demo', '2026-05-19', '15:25', 'Восточная объездная дорога', 'Легкое касание ограждения без пострадавших, машина поставлена на ремонт', 'WL-2026-0014')
) AS v(type_name, board_number, driver_iin, mechanic_email, incident_date, incident_time, location, description, tripsheet_number)
JOIN incident_types it ON it.name = v.type_name
JOIN vehicles veh ON veh.board_number = v.board_number
JOIN drivers d ON d.iin = v.driver_iin
JOIN users mech ON mech.email = v.mechanic_email
LEFT JOIN tripsheets t ON t.tripsheet_number = v.tripsheet_number
LEFT JOIN LATERAL (
    SELECT id
    FROM mechanic_shifts ms
    WHERE ms.user_id = mech.id
      AND ms.shift_date = v.incident_date::date
    ORDER BY ms.id DESC
    LIMIT 1
) ms ON TRUE
WHERE NOT EXISTS (
    SELECT 1 FROM incidents i
    WHERE i.vehicle_id = veh.id AND i.incident_date = v.incident_date::date AND i.incident_time = v.incident_time::time
);

INSERT INTO invoices (invoice_number, invoice_date, part_request_id, request_number)
SELECT v.invoice_number, v.invoice_date::date, pr.id, v.request_number
FROM (VALUES
    ('INV-DEMO-223780', '2026-05-07', 'ТО B-102: замена масляного фильтра', 'REQ-DEMO-0002'),
    ('INV-DEMO-223781', '2026-05-07', 'ТО B-102: моторное масло по регламенту', 'REQ-DEMO-0003'),
    ('INV-DEMO-223782', '2026-05-09', 'B-104: перегорели лампы ближнего света', 'REQ-DEMO-0005'),
    ('INV-DEMO-223783', '2026-05-12', 'B-107: износ щеток стеклоочистителя', 'REQ-DEMO-0008'),
    ('INV-DEMO-223784', '2026-05-14', 'B-109: плановая замена топливного фильтра', 'REQ-DEMO-0010'),
    ('INV-DEMO-223785', '2026-05-18', 'B-113: пополнение AdBlue перед дальним рейсом', 'REQ-DEMO-0013'),
    ('INV-DEMO-223786', '2026-05-21', 'B-116: пополнить аварийный набор предохранителей', 'REQ-DEMO-0016')
) AS v(invoice_number, invoice_date, mechanic_comment, request_number)
JOIN part_requests pr ON pr.mechanic_comment = v.mechanic_comment
ON CONFLICT (invoice_number) DO UPDATE SET
    invoice_date = EXCLUDED.invoice_date,
    part_request_id = EXCLUDED.part_request_id,
    request_number = EXCLUDED.request_number;

INSERT INTO notifications (
    user_id, notification_type_id, title, message, context, is_readed, read_at, dedup_key
)
SELECT u.id, nt.id, v.title, v.message, v.context::jsonb, v.is_readed, v.read_at::timestamptz, v.dedup_key
FROM (VALUES
    ('warehouse@autopark.demo', 'part_request_created', 'Новая заявка на деталь', 'Механик запросил колодки для B-101', '{"board":"B-101","part":"DEMO-BRK-PAD-FR"}', TRUE, '2026-05-06 10:35:00+05', 'demo-pr-001-created'),
    ('warehouse@autopark.demo', 'part_request_approved', 'Заявка утверждена', 'Заявка по B-101 готова к выдаче', '{"board":"B-101","request":"REQ-DEMO-0001"}', TRUE, '2026-05-06 11:20:00+05', 'demo-pr-001-approved'),
    ('mechanic2@autopark.demo', 'part_request_approved', 'Детали выданы', 'Фильтр и масло по B-102 выданы со склада', '{"board":"B-102"}', TRUE, '2026-05-07 12:05:00+05', 'demo-pr-002-issued'),
    ('manager@autopark.demo', 'vehicle_part_replacement_7_days', 'Скоро замена детали', 'B-114: помпа запланирована к контролю', '{"board":"B-114","days":7}', FALSE, NULL, 'demo-vpi-b114-7d'),
    ('dispatcher@autopark.demo', 'part_request_created', 'Новая заявка', 'B-105: аккумулятор ожидает согласования', '{"board":"B-105"}', FALSE, NULL, 'demo-pr-006-created'),
    ('warehouse2@autopark.demo', 'part_request_rejected', 'Заявка отклонена', 'B-115: требуется акт осмотра зеркала', '{"board":"B-115"}', TRUE, '2026-05-20 11:20:00+05', 'demo-pr-015-rejected'),
    ('mechanic1@autopark.demo', 'vehicle_part_replacement_today', 'Проверить замену сегодня', 'B-118: тормозной диск в плане на текущий день', '{"board":"B-118"}', FALSE, NULL, 'demo-vpi-b118-today'),
    ('manager2@autopark.demo', 'part_request_created', 'Заявка на склад', 'B-120: подготовка комплекта шин', '{"board":"B-120"}', FALSE, NULL, 'demo-pr-020-created')
) AS v(email, type_code, title, message, context, is_readed, read_at, dedup_key)
JOIN users u ON u.email = v.email
JOIN notification_types nt ON nt.code = v.type_code
ON CONFLICT (user_id, dedup_key) WHERE dedup_key IS NOT NULL DO UPDATE SET
    title = EXCLUDED.title,
    message = EXCLUDED.message,
    context = EXCLUDED.context,
    is_readed = EXCLUDED.is_readed,
    read_at = EXCLUDED.read_at,
    updated_at = NOW();

INSERT INTO vehicle_services (type_id, part_id, vehicle_id, service_date)
SELECT st.id, pcoll.id, veh.id, v.service_date::date
FROM (VALUES
    ('Полировка', 'Весь кузов', 'B-101', '2026-05-03'),
    ('Тонировка', 'Боковые стекла', 'B-104', '2026-05-04'),
    ('Нанесение защитной пленки', 'Капот', 'B-109', '2026-05-05'),
    ('Антигравийная защита', 'Передний бампер', 'B-110', '2026-05-06'),
    ('Химчистка', 'Весь кузов', 'B-118', '2026-05-07'),
    ('Детейлинг', 'Весь кузов', 'B-120', '2026-05-08'),
    ('Покраска', 'Задний бампер', 'B-105', '2026-05-10'),
    ('Рихтовка', 'Переднее левое крыло', 'B-114', '2026-05-19'),
    ('Мойка двигателя', 'Капот', 'B-116', '2026-05-21'),
    ('Обработка антикором', 'Крыша', 'B-117', '2026-05-22')
) AS v(service_name, part_name, board_number, service_date)
JOIN service_types st ON st.name = v.service_name
JOIN parts_collection pcoll ON pcoll.name = v.part_name
JOIN vehicles veh ON veh.board_number = v.board_number
WHERE NOT EXISTS (
    SELECT 1 FROM vehicle_services vs
    WHERE vs.type_id = st.id AND vs.part_id = pcoll.id AND vs.vehicle_id = veh.id AND vs.service_date = v.service_date::date
);

INSERT INTO maintenance_schedules (
    is_draft, date_from, date_to, consecutive_count, consecutive_unit,
    every_count, every_unit, time_from, time_to, duration_value, duration_unit,
    limit_boards_by_time, categories, boards, mechanics
)
SELECT
    v.is_draft,
    v.date_from::date,
    v.date_to::date,
    v.consecutive_count,
    v.consecutive_unit,
    v.every_count,
    v.every_unit,
    v.time_from,
    v.time_to,
    v.duration_value,
    v.duration_unit,
    v.limit_boards_by_time,
    v.categories::jsonb,
    v.boards::jsonb,
    v.mechanics::jsonb
FROM (VALUES
    (FALSE, '2026-05-06', '2026-05-10', 1, 'day', 1, 'day', '09:00', '12:00', 90, 'minute', TRUE, '["регламентное ТО","фильтры"]', '["B-101","B-102","B-103"]', '["Руслан Абдрахманов","Николай Петров"]'),
    (FALSE, '2026-05-13', '2026-05-17', 1, 'day', 1, 'day', '10:00', '13:00', 120, 'minute', TRUE, '["подвеска","тормоза"]', '["B-109","B-110","B-111"]', '["Ермек Тлеубаев"]'),
    (FALSE, '2026-05-20', '2026-05-25', 1, 'day', 1, 'day', '08:30', '12:30', 120, 'minute', TRUE, '["предрейсовый осмотр","электрика"]', '["B-116","B-117","B-118","B-120"]', '["Данияр Касенов","Руслан Абдрахманов"]'),
    (TRUE, '2026-06-01', '2026-06-07', 1, 'day', 2, 'day', '09:00', '11:00', 60, 'minute', FALSE, '["летняя подготовка"]', '["B-104","B-105","B-106"]', '["Николай Петров"]')
) AS v(is_draft, date_from, date_to, consecutive_count, consecutive_unit, every_count, every_unit, time_from, time_to, duration_value, duration_unit, limit_boards_by_time, categories, boards, mechanics)
WHERE NOT EXISTS (
    SELECT 1 FROM maintenance_schedules ms
    WHERE ms.date_from = v.date_from::date AND ms.date_to = v.date_to::date AND ms.boards = v.boards::jsonb
);

INSERT INTO maintenance_executions (
    schedule_id, board, action, comment, responsible_mechanic
)
SELECT ms.id, v.board, v.action, v.comment, v.responsible_mechanic
FROM (VALUES
    ('2026-05-06', 'B-101', 'serviced_replaced', 'Заменены передние колодки, проверен уровень жидкостей', 'Руслан Абдрахманов'),
    ('2026-05-06', 'B-102', 'serviced_replaced', 'Масло и фильтр заменены, замечаний нет', 'Николай Петров'),
    ('2026-05-06', 'B-103', 'defect_parked', 'Требуется дополнительная диагностика давления масла', 'Руслан Абдрахманов'),
    ('2026-05-13', 'B-109', 'serviced_replaced', 'Заменен топливный фильтр', 'Ермек Тлеубаев'),
    ('2026-05-13', 'B-110', 'serviced_replaced', 'Заменен ремень генератора', 'Ермек Тлеубаев'),
    ('2026-05-13', 'B-111', 'defect_parked', 'Подвеска требует замены амортизаторов', 'Ермек Тлеубаев'),
    ('2026-05-20', 'B-116', 'serviced_replaced', 'Пополнен набор предохранителей, проверены лампы', 'Данияр Касенов'),
    ('2026-05-20', 'B-117', 'serviced_replaced', 'Заменены задние амортизаторы', 'Данияр Касенов'),
    ('2026-05-20', 'B-118', 'serviced', 'Осмотр тормозной системы, запланирована замена диска', 'Руслан Абдрахманов'),
    ('2026-05-20', 'B-120', 'serviced', 'Предрейсовый осмотр выполнен, замечаний нет', 'Руслан Абдрахманов')
) AS v(schedule_date_from, board, action, comment, responsible_mechanic)
JOIN maintenance_schedules ms ON ms.date_from = v.schedule_date_from::date
WHERE NOT EXISTS (
    SELECT 1 FROM maintenance_executions me
    WHERE me.schedule_id = ms.id AND me.board = v.board AND me.comment = v.comment
);

INSERT INTO audit_logs (level, function, from_status, to_status, actor, message, created_at)
SELECT v.level, v.function, v.from_status, v.to_status, v.actor, v.message, v.created_at::timestamptz
FROM (VALUES
    ('success', 'user', NULL, 'created', 'manager@autopark.demo', 'Создан демонстрационный профиль диспетчера', '2026-05-01 09:10:00+05'),
    ('success', 'vehicle', 'Не используется', 'В использовании', 'dispatcher@autopark.demo', 'B-101 выпущен на линию после проверки документов', '2026-05-06 08:05:00+05'),
    ('info', 'tripsheet', 'Создано', 'Окончен', 'dispatcher@autopark.demo', 'Путевой лист WL-2026-0001 закрыт без замечаний', '2026-05-06 17:15:00+05'),
    ('success', 'arrival', 'draft', 'accepted', 'warehouse@autopark.demo', 'Принята поставка ARR-DEMO-2026-001', '2026-05-02 15:35:00+05'),
    ('warning', 'incident', NULL, 'registered', 'mechanic3@autopark.demo', 'Зафиксирована неисправность B-103 на маршруте', '2026-05-08 15:30:00+05'),
    ('success', 'request', 'new', 'issued', 'warehouse@autopark.demo', 'Выданы масло и фильтр по заявке B-102', '2026-05-07 12:10:00+05'),
    ('warning', 'vehicle', 'В использовании', 'На ремонте', 'mechanic1@autopark.demo', 'B-114 переведен в ремонт после инцидента', '2026-05-19 16:20:00+05'),
    ('info', 'shift', NULL, 'active', 'dispatcher2@autopark.demo', 'Открыта активная смена водителя WL-2026-0020', '2026-05-25 07:30:00+05'),
    ('success', 'request', 'new', 'approved', 'manager@autopark.demo', 'Согласована замена задних амортизаторов B-117', '2026-05-22 12:00:00+05'),
    ('error', 'incident', NULL, 'needs_review', 'mechanic2@autopark.demo', 'Заявка по зеркалу B-115 отклонена до получения акта', '2026-05-20 11:15:00+05')
) AS v(level, function, from_status, to_status, actor, message, created_at)
WHERE NOT EXISTS (
    SELECT 1 FROM audit_logs al
    WHERE al.actor = v.actor AND al.message = v.message AND al.created_at = v.created_at::timestamptz
);
