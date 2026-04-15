package models

import "time"

type FuelRefill struct {
	ID          int64
	TripsheetID int64
	VehicleID   int64
	FuelAmount  float64
	Date        time.Time
	Time        time.Time
	Location    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateFuelRefillInput struct {
	TripsheetID int64
	VehicleID   int64
	FuelAmount  float64
	Date        time.Time
	Time        time.Time
	Location    *string
}

type UpdateFuelRefillInput struct {
	ID          int64
	TripsheetID int64
	VehicleID   int64
	FuelAmount  float64
	Date        time.Time
	Time        time.Time
	Location    *string
}
