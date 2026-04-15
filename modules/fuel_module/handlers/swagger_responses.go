package handlers

import "auto_park/modules/fuel_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"fuel refill deleted successfully"`
}

type FuelRefillResponseWrap struct {
	Success bool                   `json:"success" example:"true"`
	Data    dto.FuelRefillResponse `json:"data"`
}

type FuelRefillListResponse struct {
	Success bool                     `json:"success" example:"true"`
	Items   []dto.FuelRefillResponse `json:"items"`
	Total   int64                    `json:"total" example:"10"`
}
