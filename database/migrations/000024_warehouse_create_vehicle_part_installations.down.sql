DROP TABLE IF EXISTS vehicle_part_installations;

ALTER TABLE parts_catalog
    DROP COLUMN IF EXISTS is_consumable;
