package repository

import (
	"context"
	"fmt"
)

func (r *vehicleRepo) UnassignTiresByVehicleID(ctx context.Context, vehicleID int64) error {
	const q = `
		UPDATE tires
		SET
			vehicle_id = NULL,
			updated_at = NOW()
		WHERE vehicle_id = $1;
	`

	_, err := r.db.ExecContext(ctx, q, vehicleID)
	if err != nil {
		return fmt.Errorf("unassign tires by vehicle id: %w", err)
	}

	return nil
}

// DeleteByID удаляет машину по id.
// Перед удалением отвязывает все шины, у которых vehicle_id = id.
// Связанные tripsheets и tripsheet_trips удаляются каскадно на уровне БД
// через FOREIGN KEY ... ON DELETE CASCADE.
func (r *vehicleRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx delete vehicle: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const unassignTiresQ = `
		UPDATE tires
		SET
			vehicle_id = NULL,
			updated_at = NOW()
		WHERE vehicle_id = $1;
	`

	if _, err := tx.ExecContext(ctx, unassignTiresQ, id); err != nil {
		return false, fmt.Errorf("unassign tires before deleting vehicle: %w", err)
	}

	const deleteVehicleQ = `
		UPDATE vehicles
		SET is_archived = TRUE,
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND is_archived = FALSE;
	`

	res, err := tx.ExecContext(ctx, deleteVehicleQ, id)
	if err != nil {
		return false, fmt.Errorf("archive vehicle by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("archive vehicle by id rows affected: %w", err)
	}

	if aff == 0 {
		return false, nil
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit archive vehicle: %w", err)
	}

	return true, nil
}
