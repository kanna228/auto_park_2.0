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

type PartRequestResponseWrap struct {
	Success bool                    `json:"success" example:"true"`
	Data    dto.PartRequestResponse `json:"data"`
}

type PartRequestListResponseWrap struct {
	Success bool                        `json:"success" example:"true"`
	Data    dto.PartRequestListResponse `json:"data"`
}

type PartRequestStatusListResponseWrap struct {
	Success bool                            `json:"success" example:"true"`
	Data    []dto.PartRequestStatusResponse `json:"data"`
}

type PartRequestCreateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type PartRequestUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type PartRequestStatusUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type PartRequestDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type VehiclePartInstallationResponseWrap struct {
	Success bool                                `json:"success" example:"true"`
	Data    dto.VehiclePartInstallationResponse `json:"data"`
}

type VehiclePartInstallationListResponseWrap struct {
	Success bool                                    `json:"success" example:"true"`
	Data    dto.VehiclePartInstallationListResponse `json:"data"`
}

type VehiclePartInstallationCreateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type VehiclePartInstallationUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type VehiclePartInstallationActivityUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type VehiclePartInstallationDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}
