package repository

import (
	"auto_park/modules/vehicle_module/models"
	"context"
	"fmt"
)

type VehicleStatusRepository interface {
	ListVehicleStatuses(ctx context.Context) ([]models.VehicleStatus, error)
}

func (r *vehicleRepo) ListVehicleStatuses(ctx context.Context) ([]models.VehicleStatus, error) {
	const q = `
		SELECT
			id,
			name,
			created_at,
			updated_at
		FROM vehicle_status
		ORDER BY id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list vehicle statuses: %w", err)
	}
	defer rows.Close()

	items := make([]models.VehicleStatus, 0)
	for rows.Next() {
		var item models.VehicleStatus
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle status: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows vehicle status: %w", err)
	}

	return items, nil
}
