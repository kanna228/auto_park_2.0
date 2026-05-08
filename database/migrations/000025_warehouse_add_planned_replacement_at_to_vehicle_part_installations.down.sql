DROP INDEX IF EXISTS idx_vehicle_part_installations_planned_replacement_at;

ALTER TABLE vehicle_part_installations
    DROP COLUMN IF EXISTS planned_replacement_at;