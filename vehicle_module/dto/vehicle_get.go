package dto

import "time"

type VehicleResponse struct {
	ID int64 `json:"id"`

	BoardNumber             string `json:"board_number"`
	TechnicalPassportNumber string `json:"technical_passport_number"`
	StateNumber             string `json:"state_number"`
	VIN                     string `json:"vin"`

	BrandModel      string    `json:"brand_model"`
	ManufactureYear int       `json:"manufacture_year"`
	ReceivedDate    time.Time `json:"received_date"`

	EmptyWeightKG  *float64 `json:"empty_weight_kg,omitempty"`
	MaxWeightKG    *float64 `json:"max_weight_kg,omitempty"`
	EngineVolumeCC *int     `json:"engine_volume_cc,omitempty"`

	InsurancePolicyNumber *string    `json:"insurance_policy_number,omitempty"`
	InsuranceExpiryDate   *time.Time `json:"insurance_expiry_date,omitempty"`

	Mileage     int64   `json:"mileage"`
	CurrentFuel float64 `json:"current_fuel"`

	DriversIDs []int64 `json:"drivers_ids"`

	PhotoPath string `json:"photo_path,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
