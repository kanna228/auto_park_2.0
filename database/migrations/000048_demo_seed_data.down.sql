DELETE FROM audit_logs
WHERE actor LIKE '%@autopark.demo'
   OR message LIKE '%ARR-DEMO-%'
   OR message LIKE '%WL-2026-%';

DELETE FROM maintenance_executions
WHERE board IN (
    'B-101','B-102','B-103','B-109','B-110','B-111','B-116','B-117','B-118','B-120'
);

DELETE FROM maintenance_schedules
WHERE boards ?| ARRAY[
    'B-101','B-102','B-103','B-104','B-105','B-106','B-109','B-110','B-111','B-116','B-117','B-118','B-120'
];

DELETE FROM vehicle_services
WHERE vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120');

DELETE FROM notifications
WHERE dedup_key LIKE 'demo-%';

DELETE FROM invoices
WHERE invoice_number LIKE 'INV-DEMO-%';

DELETE FROM incidents
WHERE vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120')
  AND incident_date BETWEEN DATE '2026-05-01' AND DATE '2026-05-31';

DELETE FROM vehicle_part_installations
WHERE part_id IN (SELECT id FROM parts_catalog WHERE part_id LIKE 'DEMO-%')
   OR vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120');

DELETE FROM part_request_history
WHERE part_request_id IN (
    SELECT pr.id
    FROM part_requests pr
    JOIN parts_catalog pc ON pc.id = pr.part_id
    WHERE pc.part_id LIKE 'DEMO-%'
);

DELETE FROM part_requests
WHERE part_id IN (SELECT id FROM parts_catalog WHERE part_id LIKE 'DEMO-%')
   OR author_user_id IN (SELECT id FROM users WHERE email LIKE '%@autopark.demo');

DELETE FROM part_stock_movements
WHERE document_number LIKE 'ARR-DEMO-%'
   OR part_id IN (SELECT id FROM parts_catalog WHERE part_id LIKE 'DEMO-%');

DELETE FROM part_arrival_items
WHERE arrival_id IN (SELECT id FROM part_arrivals WHERE document_number LIKE 'ARR-DEMO-%')
   OR part_id IN (SELECT id FROM parts_catalog WHERE part_id LIKE 'DEMO-%');

DELETE FROM part_arrivals
WHERE document_number LIKE 'ARR-DEMO-%';

DELETE FROM fuel_refills
WHERE tripsheet_id IN (SELECT id FROM tripsheets WHERE tripsheet_number LIKE 'WL-2026-%');

DELETE FROM waybill_route_points
WHERE waybill_id IN (SELECT id FROM tripsheets WHERE tripsheet_number LIKE 'WL-2026-%');

DELETE FROM tripsheet_trips
WHERE tripsheet_id IN (SELECT id FROM tripsheets WHERE tripsheet_number LIKE 'WL-2026-%');

DELETE FROM tripsheets
WHERE tripsheet_number LIKE 'WL-2026-%';

DELETE FROM driver_shifts
WHERE driver_id IN (SELECT id FROM drivers WHERE mail LIKE '%@autopark.demo')
  AND shift_date BETWEEN DATE '2026-05-01' AND DATE '2026-05-31';

DELETE FROM mechanic_shifts
WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'mechanic%@autopark.demo')
  AND shift_date BETWEEN DATE '2026-05-01' AND DATE '2026-05-31';

DELETE FROM tires
WHERE tire LIKE '%DEMO-%';

DELETE FROM fuel_norms
WHERE vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120');

DELETE FROM vehicle_documents
WHERE number LIKE 'DOC-INS-B%'
   OR number LIKE 'DOC-TCH-B%';

DELETE FROM technical_inspection
WHERE vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120');

DELETE FROM insurance
WHERE vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120');

DELETE FROM vehicle_status_history
WHERE vehicle_id IN (SELECT id FROM vehicles WHERE board_number BETWEEN 'B-101' AND 'B-120');

DELETE FROM vehicles
WHERE board_number BETWEEN 'B-101' AND 'B-120';

DELETE FROM parts_catalog
WHERE part_id LIKE 'DEMO-%';

DELETE FROM drivers
WHERE mail LIKE '%@autopark.demo';

DELETE FROM users
WHERE email LIKE '%@autopark.demo';
