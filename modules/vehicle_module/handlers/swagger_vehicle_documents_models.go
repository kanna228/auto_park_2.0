package handlers

import "auto_park/modules/vehicle_module/dto"

type InsuranceResponseWrap struct {
	Success bool                  `json:"success" example:"true"`
	Data    dto.InsuranceResponse `json:"data"`
}

type InsuranceListResponseWrap struct {
	Success bool                      `json:"success" example:"true"`
	Data    dto.InsuranceListResponse `json:"data"`
}

type InsuranceCreateResponseWrap struct {
	Success bool                        `json:"success" example:"true"`
	Data    dto.InsuranceCreateResponse `json:"data"`
}

type InsuranceDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type InsuranceUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type TechnicalInspectionResponseWrap struct {
	Success bool                            `json:"success" example:"true"`
	Data    dto.TechnicalInspectionResponse `json:"data"`
}

type TechnicalInspectionListResponseWrap struct {
	Success bool                                `json:"success" example:"true"`
	Data    dto.TechnicalInspectionListResponse `json:"data"`
}

type TechnicalInspectionCreateResponseWrap struct {
	Success bool                                  `json:"success" example:"true"`
	Data    dto.TechnicalInspectionCreateResponse `json:"data"`
}

type TechnicalInspectionDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type TechnicalInspectionUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
