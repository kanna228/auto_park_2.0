package repository

import (
	"auto_park/vehicle_module/models"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

func (r *vehicleRepo) GetByID(ctx context.Context, id int64) (*models.Vehicle, error) {
	const q = `
		SELECT
			v.id,
			v.board_number,
			v.technical_passport_number,
			v.state_number,
			v.vin,
			v.brand_model,
			v.manufacture_year,
			v.received_date,
			v.empty_weight_kg,
			v.max_weight_kg,
			v.engine_volume_cc,
			v.insurance_policy_number,
			v.insurance_expiry_date,
			v.mileage,
			v.current_fuel,
			v.status_id,
			vs.name AS status_name,
			v.drivers_ids,
			v.photo_path,
			v.created_at,
			v.updated_at
		FROM vehicles v
		JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE v.id = $1;
	`

	var v models.Vehicle
	var drivers pq.Int64Array
	var photoPath sql.NullString

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&v.ID,
		&v.BoardNumber,
		&v.TechnicalPassportNumber,
		&v.StateNumber,
		&v.VIN,
		&v.BrandModel,
		&v.ManufactureYear,
		&v.ReceivedDate,
		&v.EmptyWeightKG,
		&v.MaxWeightKG,
		&v.EngineVolumeCC,
		&v.InsurancePolicyNumber,
		&v.InsuranceExpiryDate,
		&v.Mileage,
		&v.CurrentFuel,
		&v.StatusID,
		&v.StatusName,
		&drivers,
		&photoPath,
		&v.CreatedAt,
		&v.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get vehicle by id: %w", err)
	}

	v.DriversIDs = []int64(drivers)
	v.PhotoPath = photoPath.String
	return &v, nil
}

type ListVehiclesParams struct {
	BoardNumber string
	StateNumber string
	VIN         string
	BrandModel  string

	StatusID *int64

	ManufactureYearFrom *int
	ManufactureYearTo   *int

	DriverID *int64

	Limit  int
	Offset int

	SortBy string
	Order  string
}

func (r *vehicleRepo) List(ctx context.Context, q ListVehiclesParams) ([]models.Vehicle, int64, error) {
	where := make([]string, 0, 10)
	args := make([]any, 0, 14)

	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if strings.TrimSpace(q.BoardNumber) != "" {
		add("v.board_number ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.BoardNumber))
	}
	if strings.TrimSpace(q.StateNumber) != "" {
		add("v.state_number ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.StateNumber))
	}
	if strings.TrimSpace(q.VIN) != "" {
		add("v.vin ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.VIN))
	}
	if strings.TrimSpace(q.BrandModel) != "" {
		add("v.brand_model ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.BrandModel))
	}

	if q.StatusID != nil {
		add("v.status_id = $%d", *q.StatusID)
	}

	if q.ManufactureYearFrom != nil {
		add("v.manufacture_year >= $%d", *q.ManufactureYearFrom)
	}
	if q.ManufactureYearTo != nil {
		add("v.manufacture_year <= $%d", *q.ManufactureYearTo)
	}

	if q.DriverID != nil {
		add("$%d = ANY(v.drivers_ids)", *q.DriverID)
	}

	whereSQL := "TRUE"
	if len(where) > 0 {
		whereSQL = strings.Join(where, " AND ")
	}

	sortCol := "v.id"
	switch q.SortBy {
	case "id":
		sortCol = "v.id"
	case "board_number":
		sortCol = "v.board_number"
	case "state_number":
		sortCol = "v.state_number"
	case "manufacture_year":
		sortCol = "v.manufacture_year"
	case "mileage":
		sortCol = "v.mileage"
	case "created_at":
		sortCol = "v.created_at"
	case "status_id":
		sortCol = "v.status_id"
	case "status_name":
		sortCol = "vs.name"
	}

	order := "DESC"
	if strings.EqualFold(q.Order, "asc") {
		order = "ASC"
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	countSQL := `
		SELECT COUNT(*)
		FROM vehicles v
		JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE ` + whereSQL + `;`

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list vehicles count: %w", err)
	}

	argsData := append([]any{}, args...)
	argsData = append(argsData, limit, offset)

	dataSQL := fmt.Sprintf(`
		SELECT
			v.id,
			v.board_number,
			v.technical_passport_number,
			v.state_number,
			v.vin,
			v.brand_model,
			v.manufacture_year,
			v.received_date,
			v.empty_weight_kg,
			v.max_weight_kg,
			v.engine_volume_cc,
			v.insurance_policy_number,
			v.insurance_expiry_date,
			v.mileage,
			v.current_fuel,
			v.status_id,
			vs.name AS status_name,
			v.drivers_ids,
			v.photo_path,
			v.created_at,
			v.updated_at
		FROM vehicles v
		JOIN vehicle_status vs ON vs.id = v.status_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortCol, order, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, dataSQL, argsData...)
	if err != nil {
		return nil, 0, fmt.Errorf("list vehicles query: %w", err)
	}
	defer rows.Close()

	items := make([]models.Vehicle, 0, limit)
	for rows.Next() {
		var v models.Vehicle
		var drivers pq.Int64Array
		var photoPath sql.NullString

		if err := rows.Scan(
			&v.ID,
			&v.BoardNumber,
			&v.TechnicalPassportNumber,
			&v.StateNumber,
			&v.VIN,
			&v.BrandModel,
			&v.ManufactureYear,
			&v.ReceivedDate,
			&v.EmptyWeightKG,
			&v.MaxWeightKG,
			&v.EngineVolumeCC,
			&v.InsurancePolicyNumber,
			&v.InsuranceExpiryDate,
			&v.Mileage,
			&v.CurrentFuel,
			&v.StatusID,
			&v.StatusName,
			&drivers,
			&photoPath,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list vehicles scan: %w", err)
		}

		v.DriversIDs = []int64(drivers)
		v.PhotoPath = photoPath.String
		items = append(items, v)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list vehicles rows: %w", err)
	}

	return items, total, nil
}
