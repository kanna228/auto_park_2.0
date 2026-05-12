UPDATE vehicle_part_installations
SET planned_replacement_at = installed_at
WHERE planned_replacement_at IS NULL;

ALTER TABLE vehicle_part_installations
    ALTER COLUMN planned_replacement_at SET NOT NULL;
