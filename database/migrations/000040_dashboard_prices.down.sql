DROP INDEX IF EXISTS idx_vehicle_part_installations_total_price;

ALTER TABLE vehicle_part_installations
    DROP COLUMN IF EXISTS total_price,
    DROP COLUMN IF EXISTS unit_price;

DROP INDEX IF EXISTS idx_parts_catalog_price;

ALTER TABLE parts_catalog
    DROP COLUMN IF EXISTS price;

DROP INDEX IF EXISTS idx_fuel_refills_price;

ALTER TABLE fuel_refills
    DROP COLUMN IF EXISTS price;
