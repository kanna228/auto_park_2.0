package handlers

import "auto_park/modules/maintenance_schedule_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type MaintenanceScheduleResponseWrap struct {
	Success bool                            `json:"success" example:"true"`
	Data    dto.MaintenanceScheduleResponse `json:"data"`
}

type MaintenanceScheduleListResponseWrap struct {
	Success bool                                `json:"success" example:"true"`
	Data    dto.MaintenanceScheduleListResponse `json:"data"`
}

type MaintenanceScheduleDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
