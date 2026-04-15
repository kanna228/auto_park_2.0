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
	GetAllByTripsheetID(ctx context.Context, tripsheetID int64) ([]dto.FuelRefillResponse, int, error)
	GetAllByVehicleID(ctx context.Context, vehicleID int64) ([]dto.FuelRefillResponse, int, error)
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
	baseQuery := fmt.Sprintf(`
		SELECT id, tripsheet_id, vehicle_id, fuel_amount, date, time, location, created_at, updated_at
		FROM %s
	`, r.table())

	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	argPos := 1

	if filter.TripsheetID != nil && *filter.TripsheetID > 0 {
		conditions = append(conditions, fmt.Sprintf("tripsheet_id = $%d", argPos))
		args = append(args, *filter.TripsheetID)
		argPos++
	}
	if filter.VehicleID != nil && *filter.VehicleID > 0 {
		conditions = append(conditions, fmt.Sprintf("vehicle_id = $%d", argPos))
		args = append(args, *filter.VehicleID)
		argPos++
	}
	if filter.DateFrom != nil && strings.TrimSpace(*filter.DateFrom) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*filter.DateFrom))
		if err != nil {
			return nil, 0, fmt.Errorf("invalid date_from format, expected YYYY-MM-DD")
		}
		conditions = append(conditions, fmt.Sprintf("date >= $%d", argPos))
		args = append(args, parsed)
		argPos++
	}
	if filter.DateTo != nil && strings.TrimSpace(*filter.DateTo) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*filter.DateTo))
		if err != nil {
			return nil, 0, fmt.Errorf("invalid date_to format, expected YYYY-MM-DD")
		}
		conditions = append(conditions, fmt.Sprintf("date <= $%d", argPos))
		args = append(args, parsed)
		argPos++
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}
	baseQuery += " ORDER BY date DESC, time DESC, id DESC"

	return r.scanMany(ctx, baseQuery, args...)
}

func (r *fuelRefillRepository) GetAllByTripsheetID(ctx context.Context, tripsheetID int64) ([]dto.FuelRefillResponse, int, error) {
	query := fmt.Sprintf(`
		SELECT id, tripsheet_id, vehicle_id, fuel_amount, date, time, location, created_at, updated_at
		FROM %s
		WHERE tripsheet_id = $1
		ORDER BY date DESC, time DESC, id DESC
	`, r.table())

	return r.scanMany(ctx, query, tripsheetID)
}

func (r *fuelRefillRepository) GetAllByVehicleID(ctx context.Context, vehicleID int64) ([]dto.FuelRefillResponse, int, error) {
	query := fmt.Sprintf(`
		SELECT id, tripsheet_id, vehicle_id, fuel_amount, date, time, location, created_at, updated_at
		FROM %s
		WHERE vehicle_id = $1
		ORDER BY date DESC, time DESC, id DESC
	`, r.table())

	return r.scanMany(ctx, query, vehicleID)
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

func (r *fuelRefillRepository) scanMany(ctx context.Context, query string, args ...interface{}) ([]dto.FuelRefillResponse, int, error) {
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

	return items, len(items), nil
}
