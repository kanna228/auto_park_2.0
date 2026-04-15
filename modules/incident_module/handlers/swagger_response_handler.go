package handlers

import "auto_park/modules/incident_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type SuccessMessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"deleted successfully"`
}

type IncidentCreateResponseWrap struct {
	Success bool                       `json:"success" example:"true"`
	Data    dto.IncidentCreateResponse `json:"data"`
}

type IncidentResponseWrap struct {
	Success bool                 `json:"success" example:"true"`
	Data    dto.IncidentResponse `json:"data"`
}

type IncidentListResponseWrap struct {
	Success bool                   `json:"success" example:"true"`
	Items   []dto.IncidentResponse `json:"items"`
	Total   int                    `json:"total" example:"1"`
}

type IncidentTypeListResponseWrap struct {
	Success bool                       `json:"success" example:"true"`
	Items   []dto.IncidentTypeResponse `json:"items"`
	Total   int                        `json:"total" example:"3"`
}
