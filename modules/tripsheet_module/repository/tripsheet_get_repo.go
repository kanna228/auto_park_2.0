package repository

import (
	"auto_park/modules/tripsheet_module/dto"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (r *tripsheetRepo) GetAll(ctx context.Context, f dto.TripsheetFilter) ([]dto.TripsheetResponse, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 16)
	argPos := 1

	add := func(condition string, value any) {
		where = append(where, fmt.Sprintf(condition, argPos))
		args = append(args, value)
		argPos++
	}

	if f.VehicleID != nil {
		add("vehicle_id = $%d", *f.VehicleID)
	}
	if f.DriverID != nil {
		add("driver_id = $%d", *f.DriverID)
	}
	if f.StatusID != nil {
		add("status_id = $%d", *f.StatusID)
	}
	if f.DateFrom != nil && strings.TrimSpace(*f.DateFrom) != "" {
		add("tripsheet_date >= $%d", strings.TrimSpace(*f.DateFrom))
	}
	if f.DateTo != nil && strings.TrimSpace(*f.DateTo) != "" {
		add("tripsheet_date <= $%d", strings.TrimSpace(*f.DateTo))
	}
	if f.TripsheetNumber != nil && strings.TrimSpace(*f.TripsheetNumber) != "" {
		add("tripsheet_number ILIKE $%d", "%"+strings.TrimSpace(*f.TripsheetNumber)+"%")
	}
	if f.VehiclePlateNumber != nil && strings.TrimSpace(*f.VehiclePlateNumber) != "" {
		add("vehicle_plate_number ILIKE $%d", "%"+strings.TrimSpace(*f.VehiclePlateNumber)+"%")
	}
	if f.VehicleBrand != nil && strings.TrimSpace(*f.VehicleBrand) != "" {
		add("vehicle_brand ILIKE $%d", "%"+strings.TrimSpace(*f.VehicleBrand)+"%")
	}
	if f.DriverLastName != nil && strings.TrimSpace(*f.DriverLastName) != "" {
		add("driver_last_name ILIKE $%d", "%"+strings.TrimSpace(*f.DriverLastName)+"%")
	}
	if f.DriverFirstName != nil && strings.TrimSpace(*f.DriverFirstName) != "" {
		add("driver_first_name ILIKE $%d", "%"+strings.TrimSpace(*f.DriverFirstName)+"%")
	}
	if f.DriverMiddleName != nil && strings.TrimSpace(*f.DriverMiddleName) != "" {
		add("driver_middle_name ILIKE $%d", "%"+strings.TrimSpace(*f.DriverMiddleName)+"%")
	}

	whereSQL := strings.Join(where, " AND ")
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE %s;
	`, r.tripsheetsTable(), whereSQL)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tripsheets: %w", err)
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := normalizeTripsheetSortBy(f.SortBy)
	order := normalizeTripsheetOrder(f.Order)

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
		WHERE %s
		ORDER BY %s %s, id DESC
		LIMIT $%d OFFSET $%d;
	`, r.tripsheetsTable(), whereSQL, sortBy, order, argPos, argPos+1)

	args = append(args, limit, offset)
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
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func normalizeTripsheetSortBy(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "id":
		return "id"
	case "tripsheet_number":
		return "tripsheet_number"
	case "vehicle_id":
		return "vehicle_id"
	case "driver_id":
		return "driver_id"
	case "status_id":
		return "status_id"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	default:
		return "tripsheet_date"
	}
}

func normalizeTripsheetOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "ASC"
	}
	return "DESC"
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
