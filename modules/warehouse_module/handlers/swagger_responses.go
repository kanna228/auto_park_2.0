package handlers

import "auto_park/modules/warehouse_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type PartResponseWrap struct {
	Success bool             `json:"success" example:"true"`
	Data    dto.PartResponse `json:"data"`
}

type PartListResponseWrap struct {
	Success bool                 `json:"success" example:"true"`
	Data    dto.PartListResponse `json:"data"`
}

type PartCreateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type PartUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type PartDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
