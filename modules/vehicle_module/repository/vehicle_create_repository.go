package repository

import (
	"auto_park/modules/vehicle_module/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type VehicleRepository interface {
	Create(ctx context.Context, v CreateVehicleParams) (int64, error)
	GetByID(ctx context.Context, id int64, includeArchived ...bool) (*models.Vehicle, error)
	List(ctx context.Context, q ListVehiclesParams) ([]models.Vehicle, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateVehicleParams) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	UpdatePhotoPath(ctx context.Context, id int64, photoPath string) (bool, error)
	UnassignTiresByVehicleID(ctx context.Context, vehicleID int64) error
	ListVehicleStatuses(ctx context.Context) ([]models.VehicleStatus, error)
	ListIncidentsByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleIncidentHistoryRow, error)
	ListInstalledPartsByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleInstalledPartHistoryRow, error)
	ListTripsheetsByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleTripsheetHistoryRow, error)
	ListVehicleServicesByVehicleID(ctx context.Context, vehicleID int64) ([]VehicleServiceHistoryRow, error)
	GetVehicleStatusHistoryByID(ctx context.Context, id int64) (*models.VehicleStatusHistory, error)
	ListVehicleStatusHistory(ctx context.Context, p ListVehicleStatusHistoryParams) ([]models.VehicleStatusHistory, int64, error)
	ListDriversByIDs(ctx context.Context, ids []int64) ([]VehicleDriverRow, error)
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
	ReceivedDate    string // YYYY-MM-DD

	EmptyWeightKG  *float64
	MaxWeightKG    *float64
	EngineVolumeCC *int

	InsurancePolicyNumber *string
	InsuranceExpiryDate   *string // YYYY-MM-DD

	Mileage     int64
	CurrentFuel float64

	StatusID int64

	DriversIDs []int64
}

func (r *vehicleRepo) Create(ctx context.Context, p CreateVehicleParams) (int64, error) {
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
			status_id,
			drivers_ids
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		q,
		p.BoardNumber,
		p.TechnicalPassportNumber,
		p.StateNumber,
		p.VIN,
		p.BrandModel,
		p.ManufactureYear,
		p.ReceivedDate,
		p.EmptyWeightKG,
		p.MaxWeightKG,
		p.EngineVolumeCC,
		p.InsurancePolicyNumber,
		p.InsuranceExpiryDate,
		p.Mileage,
		p.CurrentFuel,
		p.StatusID,
		pq.Int64Array(p.DriversIDs),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create vehicle: %w", err)
	}

	return id, nil
}
