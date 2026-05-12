package dto

import "time"

type VehiclePartInstallationCreateRequest struct {
	PartID               int64  `json:"part_id" binding:"required"`
	VehicleID            int64  `json:"vehicle_id" binding:"required"`
	MechanicShiftID      int64  `json:"mechanic_shift_id" binding:"required"`
	InstalledAt          string `json:"installed_at" binding:"required" example:"2026-05-12"`
	PlannedReplacementAt string `json:"planned_replacement_at" binding:"required" example:"2026-08-12"`
	Quantity             int64  `json:"quantity" binding:"required"`
}

type VehiclePartInstallationUpdateRequest struct {
	PartID               int64  `json:"part_id" binding:"required"`
	VehicleID            int64  `json:"vehicle_id" binding:"required"`
	MechanicShiftID      int64  `json:"mechanic_shift_id" binding:"required"`
	InstalledAt          string `json:"installed_at" binding:"required" example:"2026-05-12"`
	PlannedReplacementAt string `json:"planned_replacement_at" binding:"required" example:"2026-08-12"`
	Quantity             int64  `json:"quantity" binding:"required"`
	InstalledByUserID    *int64 `json:"installed_by_user_id,omitempty"`
	IsActive             *bool  `json:"is_active,omitempty"`
}

type VehiclePartInstallationActivityUpdateRequest struct {
	IsActive bool `json:"is_active"`
}

type VehiclePartInstallationListQuery struct {
	PartID            int64
	VehicleID         int64
	MechanicShiftID   int64
	InstalledByUserID int64
	IsActive          *bool
	DateFrom          string
	DateTo            string
	ReplacementFrom   string
	ReplacementTo     string
	Limit             int
	Offset            int
	SortBy            string
	Order             string
}

type VehiclePartInstallationPartBriefResponse struct {
	ID            int64  `json:"id"`
	CatalogPartID string `json:"catalog_part_id"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	IsConsumable  bool   `json:"is_consumable"`
}

type VehiclePartInstallationVehicleBriefResponse struct {
	ID          int64  `json:"id"`
	StateNumber string `json:"state_number"`
	BrandModel  string `json:"brand_model"`
}

type VehiclePartInstallationMechanicShiftBriefResponse struct {
	ID               int64   `json:"id"`
	UserID           int64   `json:"user_id"`
	ShiftDate        string  `json:"shift_date"`
	TimeFrom         string  `json:"time_from"`
	TimeTo           *string `json:"time_to,omitempty"`
	MechanicEmail    *string `json:"mechanic_email,omitempty"`
	MechanicFullName *string `json:"mechanic_full_name,omitempty"`
}

type VehiclePartInstallationResponse struct {
	ID                   int64                                              `json:"id"`
	PartID               int64                                              `json:"part_id"`
	Part                 VehiclePartInstallationPartBriefResponse           `json:"part"`
	VehicleID            int64                                              `json:"vehicle_id"`
	Vehicle              VehiclePartInstallationVehicleBriefResponse        `json:"vehicle"`
	MechanicShiftID      *int64                                             `json:"mechanic_shift_id"`
	MechanicShift        *VehiclePartInstallationMechanicShiftBriefResponse `json:"mechanic_shift,omitempty"`
	InstalledAt          string                                             `json:"installed_at"`
	PlannedReplacementAt string                                             `json:"planned_replacement_at"`
	Quantity             int64                                              `json:"quantity"`
	InstalledByUserID    int64                                              `json:"installed_by_user_id"`
	InstallerEmail       *string                                            `json:"installer_email,omitempty"`
	InstallerFullName    *string                                            `json:"installer_full_name,omitempty"`
	IsActive             bool                                               `json:"is_active"`
	CreatedAt            time.Time                                          `json:"created_at"`
	UpdatedAt            time.Time                                          `json:"updated_at"`
}

type VehiclePartInstallationListResponse struct {
	Items  []VehiclePartInstallationResponse `json:"items"`
	Total  int64                             `json:"total"`
	Limit  int                               `json:"limit"`
	Offset int                               `json:"offset"`
}
