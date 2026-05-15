package models

import "time"

type Tripsheet struct {
	ID                 int64
	TripsheetNumber    string
	TripsheetDate      time.Time
	VehicleID          *int64
	VehicleBrand       *string
	VehiclePlateNumber string

	DriverLastName   *string
	DriverFirstName  *string
	DriverMiddleName *string
	DriverID         *int64
	DriverShiftID    *int64

	StartTime *time.Time
	EndTime   *time.Time

	MileageStart int
	MileageEnd   int

	FuelStart                  int
	FuelIssued                 int
	FuelConsumptionTheoretical int
	FuelConsumptionActual      int

	StatusID int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateTripsheetInput struct {
	TripsheetNumber    string
	TripsheetDate      time.Time
	VehicleID          *int64
	VehicleBrand       *string
	VehiclePlateNumber string

	DriverLastName   *string
	DriverFirstName  *string
	DriverMiddleName *string
	DriverID         *int64
	DriverShiftID    *int64

	StartTime *time.Time
	EndTime   *time.Time

	MileageStart int
	MileageEnd   int

	FuelStart                  int
	FuelIssued                 int
	FuelConsumptionTheoretical int
	FuelConsumptionActual      int

	StatusID *int64
}

type UpdateTripsheetInput struct {
	ID                 int64
	TripsheetNumber    string
	TripsheetDate      time.Time
	VehicleID          *int64
	VehicleBrand       *string
	VehiclePlateNumber string

	DriverLastName   *string
	DriverFirstName  *string
	DriverMiddleName *string
	DriverID         *int64
	DriverShiftID    *int64

	StartTime *time.Time
	EndTime   *time.Time

	MileageStart int
	MileageEnd   int

	FuelStart                  int
	FuelIssued                 int
	FuelConsumptionTheoretical int
	FuelConsumptionActual      int

	StatusID *int64
}
