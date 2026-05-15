ALTER TABLE fuel_refills
    ADD COLUMN IF NOT EXISTS price NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (price >= 0);

CREATE INDEX IF NOT EXISTS idx_fuel_refills_price
    ON fuel_refills(price);

ALTER TABLE parts_catalog
    ADD COLUMN IF NOT EXISTS price NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (price >= 0);

CREATE INDEX IF NOT EXISTS idx_parts_catalog_price
    ON parts_catalog(price);

ALTER TABLE vehicle_part_installations
    ADD COLUMN IF NOT EXISTS unit_price NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (unit_price >= 0),
    ADD COLUMN IF NOT EXISTS total_price NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (total_price >= 0);

UPDATE vehicle_part_installations vpi
SET unit_price = COALESCE(pc.price, 0),
    total_price = COALESCE(pc.price, 0) * vpi.quantity,
    updated_at = NOW()
FROM parts_catalog pc
WHERE pc.id = vpi.part_id
  AND (vpi.unit_price = 0 OR vpi.total_price = 0);

CREATE INDEX IF NOT EXISTS idx_vehicle_part_installations_total_price
    ON vehicle_part_installations(total_price);
