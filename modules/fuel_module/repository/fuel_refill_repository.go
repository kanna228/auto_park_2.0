package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/fuel_module/dto"
	"auto_park/modules/fuel_module/models"
)

type FuelRefillRepository interface {
	Create(ctx context.Context, input models.CreateFuelRefillInput) (*models.FuelRefill, error)
	Update(ctx context.Context, input models.UpdateFuelRefillInput) (*models.FuelRefill, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dto.FuelRefillResponse, error)
	GetAll(ctx context.Context, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error)
	GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error)
	GetAllByVehicleID(ctx context.Context, vehicleID int64, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error)
}

type fuelRefillRepository struct {
	db *sql.DB
}

func NewFuelRefillRepository(db *sql.DB) FuelRefillRepository {
	return &fuelRefillRepository{db: db}
}

func (r *fuelRefillRepository) table() string { return "fuel_refills" }

func (r *fuelRefillRepository) Create(ctx context.Context, input models.CreateFuelRefillInput) (*models.FuelRefill, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			tripsheet_id,
			vehicle_id,
			fuel_amount,
			date,
			time,
			location
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, tripsheet_id, vehicle_id, fuel_amount, date, time, location, created_at, updated_at
	`, r.table())

	var item models.FuelRefill
	var location sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.TripsheetID,
		input.VehicleID,
		input.FuelAmount,
		input.Date,
		input.Time,
		input.Location,
	).Scan(
		&item.ID,
		&item.TripsheetID,
		&item.VehicleID,
		&item.FuelAmount,
		&item.Date,
		&item.Time,
		&location,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create fuel refill: %w", err)
	}

	if location.Valid {
		item.Location = &location.String
	}

	return &item, nil
}

func (r *fuelRefillRepository) Update(ctx context.Context, input models.UpdateFuelRefillInput) (*models.FuelRefill, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET
			tripsheet_id = $1,
			vehicle_id = $2,
			fuel_amount = $3,
			date = $4,
			time = $5,
			location = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING id, tripsheet_id, vehicle_id, fuel_amount, date, time, location, created_at, updated_at
	`, r.table())

	var item models.FuelRefill
	var location sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.TripsheetID,
		input.VehicleID,
		input.FuelAmount,
		input.Date,
		input.Time,
		input.Location,
		input.ID,
	).Scan(
		&item.ID,
		&item.TripsheetID,
		&item.VehicleID,
		&item.FuelAmount,
		&item.Date,
		&item.Time,
		&location,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("update fuel refill: %w", err)
	}

	if location.Valid {
		item.Location = &location.String
	}

	return &item, nil
}

func (r *fuelRefillRepository) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.table())

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete fuel refill: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected after delete fuel refill: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *fuelRefillRepository) GetByID(ctx context.Context, id int64) (*dto.FuelRefillResponse, error) {
	query := fmt.Sprintf(`
		SELECT id, tripsheet_id, vehicle_id, fuel_amount, date, time, location, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, r.table())

	return r.scanOne(ctx, query, id)
}

func (r *fuelRefillRepository) GetAll(ctx context.Context, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error) {
	return r.listFuelRefills(ctx, filter)
}

func (r *fuelRefillRepository) GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error) {
	filter.TripsheetID = &tripsheetID
	return r.listFuelRefills(ctx, filter)
}

func (r *fuelRefillRepository) GetAllByVehicleID(ctx context.Context, vehicleID int64, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error) {
	filter.VehicleID = &vehicleID
	return r.listFuelRefills(ctx, filter)
}

func (r *fuelRefillRepository) listFuelRefills(ctx context.Context, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error) {
	where := []string{"1=1"}
	args := make([]any, 0, 10)
	argPos := 1

	add := func(condition string, value any) {
		where = append(where, fmt.Sprintf(condition, argPos))
		args = append(args, value)
		argPos++
	}

	if filter.TripsheetID != nil && *filter.TripsheetID > 0 {
		add("fr.tripsheet_id = $%d", *filter.TripsheetID)
	}
	if filter.VehicleID != nil && *filter.VehicleID > 0 {
		add("fr.vehicle_id = $%d", *filter.VehicleID)
	}
	if filter.DriverID != nil && *filter.DriverID > 0 {
		add("t.driver_id = $%d", *filter.DriverID)
	}
	if filter.DateFrom != nil && strings.TrimSpace(*filter.DateFrom) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*filter.DateFrom))
		if err != nil {
			return nil, 0, fmt.Errorf("invalid date_from format, expected YYYY-MM-DD")
		}
		add("fr.date >= $%d", parsed)
	}
	if filter.DateTo != nil && strings.TrimSpace(*filter.DateTo) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*filter.DateTo))
		if err != nil {
			return nil, 0, fmt.Errorf("invalid date_to format, expected YYYY-MM-DD")
		}
		add("fr.date <= $%d", parsed)
	}

	whereSQL := strings.Join(where, " AND ")
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s fr
		LEFT JOIN tripsheets t ON t.id = fr.tripsheet_id
		WHERE %s;
	`, r.table(), whereSQL)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count fuel refills: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	sortBy := normalizeFuelRefillSortBy(filter.SortBy)
	order := normalizeFuelRefillOrder(filter.Order)
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT fr.id, fr.tripsheet_id, fr.vehicle_id, fr.fuel_amount, fr.date, fr.time, fr.location, fr.created_at, fr.updated_at
		FROM %s fr
		LEFT JOIN tripsheets t ON t.id = fr.tripsheet_id
		WHERE %s
		ORDER BY %s %s, fr.id DESC
		LIMIT $%d OFFSET $%d;
	`, r.table(), whereSQL, sortBy, order, len(args)-1, len(args))

	return r.scanMany(ctx, total, query, args...)
}

func normalizeFuelRefillSortBy(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "id":
		return "fr.id"
	case "tripsheet_id":
		return "fr.tripsheet_id"
	case "vehicle_id":
		return "fr.vehicle_id"
	case "driver_id":
		return "t.driver_id"
	case "fuel_amount":
		return "fr.fuel_amount"
	case "time":
		return "fr.time"
	case "created_at":
		return "fr.created_at"
	case "updated_at":
		return "fr.updated_at"
	default:
		return "fr.date"
	}
}

func normalizeFuelRefillOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "ASC"
	}
	return "DESC"
}

func (r *fuelRefillRepository) scanOne(ctx context.Context, query string, args ...interface{}) (*dto.FuelRefillResponse, error) {
	var (
		item      dto.FuelRefillResponse
		dateVal   time.Time
		timeVal   time.Time
		createdAt time.Time
		updatedAt time.Time
		location  sql.NullString
	)

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&item.ID,
		&item.TripsheetID,
		&item.VehicleID,
		&item.FuelAmount,
		&dateVal,
		&timeVal,
		&location,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	item.Date = dateVal.Format("2006-01-02")
	item.Time = timeVal.Format("15:04:05")
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	if location.Valid {
		item.Location = &location.String
	}

	return &item, nil
}

func (r *fuelRefillRepository) scanMany(ctx context.Context, total int, query string, args ...interface{}) ([]dto.FuelRefillResponse, int, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query fuel refills: %w", err)
	}
	defer rows.Close()

	items := make([]dto.FuelRefillResponse, 0)
	for rows.Next() {
		var (
			item      dto.FuelRefillResponse
			dateVal   time.Time
			timeVal   time.Time
			createdAt time.Time
			updatedAt time.Time
			location  sql.NullString
		)

		if err := rows.Scan(
			&item.ID,
			&item.TripsheetID,
			&item.VehicleID,
			&item.FuelAmount,
			&dateVal,
			&timeVal,
			&location,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan fuel refill: %w", err)
		}

		item.Date = dateVal.Format("2006-01-02")
		item.Time = timeVal.Format("15:04:05")
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		if location.Valid {
			item.Location = &location.String
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate fuel refills: %w", err)
	}

	return items, total, nil
}
