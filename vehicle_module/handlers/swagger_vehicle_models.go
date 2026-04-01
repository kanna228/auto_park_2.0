// services/vehicle_service/internal/handlers/swagger_vehicle_models.go
package handlers

import (
	"time"

	"auto_park/vehicle_module/dto"
)

// ErrorResponse — единый формат ошибки
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

// VehicleCreateResponseWrap — ответ на создание
type VehicleCreateResponseWrap struct {
	Success bool                      `json:"success" example:"true"`
	Data    dto.VehicleCreateResponse `json:"data"`
	TS      time.Time                 `json:"ts" example:"2026-02-27T12:00:00Z"`
}

// VehicleResponseWrap — ответ с одной машиной
type VehicleResponseWrap struct {
	Success bool                `json:"success" example:"true"`
	Data    dto.VehicleResponse `json:"data"`
}

// VehicleListResponseWrap — ответ списка (filters + pagination)
type VehicleListResponseWrap struct {
	Success bool                    `json:"success" example:"true"`
	Data    dto.VehicleListResponse `json:"data"`
}

// DeleteVehicleResponseWrap — ответ при удалении
type DeleteVehicleResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"123"`
	} `json:"data"`
}

// UpdateVehicleResponseWrap — ответ при апдейте (у тебя сейчас возвращается только id)
type UpdateVehicleResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"123"`
	} `json:"data"`
}
