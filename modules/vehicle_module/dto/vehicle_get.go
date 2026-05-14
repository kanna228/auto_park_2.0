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

type VehicleInstalledPartHistoryItem struct {
	ID                   int64      `json:"id"`
	PartID               int64      `json:"part_id"`
	CatalogPartID        string     `json:"catalog_part_id"`
	PartName             string     `json:"part_name"`
	PartCategory         string     `json:"part_category"`
	IsConsumable         bool       `json:"is_consumable"`
	VehicleID            int64      `json:"vehicle_id"`
	InstalledAt          time.Time  `json:"installed_at"`
	PlannedReplacementAt *time.Time `json:"planned_replacement_at,omitempty"`
	Quantity             int64      `json:"quantity"`
	InstalledByUserID    int64      `json:"installed_by_user_id"`
	InstallerEmail       *string    `json:"installer_email,omitempty"`
	InstallerFullName    *string    `json:"installer_full_name,omitempty"`
	IsActive             bool       `json:"is_active"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type VehicleStatusHistoryItem struct {
	ID             int64                  `json:"id"`
	VehicleID      int64                  `json:"vehicle_id"`
	StatusID       int64                  `json:"status_id"`
	Status         VehicleStatusBriefInfo `json:"status"`
	StartDate      time.Time              `json:"start_date"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	EndDateDisplay string                 `json:"end_date_display"`
	IsCurrent      bool                   `json:"is_current"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type VehicleStatusBriefInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type VehicleTripsheetStatusBriefInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type VehicleTripsheetDriverInfo struct {
	ID         *int64  `json:"id,omitempty"`
	IIN        *string `json:"iin,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	MiddleName *string `json:"middle_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty"`
}

type VehicleTripsheetTripHistoryItem struct {
	ID               int64                           `json:"id"`
	TripsheetID      int64                           `json:"tripsheet_id"`
	RouteDescription string                          `json:"route_description"`
	StartTime        *time.Time                      `json:"start_time,omitempty"`
	EndTime          *time.Time                      `json:"end_time,omitempty"`
	DistancePassed   int                             `json:"distance_passed"`
	StatusID         int64                           `json:"status_id"`
	Status           VehicleTripsheetStatusBriefInfo `json:"status"`
	CreatedAt        time.Time                       `json:"created_at"`
	UpdatedAt        time.Time                       `json:"updated_at"`
}

type VehicleTripsheetHistoryItem struct {
	ID                         int64                             `json:"id"`
	TripsheetNumber            string                            `json:"tripsheet_number"`
	TripsheetDate              string                            `json:"tripsheet_date"`
	VehicleID                  *int64                            `json:"vehicle_id,omitempty"`
	VehicleBrand               *string                           `json:"vehicle_brand,omitempty"`
	VehiclePlateNumber         string                            `json:"vehicle_plate_number"`
	DriverID                   *int64                            `json:"driver_id,omitempty"`
	Driver                     *VehicleTripsheetDriverInfo       `json:"driver,omitempty"`
	DriverLastName             *string                           `json:"driver_last_name,omitempty"`
	DriverFirstName            *string                           `json:"driver_first_name,omitempty"`
	DriverMiddleName           *string                           `json:"driver_middle_name,omitempty"`
	StartTime                  *time.Time                        `json:"start_time,omitempty"`
	EndTime                    *time.Time                        `json:"end_time,omitempty"`
	MileageStart               int                               `json:"mileage_start"`
	MileageEnd                 int                               `json:"mileage_end"`
	FuelStart                  int                               `json:"fuel_start"`
	FuelIssued                 int                               `json:"fuel_issued"`
	FuelConsumptionTheoretical int                               `json:"fuel_consumption_theoretical"`
	FuelConsumptionActual      int                               `json:"fuel_consumption_actual"`
	StatusID                   int64                             `json:"status_id"`
	Status                     VehicleTripsheetStatusBriefInfo   `json:"status"`
	Trips                      []VehicleTripsheetTripHistoryItem `json:"trips"`
	CreatedAt                  time.Time                         `json:"created_at"`
	UpdatedAt                  time.Time                         `json:"updated_at"`
}

type VehicleServicePartInfo struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type VehicleServiceTypeInfo struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type VehicleServiceHistoryItem struct {
	ID        int64                  `json:"id"`
	TypeID    int64                  `json:"type_id"`
	Type      VehicleServiceTypeInfo `json:"type"`
	PartID    int64                  `json:"part_id"`
	Part      VehicleServicePartInfo `json:"part"`
	VehicleID int64                  `json:"vehicle_id"`
	Date      time.Time              `json:"date"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
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
	InstalledParts       []VehicleInstalledPartHistoryItem       `json:"installed_parts"`
	StatusHistory        []VehicleStatusHistoryItem              `json:"status_history"`
	Tripsheets           []VehicleTripsheetHistoryItem           `json:"tripsheets"`
	Services             []VehicleServiceHistoryItem             `json:"services"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleStatusResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type VehicleStatusListResponse struct {
	Items  []VehicleStatusResponse `json:"items"`
	Total  int64                   `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
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
