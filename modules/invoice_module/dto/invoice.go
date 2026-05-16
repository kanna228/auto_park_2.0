package dto

import "time"

type InvoiceCreateRequest struct {
	InvoiceNumber string `json:"invoice_number" example:"A223780"`
	InvoiceDate   string `json:"invoice_date" example:"2026-03-21"`
	PartRequestID *int64 `json:"part_request_id" example:"388"`
	RequestNumber string `json:"request_number" example:"388"`
}

type InvoiceResponse struct {
	ID            int64     `json:"id" example:"1"`
	InvoiceNumber string    `json:"invoice_number" example:"A223780"`
	InvoiceDate   string    `json:"invoice_date" example:"2026-03-21"`
	PartRequestID *int64    `json:"part_request_id" example:"388"`
	RequestNumber string    `json:"request_number" example:"388"`
	CreatedAt     time.Time `json:"created_at" example:"2026-03-21T09:44:00Z"`
}

type InvoiceListQuery struct {
	Date   string
	Search string
	Limit  int
	Offset int
}

type InvoiceListResponse struct {
	Items  []InvoiceResponse `json:"items"`
	Total  int64             `json:"total" example:"1"`
	Limit  int               `json:"limit" example:"50"`
	Offset int               `json:"offset" example:"0"`
}
