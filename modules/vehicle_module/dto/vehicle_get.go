package dto

import "time"

type VehicleInsuranceHistoryItem struct {
	ID        int64     `json:"id"`
	VehicleID int64     `json:"vehicle_id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	FilePath  string    `json:"file_path,omitempty"`
	FileURL   string    `json:"file_url,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleTechnicalInspectionHistoryItem struct {
	ID        int64     `json:"id"`
	VehicleID int64     `json:"vehicle_id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	FilePath  string    `json:"file_path,omitempty"`
	FileURL   string    `json:"file_url,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

	StatusID   int64  `json:"status_id"`
	StatusName string `json:"status_name"`

	DriversIDs []int64             `json:"drivers_ids"`
	Drivers    []VehicleDriverInfo `json:"drivers"`

	PhotoPath string `json:"photo_path,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`

	Insurances           []VehicleInsuranceHistoryItem           `json:"insurances"`
	TechnicalInspections []VehicleTechnicalInspectionHistoryItem `json:"technical_inspections"`
	Incidents            []VehicleIncidentHistoryItem            `json:"incidents"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleStatusResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type VehicleStatusListResponse struct {
	Items []VehicleStatusResponse `json:"items"`
	Total int64                   `json:"total"`
}
type VehicleDriverInfo struct {
	ID         int64   `json:"id"`
	IIN        string  `json:"iin"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty"`
}
