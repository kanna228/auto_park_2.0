package handlers

import "auto_park/vehicle_module/dto"

// TirePlaceCreateResponseWrap — ответ при создании места шины
type TirePlaceCreateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

// TirePlaceResponseWrap — ответ с одним местом шины
type TirePlaceResponseWrap struct {
	Success bool                  `json:"success" example:"true"`
	Data    dto.TirePlaceResponse `json:"data"`
}

// TirePlaceListResponseWrap — ответ со списком мест шин
type TirePlaceListResponseWrap struct {
	Success bool                      `json:"success" example:"true"`
	Data    dto.TirePlaceListResponse `json:"data"`
}

// TirePlaceUpdateResponseWrap — ответ при обновлении
type TirePlaceUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

// TirePlaceDeleteResponseWrap — ответ при удалении
type TirePlaceDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
