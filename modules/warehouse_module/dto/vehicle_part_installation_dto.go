package dto

import "time"

type VehiclePartInstallationCreateRequest struct {
	PartID               int64  `json:"part_id" binding:"required" example:"1"`
	VehicleID            int64  `json:"vehicle_id" binding:"required" example:"12"`
	InstalledAt          string `json:"installed_at" binding:"required" example:"2026-05-08"`
	PlannedReplacementAt string `json:"planned_replacement_at" binding:"required" example:"2027-05-08"`
	Quantity             int64  `json:"quantity" binding:"required" example:"1"`
}

type VehiclePartInstallationUpdateRequest struct {
	PartID               int64  `json:"part_id" binding:"required" example:"1"`
	VehicleID            int64  `json:"vehicle_id" binding:"required" example:"12"`
	InstalledAt          string `json:"installed_at" binding:"required" example:"2026-05-08"`
	PlannedReplacementAt string `json:"planned_replacement_at" binding:"required" example:"2027-05-08"`
	Quantity             int64  `json:"quantity" binding:"required" example:"1"`
	InstalledByUserID    *int64 `json:"installed_by_user_id,omitempty" example:"5"`
	IsActive             *bool  `json:"is_active,omitempty" example:"true"`
}

type VehiclePartInstallationActivityUpdateRequest struct {
	IsActive bool `json:"is_active" example:"false"`
}

type VehiclePartInstallationListQuery struct {
	PartID            int64
	VehicleID         int64
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
	ID            int64  `json:"id" example:"1"`
	CatalogPartID string `json:"catalog_part_id" example:"BRK-PAD-001"`
	Name          string `json:"name" example:"Brake Pad Front"`
	Category      string `json:"category" example:"brake_system"`
	IsConsumable  bool   `json:"is_consumable" example:"false"`
}

type VehiclePartInstallationVehicleBriefResponse struct {
	ID          int64  `json:"id" example:"12"`
	StateNumber string `json:"state_number" example:"123ABC01"`
	BrandModel  string `json:"brand_model" example:"Toyota Camry"`
}

type VehiclePartInstallationResponse struct {
	ID                   int64                                       `json:"id" example:"1"`
	PartID               int64                                       `json:"part_id" example:"1"`
	Part                 VehiclePartInstallationPartBriefResponse    `json:"part"`
	VehicleID            int64                                       `json:"vehicle_id" example:"12"`
	Vehicle              VehiclePartInstallationVehicleBriefResponse `json:"vehicle"`
	InstalledAt          string                                      `json:"installed_at" example:"2026-05-08"`
	PlannedReplacementAt string                                      `json:"planned_replacement_at" example:"2027-05-08"`
	Quantity             int64                                       `json:"quantity" example:"1"`
	InstalledByUserID    int64                                       `json:"installed_by_user_id" example:"5"`
	InstallerEmail       *string                                     `json:"installer_email,omitempty" example:"mechanic@example.com"`
	InstallerFullName    *string                                     `json:"installer_full_name,omitempty" example:"Ivan Ivanov"`
	IsActive             bool                                        `json:"is_active" example:"true"`
	CreatedAt            time.Time                                   `json:"created_at"`
	UpdatedAt            time.Time                                   `json:"updated_at"`
}

type VehiclePartInstallationListResponse struct {
	Items  []VehiclePartInstallationResponse `json:"items"`
	Total  int64                             `json:"total" example:"1"`
	Limit  int                               `json:"limit" example:"50"`
	Offset int                               `json:"offset" example:"0"`
}
