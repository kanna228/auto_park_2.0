package repository

import (
	"context"
	"fmt"

	"github.com/lib/pq"
)

// UpdateVehicleParams — параметры для update
type UpdateVehicleParams struct {
	BoardNumber             string
	TechnicalPassportNumber string
	StateNumber             string
	VIN                     string

	BrandModel      string
	ManufactureYear int
	ReceivedDate    string // "YYYY-MM-DD"

	EmptyWeightKG  *float64
	MaxWeightKG    *float64
	EngineVolumeCC *int

	InsurancePolicyNumber *string
	InsuranceExpiryDate   *string // "YYYY-MM-DD"

	Mileage     int64
	CurrentFuel float64

	DriversIDs []int64
}

// VehicleRepoUpdateByID — реализация метода общего VehicleRepository
func (r *vehicleRepo) UpdateByID(ctx context.Context, id int64, p UpdateVehicleParams) (bool, error) {
	const q = `
		UPDATE vehicles
		SET
			board_number = $1,
			technical_passport_number = $2,
			state_number = $3,
			vin = $4,
			brand_model = $5,
			manufacture_year = $6,
			received_date = $7,
			empty_weight_kg = $8,
			max_weight_kg = $9,
			engine_volume_cc = $10,
			insurance_policy_number = $11,
			insurance_expiry_date = $12,
			mileage = $13,
			current_fuel = $14,
			drivers_ids = $15,
			updated_at = NOW()
		WHERE id = $16;
	`

	res, err := r.db.ExecContext(
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
		pq.Int64Array(p.DriversIDs),
		id,
	)
	if err != nil {
		return false, fmt.Errorf("update vehicle by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vehicle by id rows affected: %w", err)
	}
	return aff > 0, nil
}
