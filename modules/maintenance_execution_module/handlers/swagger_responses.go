package handlers

import "auto_park/modules/maintenance_execution_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type MaintenanceExecutionResponseWrap struct {
	Success bool                             `json:"success" example:"true"`
	Data    dto.MaintenanceExecutionResponse `json:"data"`
}

type MaintenanceExecutionListResponseWrap struct {
	Success bool                                 `json:"success" example:"true"`
	Data    dto.MaintenanceExecutionListResponse `json:"data"`
}

type MaintenanceExecutionDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
