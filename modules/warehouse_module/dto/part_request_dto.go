package dto

import "time"

type PartRequestCreateRequest struct {
	PartID          int64  `json:"part_id" binding:"required" example:"1"`
	Quantity        int64  `json:"quantity" binding:"required" example:"5"`
	MechanicComment string `json:"mechanic_comment" binding:"required" example:"Need new brake pads for scheduled maintenance"`
}

type PartRequestUpdateRequest struct {
	PartID          int64  `json:"part_id" binding:"required" example:"1"`
	Quantity        int64  `json:"quantity" binding:"required" example:"7"`
	MechanicComment string `json:"mechanic_comment" binding:"required" example:"Updated quantity after inspection"`
	StatusID        int64  `json:"status_id" binding:"required" example:"1"`
}

type PartRequestStatusUpdateRequest struct {
	StatusID int64 `json:"status_id" binding:"required" example:"3"`
}

type PartRequestListQuery struct {
	PartID       int64
	StatusID     int64
	StatusCode   string
	AuthorUserID int64
	Limit        int
	Offset       int
	SortBy       string
	Order        string
}

type PartRequestStatusResponse struct {
	ID   int64  `json:"id" example:"1"`
	Code string `json:"code" example:"new"`
	Name string `json:"name" example:"Новая"`
}

type PartRequestPartBriefResponse struct {
	ID            int64  `json:"id" example:"1"`
	CatalogPartID string `json:"catalog_part_id" example:"BRK-PAD-001"`
	Name          string `json:"name" example:"Brake Pad Front"`
	Category      string `json:"category" example:"brake_system"`
}

type PartRequestResponse struct {
	ID              int64                        `json:"id" example:"1"`
	PartID          int64                        `json:"part_id" example:"1"`
	Part            PartRequestPartBriefResponse `json:"part"`
	Quantity        int64                        `json:"quantity" example:"5"`
	MechanicComment string                       `json:"mechanic_comment" example:"Need new brake pads for scheduled maintenance"`
	StatusID        int64                        `json:"status_id" example:"1"`
	Status          PartRequestStatusResponse    `json:"status"`
	AuthorUserID    int64                        `json:"author_user_id" example:"10"`
	AuthorEmail     *string                      `json:"author_email,omitempty" example:"mechanic@example.com"`
	AuthorFullName  *string                      `json:"author_full_name,omitempty" example:"Ivan Ivanov"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
}

type PartRequestListResponse struct {
	Items  []PartRequestResponse `json:"items"`
	Total  int64                 `json:"total" example:"1"`
	Limit  int                   `json:"limit" example:"50"`
	Offset int                   `json:"offset" example:"0"`
}
