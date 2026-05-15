package models

import "time"

type DriverShift struct {
	ID               int64
	DriverID         int64
	DriverIIN        string
	DriverName       string
	DriverSurname    string
	DriverMiddlename string
	DriverPhone      string
	DriverMail       string
	DriverStatusID   int64
	DriverStatusCode string
	DriverStatusName string
	ShiftDate        time.Time
	TimeFrom         string
	TimeTo           *string
	Comment          *string
	IsActive         bool
	TripsheetsCount  int64
	Tripsheets       []DriverShiftTripsheet
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DriverShiftTripsheet struct {
	ID                 int64
	TripsheetNumber    string
	TripsheetDate      time.Time
	VehicleID          *int64
	VehicleBrand       *string
	VehiclePlateNumber string
	StartTime          *time.Time
	EndTime            *time.Time
	StatusID           int64
	StatusName         *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
