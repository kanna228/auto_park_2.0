package handlers

import "auto_park/modules/dashboard_module/dto"

type DashboardStatsResponseWrap struct {
	Success bool                       `json:"success" example:"true"`
	Data    dto.DashboardStatsResponse `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}
