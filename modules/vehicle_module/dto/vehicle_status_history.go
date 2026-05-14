package dto

import "time"

type VehicleStatusHistoryListQuery struct {
	VehicleID int64
	StatusID  int64
	StartFrom string
	StartTo   string
	EndFrom   string
	EndTo     string
	IsCurrent *bool
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type VehicleStatusHistoryStatusBriefResponse struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"В использовании"`
}

type VehicleStatusHistoryVehicleBriefResponse struct {
	ID          int64  `json:"id" example:"12"`
	StateNumber string `json:"state_number" example:"123ABC01"`
	BrandModel  string `json:"brand_model" example:"Toyota Camry"`
}

type VehicleStatusHistoryResponse struct {
	ID             int64                                    `json:"id" example:"1"`
	VehicleID      int64                                    `json:"vehicle_id" example:"12"`
	Vehicle        VehicleStatusHistoryVehicleBriefResponse `json:"vehicle"`
	StatusID       int64                                    `json:"status_id" example:"2"`
	Status         VehicleStatusHistoryStatusBriefResponse  `json:"status"`
	StartDate      time.Time                                `json:"start_date"`
	EndDate        *time.Time                               `json:"end_date,omitempty"`
	EndDateDisplay string                                   `json:"end_date_display" example:"По настоящее время"`
	IsCurrent      bool                                     `json:"is_current" example:"true"`
	CreatedAt      time.Time                                `json:"created_at"`
	UpdatedAt      time.Time                                `json:"updated_at"`
}

type VehicleStatusHistoryListResponse struct {
	Items  []VehicleStatusHistoryResponse `json:"items"`
	Total  int64                          `json:"total" example:"1"`
	Limit  int                            `json:"limit" example:"50"`
	Offset int                            `json:"offset" example:"0"`
}
