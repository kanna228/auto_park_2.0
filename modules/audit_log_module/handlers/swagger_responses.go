package handlers

import "auto_park/modules/audit_log_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type AuditLogListResponseWrap struct {
	Success bool                     `json:"success" example:"true"`
	Data    dto.AuditLogListResponse `json:"data"`
}
