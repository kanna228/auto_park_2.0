package dto

import "time"

type CreateDriverShiftRequest struct {
	DriverID  int64   `json:"driver_id" binding:"required" example:"1"`
	ShiftDate string  `json:"shift_date" binding:"required" example:"2026-05-14"`
	TimeFrom  string  `json:"time_from" binding:"required" example:"08:00"`
	TimeTo    *string `json:"time_to,omitempty" example:"20:00"`
	Comment   *string `json:"comment,omitempty" example:"Day driver shift"`
	IsActive  *bool   `json:"is_active,omitempty" example:"true"`
}

type UpdateDriverShiftRequest struct {
	DriverID  *int64  `json:"driver_id,omitempty" example:"1"`
	ShiftDate *string `json:"shift_date,omitempty" example:"2026-05-14"`
	TimeFrom  *string `json:"time_from,omitempty" example:"08:00"`
	TimeTo    *string `json:"time_to,omitempty" example:"20:00"`
	Comment   *string `json:"comment,omitempty" example:"Updated driver shift note"`
	IsActive  *bool   `json:"is_active,omitempty" example:"true"`
}

type UpdateDriverShiftActivityRequest struct {
	IsActive bool `json:"is_active" example:"false"`
}

type DriverShiftListQuery struct {
	DriverID int64
	DateFrom string
	DateTo   string
	IsActive *bool
	Limit    int
	Offset   int
	SortBy   string
	Order    string
}

type DriverShiftDriverBriefResponse struct {
	ID         int64  `json:"id" example:"1"`
	IIN        string `json:"iin" example:"990101350011"`
	Name       string `json:"name" example:"Ivan"`
	Surname    string `json:"surname" example:"Ivanov"`
	Middlename string `json:"middlename,omitempty" example:"Ivanovich"`
	Phone      string `json:"phone,omitempty" example:"+77001234567"`
	Mail       string `json:"mail,omitempty" example:"driver@mail.com"`
}

type DriverShiftTripsheetBriefResponse struct {
	ID                 int64   `json:"id" example:"1"`
	TripsheetNumber    string  `json:"tripsheet_number" example:"TS-0001"`
	TripsheetDate      string  `json:"tripsheet_date" example:"2026-05-14"`
	VehicleID          *int64  `json:"vehicle_id,omitempty" example:"1"`
	VehicleBrand       *string `json:"vehicle_brand,omitempty" example:"Toyota Camry"`
	VehiclePlateNumber string  `json:"vehicle_plate_number" example:"777ABC01"`
	StartTime          *string `json:"start_time,omitempty" example:"2026-05-14T08:00:00Z"`
	EndTime            *string `json:"end_time,omitempty" example:"2026-05-14T18:00:00Z"`
	StatusID           int64   `json:"status_id" example:"1"`
	StatusName         *string `json:"status_name,omitempty" example:"Создано"`
	CreatedAt          string  `json:"created_at" example:"2026-05-14T06:00:00Z"`
	UpdatedAt          string  `json:"updated_at" example:"2026-05-14T06:00:00Z"`
}

type DriverShiftResponse struct {
	ID              int64                               `json:"id" example:"1"`
	DriverID        int64                               `json:"driver_id" example:"1"`
	Driver          DriverShiftDriverBriefResponse      `json:"driver"`
	ShiftDate       string                              `json:"shift_date" example:"2026-05-14"`
	TimeFrom        string                              `json:"time_from" example:"08:00"`
	TimeTo          *string                             `json:"time_to,omitempty" example:"20:00"`
	Comment         *string                             `json:"comment,omitempty" example:"Day driver shift"`
	IsActive        bool                                `json:"is_active" example:"true"`
	TripsheetsCount int64                               `json:"tripsheets_count" example:"3"`
	Tripsheets      []DriverShiftTripsheetBriefResponse `json:"tripsheets"`
	CreatedAt       time.Time                           `json:"created_at"`
	UpdatedAt       time.Time                           `json:"updated_at"`
}

type DriverShiftListResponse struct {
	Items  []DriverShiftResponse `json:"items"`
	Total  int64                 `json:"total" example:"1"`
	Limit  int                   `json:"limit" example:"50"`
	Offset int                   `json:"offset" example:"0"`
}
