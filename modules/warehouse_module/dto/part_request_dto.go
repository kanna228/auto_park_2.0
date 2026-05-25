package dto

import "time"

type PartRequestCreateRequest struct {
	PartID               int64  `json:"part_id" binding:"gt=0" example:"1"`
	Quantity             int64  `json:"quantity" binding:"gt=0" example:"5"`
	MechanicComment      string `json:"mechanic_comment" binding:"required" example:"Need new brake pads for scheduled maintenance"`
	VehicleID            *int64 `json:"vehicle_id,omitempty" example:"12"`
	MechanicShiftID      *int64 `json:"mechanic_shift_id,omitempty" example:"3"`
	PlannedReplacementAt string `json:"planned_replacement_at,omitempty" example:"2026-11-25"`
}

type PartRequestUpdateRequest struct {
	PartID               int64  `json:"part_id" binding:"gt=0" example:"1"`
	Quantity             int64  `json:"quantity" binding:"gt=0" example:"7"`
	MechanicComment      string `json:"mechanic_comment" binding:"required" example:"Updated quantity after inspection"`
	StatusID             int64  `json:"status_id" binding:"gt=0" example:"1"`
	RejectionComment     string `json:"rejection_comment,omitempty" example:"Недостаточно данных для заказа детали"`
	HistoryComment       string `json:"history_comment,omitempty" example:"Updated request after additional vehicle inspection"`
	VehicleID            *int64 `json:"vehicle_id,omitempty" example:"12"`
	MechanicShiftID      *int64 `json:"mechanic_shift_id,omitempty" example:"3"`
	PlannedReplacementAt string `json:"planned_replacement_at,omitempty" example:"2026-11-25"`
}

type PartRequestStatusUpdateRequest struct {
	StatusID         int64  `json:"status_id" binding:"gt=0" example:"3"`
	Comment          string `json:"comment,omitempty" example:"Approved by warehouse manager"`
	RejectionComment string `json:"rejection_comment,omitempty" example:"Недостаточно данных для заказа детали"`
}

type PartRequestRepairStatusUpdateRequest struct {
	Status               string `json:"status" binding:"required" example:"completed"`
	VehicleID            *int64 `json:"vehicle_id,omitempty" example:"12"`
	MechanicShiftID      *int64 `json:"mechanic_shift_id,omitempty" example:"3"`
	InstalledAt          string `json:"installed_at,omitempty" example:"2026-05-25"`
	PlannedReplacementAt string `json:"planned_replacement_at,omitempty" example:"2026-11-25"`
	Comment              string `json:"comment,omitempty" example:"Repair completed and part installed"`
}

type PartRequestListQuery struct {
	PartID       int64
	StatusID     int64
	StatusCode   string
	AuthorUserID int64
	DateFrom     string
	DateTo       string
	Limit        int
	Offset       int
	SortBy       string
	Order        string
}

type PartRequestHistoryListQuery struct {
	PartRequestID   int64
	StatusID        int64
	StatusCode      string
	ChangedByUserID int64
	DateFrom        string
	DateTo          string
	Limit           int
	Offset          int
	SortBy          string
	Order           string
}

type PartRequestStatusResponse struct {
	ID   int64  `json:"id" example:"1"`
	Code string `json:"code" example:"new"`
	Name string `json:"name" example:"Новая"`
}

type PartRequestStatusListResponse struct {
	Items  []PartRequestStatusResponse `json:"items"`
	Total  int64                       `json:"total" example:"3"`
	Limit  int                         `json:"limit" example:"50"`
	Offset int                         `json:"offset" example:"0"`
}

type PartRequestPartBriefResponse struct {
	ID                int64  `json:"id" example:"1"`
	CatalogPartID     string `json:"catalog_part_id" example:"BRK-PAD-001"`
	Name              string `json:"name" example:"Brake Pad Front"`
	Category          string `json:"category" example:"brake_system"`
	AvailableQuantity int64  `json:"available_quantity" example:"12"`
}

type PartRequestHistoryResponse struct {
	ID                int64                     `json:"id" example:"1"`
	PartRequestID     int64                     `json:"part_request_id" example:"15"`
	StatusID          int64                     `json:"status_id" example:"3"`
	Status            PartRequestStatusResponse `json:"status"`
	ChangedByUserID   int64                     `json:"changed_by_user_id" example:"5"`
	ChangedByEmail    *string                   `json:"changed_by_email,omitempty" example:"manager@example.com"`
	ChangedByFullName *string                   `json:"changed_by_full_name,omitempty" example:"Ivan Ivanov"`
	Comment           string                    `json:"comment" example:"Approved by warehouse manager"`
	ChangedAt         time.Time                 `json:"changed_at"`
}

type PartRequestResponse struct {
	ID                        int64                        `json:"id" example:"1"`
	PartID                    int64                        `json:"part_id" example:"1"`
	Part                      PartRequestPartBriefResponse `json:"part"`
	Quantity                  int64                        `json:"quantity" example:"5"`
	MechanicComment           string                       `json:"mechanic_comment" example:"Need new brake pads for scheduled maintenance"`
	RejectionComment          *string                      `json:"rejection_comment,omitempty" example:"Недостаточно данных для заказа детали"`
	VehicleID                 *int64                       `json:"vehicle_id,omitempty" example:"12"`
	MechanicShiftID           *int64                       `json:"mechanic_shift_id,omitempty" example:"3"`
	PlannedReplacementAt      *time.Time                   `json:"planned_replacement_at,omitempty"`
	RepairStatus              string                       `json:"repair_status" example:"in_progress"`
	CompletedAt               *time.Time                   `json:"completed_at,omitempty"`
	CompletedByUserID         *int64                       `json:"completed_by_user_id,omitempty" example:"5"`
	VehiclePartInstallationID *int64                       `json:"vehicle_part_installation_id,omitempty" example:"30"`
	StatusID                  int64                        `json:"status_id" example:"1"`
	Status                    PartRequestStatusResponse    `json:"status"`
	AuthorUserID              int64                        `json:"author_user_id" example:"10"`
	AuthorEmail               *string                      `json:"author_email,omitempty" example:"mechanic@example.com"`
	AuthorFullName            *string                      `json:"author_full_name,omitempty" example:"Ivan Ivanov"`
	History                   []PartRequestHistoryResponse `json:"history"`
	CreatedAt                 time.Time                    `json:"created_at"`
	UpdatedAt                 time.Time                    `json:"updated_at"`
}

type PartRequestListResponse struct {
	Items  []PartRequestResponse `json:"items"`
	Total  int64                 `json:"total" example:"1"`
	Limit  int                   `json:"limit" example:"50"`
	Offset int                   `json:"offset" example:"0"`
}

type PartRequestHistoryListResponse struct {
	Items  []PartRequestHistoryResponse `json:"items"`
	Total  int64                        `json:"total" example:"1"`
	Limit  int                          `json:"limit" example:"50"`
	Offset int                          `json:"offset" example:"0"`
}
