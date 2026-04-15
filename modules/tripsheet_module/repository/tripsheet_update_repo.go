package repository

import (
	"auto_park/modules/tripsheet_module/models"
	"context"
	"database/sql"
	"fmt"
)

func (r *tripsheetRepo) Update(ctx context.Context, input models.UpdateTripsheetInput) (*models.Tripsheet, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET
			tripsheet_number = $1,
			tripsheet_date = $2,
			vehicle_id = $3,
			vehicle_brand = $4,
			vehicle_plate_number = $5,
			driver_last_name = $6,
			driver_first_name = $7,
			driver_middle_name = $8,
			driver_id = $9,
			start_time = $10,
			end_time = $11,
			mileage_start = $12,
			mileage_end = $13,
			fuel_start = $14,
			fuel_issued = $15,
			fuel_consumption_theoretical = $16,
			fuel_consumption_actual = $17,
			status_id = $18,
			updated_at = NOW()
		WHERE id = $19
		RETURNING
			id,
			tripsheet_number,
			tripsheet_date,
			vehicle_id,
			vehicle_brand,
			vehicle_plate_number,
			driver_last_name,
			driver_first_name,
			driver_middle_name,
			driver_id,
			start_time,
			end_time,
			mileage_start,
			mileage_end,
			fuel_start,
			fuel_issued,
			fuel_consumption_theoretical,
			fuel_consumption_actual,
			status_id,
			created_at,
			updated_at
	`, r.tripsheetsTable())

	var item models.Tripsheet

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.TripsheetNumber,
		input.TripsheetDate,
		input.VehicleID,
		input.VehicleBrand,
		input.VehiclePlateNumber,
		input.DriverLastName,
		input.DriverFirstName,
		input.DriverMiddleName,
		input.DriverID,
		input.StartTime,
		input.EndTime,
		input.MileageStart,
		input.MileageEnd,
		input.FuelStart,
		input.FuelIssued,
		input.FuelConsumptionTheoretical,
		input.FuelConsumptionActual,
		*input.StatusID,
		input.ID,
	).Scan(
		&item.ID,
		&item.TripsheetNumber,
		&item.TripsheetDate,
		&item.VehicleID,
		&item.VehicleBrand,
		&item.VehiclePlateNumber,
		&item.DriverLastName,
		&item.DriverFirstName,
		&item.DriverMiddleName,
		&item.DriverID,
		&item.StartTime,
		&item.EndTime,
		&item.MileageStart,
		&item.MileageEnd,
		&item.FuelStart,
		&item.FuelIssued,
		&item.FuelConsumptionTheoretical,
		&item.FuelConsumptionActual,
		&item.StatusID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("update tripsheet: %w", err)
	}

	return &item, nil
}
