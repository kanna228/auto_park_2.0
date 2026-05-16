package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/fuel_norm_module/dto"
)

var ErrFuelNormDuplicateVehicle = errors.New("fuel norm for vehicle already exists")
var ErrFuelNormVehicleNotFound = errors.New("vehicle not found")

type FuelNormRepository interface {
	Create(ctx context.Context, req dto.FuelNormRequest) (*dto.FuelNormResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.FuelNormResponse, error)
	List(ctx context.Context, limit int, offset int) ([]dto.FuelNormResponse, int64, error)
	Update(ctx context.Context, id int64, req dto.FuelNormRequest) (*dto.FuelNormResponse, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type fuelNormRepo struct {
	db *sql.DB
}

func NewFuelNormRepository(db *sql.DB) FuelNormRepository {
	return &fuelNormRepo{db: db}
}

func (r *fuelNormRepo) Create(ctx context.Context, req dto.FuelNormRequest) (*dto.FuelNormResponse, error) {
	const q = `
		INSERT INTO fuel_norms (
			vehicle_id,
			norm_per_100km,
			summer_norm,
			winter_norm,
			cold_air_norm,
			warm_air_norm,
			deviation,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, vehicle_id, norm_per_100km, summer_norm, winter_norm, cold_air_norm, warm_air_norm, deviation, created_at, updated_at;
	`

	item, err := scanFuelNorm(r.db.QueryRowContext(
		ctx,
		q,
		req.VehicleID,
		req.NormPer100KM,
		req.SummerNorm,
		req.WinterNorm,
		req.ColdAirNorm,
		req.WarmAirNorm,
		req.Deviation,
	))
	if err != nil {
		return nil, mapFuelNormError(err)
	}
	return item, nil
}

func (r *fuelNormRepo) GetByID(ctx context.Context, id int64) (*dto.FuelNormResponse, error) {
	const q = `
		SELECT id, vehicle_id, norm_per_100km, summer_norm, winter_norm, cold_air_norm, warm_air_norm, deviation, created_at, updated_at
		FROM fuel_norms
		WHERE id = $1;
	`

	item, err := scanFuelNorm(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get fuel norm by id: %w", err)
	}
	return item, nil
}

func (r *fuelNormRepo) List(ctx context.Context, limit int, offset int) ([]dto.FuelNormResponse, int64, error) {
	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fuel_norms;`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list fuel norms count: %w", err)
	}

	const q = `
		SELECT id, vehicle_id, norm_per_100km, summer_norm, winter_norm, cold_air_norm, warm_air_norm, deviation, created_at, updated_at
		FROM fuel_norms
		ORDER BY id ASC
		LIMIT $1 OFFSET $2;
	`
	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list fuel norms: %w", err)
	}
	defer rows.Close()

	items := make([]dto.FuelNormResponse, 0)
	for rows.Next() {
		item, err := scanFuelNorm(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list fuel norms scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list fuel norms rows: %w", err)
	}
	return items, total, nil
}

func (r *fuelNormRepo) Update(ctx context.Context, id int64, req dto.FuelNormRequest) (*dto.FuelNormResponse, error) {
	const q = `
		UPDATE fuel_norms
		SET vehicle_id = $1,
			norm_per_100km = $2,
			summer_norm = $3,
			winter_norm = $4,
			cold_air_norm = $5,
			warm_air_norm = $6,
			deviation = $7,
			updated_at = NOW()
		WHERE id = $8
		RETURNING id, vehicle_id, norm_per_100km, summer_norm, winter_norm, cold_air_norm, warm_air_norm, deviation, created_at, updated_at;
	`

	item, err := scanFuelNorm(r.db.QueryRowContext(
		ctx,
		q,
		req.VehicleID,
		req.NormPer100KM,
		req.SummerNorm,
		req.WinterNorm,
		req.ColdAirNorm,
		req.WarmAirNorm,
		req.Deviation,
		id,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapFuelNormError(err)
	}
	return item, nil
}

func (r *fuelNormRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM fuel_norms WHERE id = $1;`, id)
	if err != nil {
		return false, fmt.Errorf("delete fuel norm: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete fuel norm rows affected: %w", err)
	}
	return affected > 0, nil
}

type fuelNormScanner interface {
	Scan(dest ...any) error
}

func scanFuelNorm(scanner fuelNormScanner) (*dto.FuelNormResponse, error) {
	var item dto.FuelNormResponse
	if err := scanner.Scan(
		&item.ID,
		&item.VehicleID,
		&item.NormPer100KM,
		&item.SummerNorm,
		&item.WinterNorm,
		&item.ColdAirNorm,
		&item.WarmAirNorm,
		&item.Deviation,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func mapFuelNormError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "fuel_norms_vehicle_id_key"):
		return ErrFuelNormDuplicateVehicle
	case strings.Contains(msg, "fuel_norms_vehicle_id_fkey"):
		return ErrFuelNormVehicleNotFound
	default:
		return err
	}
}
