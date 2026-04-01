package handlers

import "auto_park/vehicle_module/dto"

// TireResponseWrap — ответ с одной шиной
type TireResponseWrap struct {
	Success bool             `json:"success" example:"true"`
	Data    dto.TireResponse `json:"data"`
}

// TireListResponseWrap — ответ со списком шин
type TireListResponseWrap struct {
	Success bool                 `json:"success" example:"true"`
	Data    dto.TireListResponse `json:"data"`
}

// TireCreateResponseWrap — ответ при создании шины
type TireCreateResponseWrap struct {
	Success bool                   `json:"success" example:"true"`
	Data    dto.TireCreateResponse `json:"data"`
}

// TireDeleteResponseWrap — ответ при удалении шины
type TireDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

// TireUpdateResponseWrap — ответ при обновлении шины
type TireUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
