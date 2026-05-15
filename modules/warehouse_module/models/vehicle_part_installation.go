package models

import "time"

type VehiclePartInstallation struct {
	ID                        int64
	PartID                    int64
	PartCatalogCode           string
	PartName                  string
	PartCategory              string
	PartIsConsumable          bool
	VehicleID                 int64
	VehicleStateNumber        string
	VehicleBrandModel         string
	MechanicShiftID           *int64
	MechanicShiftUserID       *int64
	MechanicShiftDate         *time.Time
	MechanicShiftTimeFrom     *time.Time
	MechanicShiftTimeTo       *time.Time
	MechanicShiftUserEmail    *string
	MechanicShiftUserFullName *string
	InstalledAt               time.Time
	PlannedReplacementAt      time.Time
	Quantity                  int64
	UnitPrice                 float64
	TotalPrice                float64
	InstalledByUserID         int64
	InstallerEmail            *string
	InstallerFullName         *string
	IsActive                  bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}
