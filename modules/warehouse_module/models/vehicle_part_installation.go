package models

import "time"

type VehiclePartInstallation struct {
	ID                   int64
	PartID               int64
	PartCatalogCode      string
	PartName             string
	PartCategory         string
	PartIsConsumable     bool
	VehicleID            int64
	VehicleStateNumber   string
	VehicleBrandModel    string
	InstalledAt          time.Time
	PlannedReplacementAt time.Time
	Quantity             int64
	InstalledByUserID    int64
	InstallerEmail       *string
	InstallerFullName    *string
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
