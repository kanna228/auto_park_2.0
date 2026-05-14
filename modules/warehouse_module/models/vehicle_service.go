package models

import "time"

type PartsCollection struct {
	ID          int64
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ServiceType struct {
	ID          int64
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type VehicleService struct {
	ID                 int64
	TypeID             int64
	TypeName           string
	TypeDescription    *string
	PartID             int64
	PartName           string
	PartDescription    *string
	VehicleID          int64
	VehicleStateNumber string
	VehicleBrandModel  string
	ServiceDate        time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
