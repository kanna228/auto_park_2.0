package dto

import "time"

type MaintenanceScheduleRequest struct {
	IsDraft           bool     `json:"is_draft" example:"false"`
	DateFrom          string   `json:"date_from" example:"2026-03-01"`
	DateTo            string   `json:"date_to" example:"2026-03-30"`
	ConsecutiveCount  int      `json:"consecutive_count" example:"13"`
	ConsecutiveUnit   string   `json:"consecutive_unit" example:"Days"`
	EveryCount        int      `json:"every_count" example:"1"`
	EveryUnit         string   `json:"every_unit" example:"Days"`
	TimeFrom          string   `json:"time_from" example:"09:00"`
	TimeTo            string   `json:"time_to" example:"12:00"`
	DurationValue     int      `json:"duration_value" example:"30"`
	DurationUnit      string   `json:"duration_unit" example:"Minutes"`
	LimitBoardsByTime bool     `json:"limit_boards_by_time" example:"false"`
	Categories        []string `json:"categories" example:"Tires,Hood,Spark plugs"`
	Boards            []string `json:"boards" example:"30,31,33"`
	Mechanics         []string `json:"mechanics" example:"Azamat Rakhimov"`
}

type MaintenanceScheduleResponse struct {
	ID                int64     `json:"id" example:"1"`
	IsDraft           bool      `json:"is_draft" example:"false"`
	DateFrom          string    `json:"date_from" example:"2026-03-01"`
	DateTo            string    `json:"date_to" example:"2026-03-30"`
	ConsecutiveCount  int       `json:"consecutive_count" example:"13"`
	ConsecutiveUnit   string    `json:"consecutive_unit" example:"Days"`
	EveryCount        int       `json:"every_count" example:"1"`
	EveryUnit         string    `json:"every_unit" example:"Days"`
	TimeFrom          string    `json:"time_from" example:"09:00"`
	TimeTo            string    `json:"time_to" example:"12:00"`
	DurationValue     int       `json:"duration_value" example:"30"`
	DurationUnit      string    `json:"duration_unit" example:"Minutes"`
	LimitBoardsByTime bool      `json:"limit_boards_by_time" example:"false"`
	Categories        []string  `json:"categories"`
	Boards            []string  `json:"boards"`
	Mechanics         []string  `json:"mechanics"`
	CreatedAt         time.Time `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt         time.Time `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

type MaintenanceScheduleListQuery struct {
	IsDraft *bool
	Limit   int
	Offset  int
}

type MaintenanceScheduleListResponse struct {
	Items  []MaintenanceScheduleResponse `json:"items"`
	Total  int64                         `json:"total" example:"1"`
	Limit  int                           `json:"limit" example:"50"`
	Offset int                           `json:"offset" example:"0"`
}
