package repository

import (
	"auto_park/tripsheet_module/dto"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *tripsheetRepo) GetAll(ctx context.Context, f dto.TripsheetFilter) ([]dto.TripsheetResponse, int, error) {
	query := fmt.Sprintf(`
		SELECT
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
		FROM %s
		WHERE 1=1
	`, r.tripsheetsTable())

	var args []any
	i := 1

	if f.VehicleID != nil {
		query += fmt.Sprintf(" AND vehicle_id = $%d", i)
		args = append(args, *f.VehicleID)
		i++
	}
	if f.DriverID != nil {
		query += fmt.Sprintf(" AND driver_id = $%d", i)
		args = append(args, *f.DriverID)
		i++
	}
	if f.StatusID != nil {
		query += fmt.Sprintf(" AND status_id = $%d", i)
		args = append(args, *f.StatusID)
		i++
	}
	if f.DateFrom != nil && strings.TrimSpace(*f.DateFrom) != "" {
		query += fmt.Sprintf(" AND tripsheet_date >= $%d", i)
		args = append(args, strings.TrimSpace(*f.DateFrom))
		i++
	}
	if f.DateTo != nil && strings.TrimSpace(*f.DateTo) != "" {
		query += fmt.Sprintf(" AND tripsheet_date <= $%d", i)
		args = append(args, strings.TrimSpace(*f.DateTo))
		i++
	}
	if f.TripsheetNumber != nil && strings.TrimSpace(*f.TripsheetNumber) != "" {
		query += fmt.Sprintf(" AND tripsheet_number ILIKE $%d", i)
		args = append(args, "%"+strings.TrimSpace(*f.TripsheetNumber)+"%")
		i++
	}
	if f.VehiclePlateNumber != nil && strings.TrimSpace(*f.VehiclePlateNumber) != "" {
		query += fmt.Sprintf(" AND vehicle_plate_number ILIKE $%d", i)
		args = append(args, "%"+strings.TrimSpace(*f.VehiclePlateNumber)+"%")
		i++
	}
	if f.VehicleBrand != nil && strings.TrimSpace(*f.VehicleBrand) != "" {
		query += fmt.Sprintf(" AND vehicle_brand ILIKE $%d", i)
		args = append(args, "%"+strings.TrimSpace(*f.VehicleBrand)+"%")
		i++
	}
	if f.DriverLastName != nil && strings.TrimSpace(*f.DriverLastName) != "" {
		query += fmt.Sprintf(" AND driver_last_name ILIKE $%d", i)
		args = append(args, "%"+strings.TrimSpace(*f.DriverLastName)+"%")
		i++
	}
	if f.DriverFirstName != nil && strings.TrimSpace(*f.DriverFirstName) != "" {
		query += fmt.Sprintf(" AND driver_first_name ILIKE $%d", i)
		args = append(args, "%"+strings.TrimSpace(*f.DriverFirstName)+"%")
		i++
	}
	if f.DriverMiddleName != nil && strings.TrimSpace(*f.DriverMiddleName) != "" {
		query += fmt.Sprintf(" AND driver_middle_name ILIKE $%d", i)
		args = append(args, "%"+strings.TrimSpace(*f.DriverMiddleName)+"%")
		i++
	}

	query += " ORDER BY id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []dto.TripsheetResponse

	for rows.Next() {
		var (
			t         dto.TripsheetResponse
			dateVal   time.Time
			startTime sql.NullTime
			endTime   sql.NullTime
			createdAt time.Time
			updatedAt time.Time
		)

		err := rows.Scan(
			&t.ID,
			&t.TripsheetNumber,
			&dateVal,
			&t.VehicleID,
			&t.VehicleBrand,
			&t.VehiclePlateNumber,
			&t.DriverLastName,
			&t.DriverFirstName,
			&t.DriverMiddleName,
			&t.DriverID,
			&startTime,
			&endTime,
			&t.MileageStart,
			&t.MileageEnd,
			&t.FuelStart,
			&t.FuelIssued,
			&t.FuelConsumptionTheoretical,
			&t.FuelConsumptionActual,
			&t.StatusID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		t.TripsheetDate = dateVal.Format("2006-01-02")

		if startTime.Valid {
			v := startTime.Time.Format(time.RFC3339)
			t.StartTime = &v
		}
		if endTime.Valid {
			v := endTime.Time.Format(time.RFC3339)
			t.EndTime = &v
		}

		t.CreatedAt = createdAt.Format(time.RFC3339)
		t.UpdatedAt = updatedAt.Format(time.RFC3339)

		result = append(result, t)
	}

	return result, len(result), nil
}

func (r *tripsheetRepo) GetByID(ctx context.Context, id int64) (*dto.TripsheetResponse, error) {
	query := fmt.Sprintf(`
		SELECT
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
		FROM %s
		WHERE id = $1
	`, r.tripsheetsTable())

	var (
		t         dto.TripsheetResponse
		dateVal   time.Time
		startTime sql.NullTime
		endTime   sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.TripsheetNumber,
		&dateVal,
		&t.VehicleID,
		&t.VehicleBrand,
		&t.VehiclePlateNumber,
		&t.DriverLastName,
		&t.DriverFirstName,
		&t.DriverMiddleName,
		&t.DriverID,
		&startTime,
		&endTime,
		&t.MileageStart,
		&t.MileageEnd,
		&t.FuelStart,
		&t.FuelIssued,
		&t.FuelConsumptionTheoretical,
		&t.FuelConsumptionActual,
		&t.StatusID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.TripsheetDate = dateVal.Format("2006-01-02")

	if startTime.Valid {
		v := startTime.Time.Format(time.RFC3339)
		t.StartTime = &v
	}
	if endTime.Valid {
		v := endTime.Time.Format(time.RFC3339)
		t.EndTime = &v
	}

	t.CreatedAt = createdAt.Format(time.RFC3339)
	t.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &t, nil
}
