package models

import "time"

type IncidentTripsheet struct {
	ID                         int64
	TripsheetNumber            string
	TripsheetDate              time.Time
	VehicleBrand               *string
	VehiclePlateNumber         string
	DriverID                   *int64
	DriverLastName             *string
	DriverFirstName            *string
	DriverMiddleName           *string
	StatusID                   int64
	StatusName                 string
	StartTime                  *time.Time
	EndTime                    *time.Time
	MileageStart               int
	MileageEnd                 int
	FuelStart                  int
	FuelIssued                 int
	FuelConsumptionTheoretical int
	FuelConsumptionActual      int
}

type Incident struct {
	ID                 int64
	IncidentTypeID     int64
	IncidentTypeName   string
	VehicleID          int64
	VehicleStateNumber string
	DriverID           int64
	DriverFullName     string
	MechanicID         int64
	MechanicFullName   string
	TripsheetID        *int64
	Tripsheet          *IncidentTripsheet
	IncidentDate       time.Time
	IncidentTime       time.Time
	Location           string
	Description        string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateIncidentInput struct {
	IncidentTypeID int64
	VehicleID      int64
	DriverID       int64
	MechanicID     int64
	TripsheetID    *int64
	IncidentDate   time.Time
	IncidentTime   time.Time
	Location       string
	Description    string
}

type UpdateIncidentInput struct {
	ID             int64
	IncidentTypeID int64
	VehicleID      int64
	DriverID       int64
	MechanicID     int64
	TripsheetID    *int64
	IncidentDate   time.Time
	IncidentTime   time.Time
	Location       string
	Description    string
}

type IncidentType struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
