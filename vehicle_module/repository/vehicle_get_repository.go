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
			id,
			board_number,
			technical_passport_number,
			state_number,
			vin,
			brand_model,
			manufacture_year,
			received_date,
			empty_weight_kg,
			max_weight_kg,
			engine_volume_cc,
			insurance_policy_number,
			insurance_expiry_date,
			mileage,
			current_fuel,
			drivers_ids,
			created_at,
			updated_at
		FROM vehicles
		WHERE id = $1;
	`

	var v models.Vehicle
	var drivers pq.Int64Array

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
		&drivers,
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
	return &v, nil
}

type ListVehiclesParams struct {
	BoardNumber string
	StateNumber string
	VIN         string
	BrandModel  string

	ManufactureYearFrom *int
	ManufactureYearTo   *int

	DriverID *int64

	Limit  int
	Offset int

	SortBy string // allowed in code
	Order  string // asc/desc
}

func (r *vehicleRepo) List(ctx context.Context, q ListVehiclesParams) ([]models.Vehicle, int64, error) {
	where := make([]string, 0, 8)
	args := make([]any, 0, 12)

	// helper add
	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	// filters (ILIKE — по подстроке)
	if strings.TrimSpace(q.BoardNumber) != "" {
		add("board_number ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.BoardNumber))
	}
	if strings.TrimSpace(q.StateNumber) != "" {
		add("state_number ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.StateNumber))
	}
	if strings.TrimSpace(q.VIN) != "" {
		add("vin ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.VIN))
	}
	if strings.TrimSpace(q.BrandModel) != "" {
		add("brand_model ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.BrandModel))
	}

	if q.ManufactureYearFrom != nil {
		add("manufacture_year >= $%d", *q.ManufactureYearFrom)
	}
	if q.ManufactureYearTo != nil {
		add("manufacture_year <= $%d", *q.ManufactureYearTo)
	}

	if q.DriverID != nil {
		add("$%d = ANY(drivers_ids)", *q.DriverID)
	}

	whereSQL := "TRUE"
	if len(where) > 0 {
		whereSQL = strings.Join(where, " AND ")
	}

	// безопасная сортировка (white-list)
	sortCol := "id"
	switch q.SortBy {
	case "id", "board_number", "state_number", "manufacture_year", "mileage", "created_at":
		sortCol = q.SortBy
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

	// total
	countSQL := `SELECT COUNT(*) FROM vehicles WHERE ` + whereSQL + `;`
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list vehicles count: %w", err)
	}

	// data
	// добавим limit/offset в конец args
	argsData := append([]any{}, args...)
	argsData = append(argsData, limit, offset)
	limitIdx := len(argsData) - 1 // не то
	_ = limitIdx

	dataSQL := fmt.Sprintf(`
		SELECT
			id,
			board_number,
			technical_passport_number,
			state_number,
			vin,
			brand_model,
			manufacture_year,
			received_date,
			empty_weight_kg,
			max_weight_kg,
			engine_volume_cc,
			insurance_policy_number,
			insurance_expiry_date,
			mileage,
			current_fuel,
			drivers_ids,
			created_at,
			updated_at
		FROM vehicles
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
			&drivers,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list vehicles scan: %w", err)
		}

		v.DriversIDs = []int64(drivers)
		items = append(items, v)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list vehicles rows: %w", err)
	}

	return items, total, nil
}
