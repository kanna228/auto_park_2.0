package handlers

import "auto_park/modules/fuel_norm_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type FuelNormResponseWrap struct {
	Success bool                 `json:"success" example:"true"`
	Data    dto.FuelNormResponse `json:"data"`
}

type FuelNormListResponseWrap struct {
	Success bool                     `json:"success" example:"true"`
	Data    dto.FuelNormListResponse `json:"data"`
}

type FuelNormDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
