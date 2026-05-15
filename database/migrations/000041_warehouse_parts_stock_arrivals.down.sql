DROP TABLE IF EXISTS part_arrival_items;
DROP TABLE IF EXISTS part_arrivals;
DROP TABLE IF EXISTS part_stock_movements;

ALTER TABLE parts_catalog
    DROP COLUMN IF EXISTS unit,
    DROP COLUMN IF EXISTS min_stock_quantity;
