package models

import "time"

type DriverStatus struct {
	ID          int64     `json:"id" example:"1"`
	Code        string    `json:"code" example:"available"`
	Name        string    `json:"name" example:"Доступен"`
	Description string    `json:"description,omitempty" example:"Водитель доступен и может быть назначен на рейс"`
	CreatedAt   time.Time `json:"created_at" example:"2026-05-14T06:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-05-14T06:00:00Z"`
}

type Driver struct {
	ID               int64                   `json:"id" example:"1"`
	IIN              string                  `json:"iin" example:"010203040506"`
	Name             string                  `json:"name" example:"Dias"`
	Surname          string                  `json:"surname" example:"Abdimanap"`
	Middlename       string                  `json:"middlename,omitempty" example:"Diasovich"`
	Phone            string                  `json:"phone,omitempty" example:"+77001234567"`
	Mail             string                  `json:"mail,omitempty" example:"dias@mail.com"`
	PhotoPath        string                  `json:"photo_path,omitempty" example:"drivers/driver_1_1710000000.jpg"`
	BirthDate        *time.Time              `json:"birth_date,omitempty" example:"1990-11-11T00:00:00Z"`
	LicenseNumber    string                  `json:"license_number,omitempty" example:"DL-123456"`
	LicenseCategory  string                  `json:"license_category,omitempty" example:"B, C"`
	ExperienceYears  *int                    `json:"experience_years,omitempty" example:"5"`
	StatusID         int64                   `json:"status_id" example:"1"`
	Status           DriverStatus            `json:"status"`
	StatusText       string                  `json:"status_text,omitempty"`
	AssignedVehicles []DriverAssignedVehicle `json:"assigned_vehicles,omitempty"`
	CreatedAt        time.Time               `json:"created_at" example:"2026-02-18T12:00:00Z"`
	UpdatedAt        time.Time               `json:"updated_at" example:"2026-02-18T12:10:00Z"`
}

type DriverAssignedVehicle struct {
	ID          int64  `json:"id" example:"1"`
	BoardNumber string `json:"board_number" example:"55"`
	StateNumber string `json:"state_number" example:"777ABC01"`
	BrandModel  string `json:"brand_model" example:"Toyota Camry"`
	StatusID    int64  `json:"status_id" example:"1"`
	StatusName  string `json:"status_name" example:"В использовании"`
}

type DriverPassportTripsheetItem struct {
	ID                 int64      `json:"id"`
	TripsheetNumber    string     `json:"tripsheet_number"`
	TripsheetDate      string     `json:"tripsheet_date"`
	VehicleID          *int64     `json:"vehicle_id,omitempty"`
	VehicleBrand       *string    `json:"vehicle_brand,omitempty"`
	VehiclePlateNumber string     `json:"vehicle_plate_number"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	WorkedHours        float64    `json:"worked_hours"`
	TripsCount         int64      `json:"trips_count"`
	StatusID           int64      `json:"status_id"`
	StatusName         *string    `json:"status_name,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type DriverPassportIncidentItem struct {
	ID                 int64     `json:"id"`
	IncidentTypeID     int64     `json:"incident_type_id"`
	IncidentTypeName   string    `json:"incident_type_name"`
	VehicleID          int64     `json:"vehicle_id"`
	VehicleStateNumber string    `json:"vehicle_state_number"`
	TripsheetID        *int64    `json:"tripsheet_id,omitempty"`
	IncidentDate       string    `json:"incident_date"`
	IncidentTime       string    `json:"incident_time"`
	Location           string    `json:"location"`
	Description        string    `json:"description,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DriverPassport struct {
	Driver           Driver                        `json:"driver"`
	Status           string                        `json:"status"`
	AssignedVehicles []DriverAssignedVehicle       `json:"assigned_vehicles"`
	TotalWorkedHours float64                       `json:"total_worked_hours"`
	IncidentsCount   int64                         `json:"incidents_count"`
	TripsheetsTotal  int64                         `json:"tripsheets_total"`
	IncidentsTotal   int64                         `json:"incidents_total"`
	Tripsheets       []DriverPassportTripsheetItem `json:"tripsheets"`
	Incidents        []DriverPassportIncidentItem  `json:"incidents"`
}
