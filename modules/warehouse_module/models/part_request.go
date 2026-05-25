package models

import "time"

type PartRequestStatus struct {
	ID   int64
	Code string
	Name string
}

type PartRequest struct {
	ID                        int64
	PartID                    int64
	PartCatalogCode           string
	PartName                  string
	PartCategory              string
	PartQuantity              int64
	Quantity                  int64
	MechanicComment           string
	RejectionComment          *string
	VehicleID                 *int64
	MechanicShiftID           *int64
	PlannedReplacementAt      *time.Time
	RepairStatus              string
	CompletedAt               *time.Time
	CompletedByUserID         *int64
	VehiclePartInstallationID *int64
	StatusID                  int64
	StatusCode                string
	StatusName                string
	AuthorUserID              int64
	AuthorEmail               *string
	AuthorFullName            *string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type PartRequestHistory struct {
	ID                int64
	PartRequestID     int64
	StatusID          int64
	StatusCode        string
	StatusName        string
	ChangedByUserID   int64
	ChangedByEmail    *string
	ChangedByFullName *string
	Comment           string
	ChangedAt         time.Time
}
