package dto

import "time"

type PurchaseRequestCreateRequest struct {
	PartID              int64   `json:"part_id" binding:"gt=0" example:"1"`
	Quantity            int64   `json:"quantity" binding:"gt=0" example:"5"`
	SourcePartRequestID *int64  `json:"source_part_request_id,omitempty" example:"10"`
	Comment             *string `json:"comment,omitempty" example:"Purchase missing brake pads for mechanic request"`
}

type PurchaseRequestListQuery struct {
	PartID              int64
	SourcePartRequestID int64
	Status              string
	Limit               int
	Offset              int
	SortBy              string
	Order               string
}

type PurchaseRequestPartBriefResponse struct {
	ID            int64  `json:"id" example:"1"`
	CatalogPartID string `json:"catalog_part_id" example:"BRK-PAD-001"`
	Name          string `json:"name" example:"Brake Pad Front"`
	Category      string `json:"category" example:"brake_system"`
}

type PurchaseRequestResponse struct {
	ID                  int64                            `json:"id" example:"1"`
	PartID              int64                            `json:"part_id" example:"1"`
	Part                PurchaseRequestPartBriefResponse `json:"part"`
	Quantity            int64                            `json:"quantity" example:"5"`
	Status              string                           `json:"status" example:"new"`
	SourcePartRequestID *int64                           `json:"source_part_request_id,omitempty" example:"10"`
	Comment             *string                          `json:"comment,omitempty"`
	CreatedByUserID     int64                            `json:"created_by_user_id" example:"6"`
	CreatedByEmail      *string                          `json:"created_by_email,omitempty"`
	CreatedByFullName   *string                          `json:"created_by_full_name,omitempty"`
	ConfirmedByUserID   *int64                           `json:"confirmed_by_user_id,omitempty" example:"6"`
	ConfirmedByEmail    *string                          `json:"confirmed_by_email,omitempty"`
	ConfirmedByFullName *string                          `json:"confirmed_by_full_name,omitempty"`
	ConfirmedAt         *time.Time                       `json:"confirmed_at,omitempty"`
	CreatedAt           time.Time                        `json:"created_at"`
	UpdatedAt           time.Time                        `json:"updated_at"`
}

type PurchaseRequestListResponse struct {
	Items  []PurchaseRequestResponse `json:"items"`
	Total  int64                     `json:"total" example:"1"`
	Limit  int                       `json:"limit" example:"50"`
	Offset int                       `json:"offset" example:"0"`
}
