package repository

import (
	"context"
	"fmt"
)

func (r *tireRepoImpl) UpdateByID(ctx context.Context, id int64, p UpdateTireParams) (bool, error) {
	const q = `
		UPDATE tires
		SET
			place_id = $1,
			vehicle_id = $2,
			tire = $3,
			mileage = $4,
			max_usage = $5,
			updated_at = NOW()
		WHERE id = $6;
	`

	res, err := r.db.ExecContext(
		ctx,
		q,
		p.PlaceID,
		p.VehicleID,
		p.Tire,
		p.Mileage,
		p.MaxUsage,
		id,
	)
	if err != nil {
		return false, fmt.Errorf("update tire by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update tire by id rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *tireRepoImpl) UpdateForVehicle(ctx context.Context, vehicleID int64, tireID int64, p UpdateVehicleTireParams) (bool, error) {
	const q = `
		UPDATE tires
		SET
			place_id = $1,
			tire = $2,
			mileage = $3,
			max_usage = $4,
			installed_at = $5::date,
			updated_at = NOW()
		WHERE id = $6
		  AND vehicle_id = $7;
	`

	res, err := r.db.ExecContext(ctx, q, p.PlaceID, p.Tire, p.Mileage, p.MaxUsage, p.InstalledAt, tireID, vehicleID)
	if err != nil {
		return false, fmt.Errorf("update vehicle tire: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vehicle tire rows affected: %w", err)
	}
	return aff > 0, nil
}
