package dto

import "time"

type TechnicalInspectionCreateRequest struct {
	VehicleID int64  `json:"vehicle_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	IsActive  *bool  `json:"is_active"`
}

type TechnicalInspectionUpdateRequest struct {
	VehicleID int64  `json:"vehicle_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	IsActive  *bool  `json:"is_active"`
}

type TechnicalInspectionCreateResponse struct {
	ID int64 `json:"id"`
}

type TechnicalInspectionResponse struct {
	ID        int64     `json:"id"`
	VehicleID int64     `json:"vehicle_id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	FilePath  string    `json:"file_path,omitempty"`
	FileURL   string    `json:"file_url,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TechnicalInspectionListQuery struct {
	VehicleID *int64
	IsActive  *bool
	Name      string
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type TechnicalInspectionListResponse struct {
	Items  []TechnicalInspectionResponse `json:"items"`
	Total  int64                         `json:"total"`
	Limit  int                           `json:"limit"`
	Offset int                           `json:"offset"`
}
