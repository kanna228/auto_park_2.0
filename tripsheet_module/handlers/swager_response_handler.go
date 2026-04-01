package handlers

import "auto_park/tripsheet_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type SuccessMessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"deleted successfully"`
}

type TripsheetCreateResponseWrap struct {
	Success bool                        `json:"success" example:"true"`
	Data    dto.CreateTripsheetResponse `json:"data"`
}

type TripsheetResponseWrap struct {
	Success bool                  `json:"success" example:"true"`
	Data    dto.TripsheetResponse `json:"data"`
}

type TripsheetListResponseWrap struct {
	Success bool                    `json:"success" example:"true"`
	Items   []dto.TripsheetResponse `json:"items"`
	Total   int                     `json:"total" example:"1"`
}

type TripsheetTripResponseWrap struct {
	Success bool                      `json:"success" example:"true"`
	Data    dto.TripsheetTripResponse `json:"data"`
}

type TripsheetTripListResponseWrap struct {
	Success bool                        `json:"success" example:"true"`
	Items   []dto.TripsheetTripResponse `json:"items"`
	Total   int                         `json:"total" example:"1"`
}
