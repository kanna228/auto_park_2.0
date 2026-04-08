package repository

import (
	"auto_park/vehicle_module/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type VehicleRepository interface {
	Create(ctx context.Context, v CreateVehicleParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Vehicle, error)
	List(ctx context.Context, q ListVehiclesParams) ([]models.Vehicle, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateVehicleParams) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	UpdatePhotoPath(ctx context.Context, id int64, photoPath string) (bool, error)
}

type vehicleRepo struct {
	db *sql.DB
}

func NewVehicleRepository(db *sql.DB) VehicleRepository {
	return &vehicleRepo{db: db}
}

type CreateVehicleParams struct {
	BoardNumber             string
	TechnicalPassportNumber string
	StateNumber             string
	VIN                     string

	BrandModel      string
	ManufactureYear int
	ReceivedDate    string

	EmptyWeightKG  *float64
	MaxWeightKG    *float64
	EngineVolumeCC *int

	InsurancePolicyNumber *string
	InsuranceExpiryDate   *string

	Mileage     int64
	CurrentFuel float64

	DriversIDs []int64
}

func (r *vehicleRepo) Create(ctx context.Context, v CreateVehicleParams) (int64, error) {
	const q = `
		INSERT INTO vehicles (
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
			drivers_ids
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		q,
		v.BoardNumber,
		v.TechnicalPassportNumber,
		v.StateNumber,
		v.VIN,
		v.BrandModel,
		v.ManufactureYear,
		v.ReceivedDate,
		v.EmptyWeightKG,
		v.MaxWeightKG,
		v.EngineVolumeCC,
		v.InsurancePolicyNumber,
		v.InsuranceExpiryDate,
		v.Mileage,
		v.CurrentFuel,
		pq.Int64Array(v.DriversIDs),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("create vehicle: %w", err)
	}

	return id, nil
}
