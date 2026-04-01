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
