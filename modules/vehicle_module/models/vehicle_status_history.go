package models

import "time"

type VehicleStatusHistory struct {
	ID                 int64
	VehicleID          int64
	VehicleStateNumber string
	VehicleBrandModel  string
	StatusID           int64
	StatusName         string
	StartDate          time.Time
	EndDate            *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
