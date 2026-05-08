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
	HistoryComment  string `json:"history_comment,omitempty" example:"Updated request after additional vehicle inspection"`
}

type PartRequestStatusUpdateRequest struct {
	StatusID int64  `json:"status_id" binding:"required" example:"3"`
	Comment  string `json:"comment,omitempty" example:"Approved by warehouse manager"`
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

type PartRequestPartBriefResponse struct {
	ID            int64  `json:"id" example:"1"`
	CatalogPartID string `json:"catalog_part_id" example:"BRK-PAD-001"`
	Name          string `json:"name" example:"Brake Pad Front"`
	Category      string `json:"category" example:"brake_system"`
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
	History         []PartRequestHistoryResponse `json:"history"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
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
