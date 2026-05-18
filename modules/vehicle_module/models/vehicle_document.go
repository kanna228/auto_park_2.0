package models

import "time"

type VehicleDocument struct {
	ID        int64
	VehicleID int64
	Type      string
	Number    string
	ValidFrom time.Time
	ValidTo   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
