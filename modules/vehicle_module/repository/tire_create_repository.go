package repository

import (
	"context"
	"database/sql"
	"fmt"

	"auto_park/modules/vehicle_module/models"
)

type TireRepository interface {
	Create(ctx context.Context, p CreateTireParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Tire, error)
	List(ctx context.Context, q ListTiresParams) ([]models.Tire, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateTireParams) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	GetByVehicleID(ctx context.Context, vehicleID int64, limit int, offset int) ([]models.Tire, int64, error)
	ListByVehicle(ctx context.Context, vehicleID int64) ([]models.Tire, error)
	CreateForVehicle(ctx context.Context, vehicleID int64, p CreateVehicleTireParams) (int64, error)
	UpdateForVehicle(ctx context.Context, vehicleID int64, tireID int64, p UpdateVehicleTireParams) (bool, error)
	DetachFromVehicle(ctx context.Context, vehicleID int64, tireID int64) (bool, error)
	EnsurePlace(ctx context.Context, name string) (int64, error)
}

type CreateTireParams struct {
	PlaceID   int64
	VehicleID *int64
	Tire      string
	Mileage   int64
	MaxUsage  int64
}

type ListTiresParams struct {
	VehicleID *int64
	PlaceID   *int64
	Tire      string
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type UpdateTireParams struct {
	PlaceID   int64
	VehicleID *int64
	Tire      string
	Mileage   int64
	MaxUsage  int64
}

type CreateVehicleTireParams struct {
	PlaceID     int64
	Tire        string
	Mileage     int64
	MaxUsage    int64
	InstalledAt string
}

type UpdateVehicleTireParams struct {
	PlaceID     int64
	Tire        string
	Mileage     int64
	MaxUsage    int64
	InstalledAt string
}
type tireRepoImpl struct {
	db *sql.DB
}

func NewTireRepository(db *sql.DB) TireRepository {
	return &tireRepoImpl{db: db}
}
func (r *tireRepoImpl) Create(ctx context.Context, p CreateTireParams) (int64, error) {
	const q = `
		INSERT INTO tires (
			place_id,
			vehicle_id,
			tire,
			mileage,
			max_usage
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		q,
		p.PlaceID,
		p.VehicleID,
		p.Tire,
		p.Mileage,
		p.MaxUsage,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create tire: %w", err)
	}

	return id, nil
}

func (r *tireRepoImpl) EnsurePlace(ctx context.Context, name string) (int64, error) {
	const q = `
		INSERT INTO tire_places (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, name).Scan(&id); err != nil {
		return 0, fmt.Errorf("ensure tire place: %w", err)
	}
	return id, nil
}

func (r *tireRepoImpl) CreateForVehicle(ctx context.Context, vehicleID int64, p CreateVehicleTireParams) (int64, error) {
	const q = `
		INSERT INTO tires (
			place_id,
			vehicle_id,
			tire,
			mileage,
			max_usage,
			installed_at
		) VALUES ($1, $2, $3, $4, $5, $6::date)
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.PlaceID, vehicleID, p.Tire, p.Mileage, p.MaxUsage, p.InstalledAt).Scan(&id); err != nil {
		return 0, fmt.Errorf("create vehicle tire: %w", err)
	}
	return id, nil
}
