package dto

import "time"

// VehicleCreateRequest — входной DTO для создания автомобиля
type VehicleCreateRequest struct {
	BoardNumber             string `json:"board_number" binding:"required"`
	TechnicalPassportNumber string `json:"technical_passport_number" binding:"required"`
	StateNumber             string `json:"state_number" binding:"required"`
	VIN                     string `json:"vin" binding:"required"`

	BrandModel      string    `json:"brand_model" binding:"required"`
	ManufactureYear int       `json:"manufacture_year" binding:"required"`
	ReceivedDate    time.Time `json:"received_date" binding:"required"` // ожидаем RFC3339 или "2006-01-02" (см. handler)

	EmptyWeightKG  *float64 `json:"empty_weight_kg,omitempty"`
	MaxWeightKG    *float64 `json:"max_weight_kg,omitempty"`
	EngineVolumeCC *int     `json:"engine_volume_cc,omitempty"`

	InsurancePolicyNumber *string    `json:"insurance_policy_number,omitempty"`
	InsuranceExpiryDate   *time.Time `json:"insurance_expiry_date,omitempty"`

	Mileage     int64   `json:"mileage"`
	CurrentFuel float64 `json:"current_fuel"`

	DriversIDs []int64 `json:"drivers_ids"`
}

type VehicleCreateResponse struct {
	ID int64 `json:"id"`
}
