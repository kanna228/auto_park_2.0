CREATE TABLE IF NOT EXISTS parts_catalog (
    id BIGSERIAL PRIMARY KEY,
    part_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL UNIQUE,
    quantity BIGINT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    category TEXT NOT NULL,
    dimensions TEXT,
    manufacturer TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_parts_catalog_part_id ON parts_catalog(part_id);
CREATE INDEX IF NOT EXISTS idx_parts_catalog_name ON parts_catalog(name);
CREATE INDEX IF NOT EXISTS idx_parts_catalog_category ON parts_catalog(category);
