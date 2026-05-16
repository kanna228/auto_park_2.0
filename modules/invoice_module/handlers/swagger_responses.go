package handlers

import "auto_park/modules/invoice_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

type InvoiceResponseWrap struct {
	Success bool                `json:"success" example:"true"`
	Data    dto.InvoiceResponse `json:"data"`
}

type InvoiceListResponseWrap struct {
	Success bool                    `json:"success" example:"true"`
	Data    dto.InvoiceListResponse `json:"data"`
}
