package dto

import "time"

type PartCreateRequest struct {
	PartID           string  `json:"part_id" binding:"required" example:"BRK-PAD-001"`
	Name             string  `json:"name" binding:"required" example:"Brake Pad Front"`
	StartQuantity    int64   `json:"start_quantity" binding:"gte=0" example:"50"`
	MinStockQuantity int64   `json:"min_stock_quantity" example:"5"`
	Unit             string  `json:"unit" example:"шт"`
	Price            float64 `json:"price" example:"25000"`
	Category         string  `json:"category" binding:"required" example:"brake_system"`
	Dimensions       *string `json:"dimensions,omitempty" example:"120x45x18 mm"`
	Manufacturer     *string `json:"manufacturer,omitempty" example:"Bosch"`
	IsConsumable     bool    `json:"is_consumable" example:"false"`
}

type PartUpdateRequest struct {
	Name             string  `json:"name" binding:"required" example:"Brake Pad Front"`
	Quantity         int64   `json:"quantity" binding:"gte=0" example:"75"`
	MinStockQuantity int64   `json:"min_stock_quantity" example:"5"`
	Unit             string  `json:"unit" example:"шт"`
	Price            float64 `json:"price" example:"25000"`
	Category         string  `json:"category" binding:"required" example:"brake_system"`
	Dimensions       *string `json:"dimensions,omitempty" example:"120x45x18 mm"`
	Manufacturer     *string `json:"manufacturer,omitempty" example:"Bosch"`
	IsConsumable     bool    `json:"is_consumable" example:"false"`
}

type PartResponse struct {
	ID               int64     `json:"id" example:"1"`
	PartID           string    `json:"part_id" example:"BRK-PAD-001"`
	Name             string    `json:"name" example:"Brake Pad Front"`
	Quantity         int64     `json:"quantity" example:"50"`
	MinStockQuantity int64     `json:"min_stock_quantity" example:"5"`
	Unit             string    `json:"unit" example:"шт"`
	Price            float64   `json:"price" example:"25000"`
	TotalValue       float64   `json:"total_value" example:"1250000"`
	Category         string    `json:"category" example:"brake_system"`
	Dimensions       *string   `json:"dimensions,omitempty" example:"120x45x18 mm"`
	Manufacturer     *string   `json:"manufacturer,omitempty" example:"Bosch"`
	IsConsumable     bool      `json:"is_consumable" example:"false"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PartListQuery struct {
	PartID   string
	Name     string
	Category string
	Limit    int
	Offset   int
	SortBy   string
	Order    string
}

type PartStockItemResponse struct {
	ID           int64   `json:"id"`
	PartID       string  `json:"part_id"`
	Name         string  `json:"name"`
	Quantity     int64   `json:"quantity"`
	Price        float64 `json:"price"`
	TotalValue   float64 `json:"total_value"`
	Category     string  `json:"category"`
	IsConsumable bool    `json:"is_consumable"`
}

type PartListResponse struct {
	Items  []PartResponse `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type PartSummaryResponse struct {
	Total           int64 `json:"total"`
	LowCount        int64 `json:"low_count"`
	CriticalCount   int64 `json:"critical_count"`
	IssuedLastMonth int64 `json:"issued_last_month"`
}

type PartStockMovementResponse struct {
	Type           string    `json:"type"`
	Quantity       int64     `json:"quantity"`
	Vehicle        *string   `json:"vehicle,omitempty"`
	PartRequestID  *int64    `json:"part_request_id,omitempty"`
	DocumentNumber *string   `json:"document_number,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	Actor          *string   `json:"actor,omitempty"`
}

type PartStockMovementListResponse struct {
	Items  []PartStockMovementResponse `json:"items"`
	Total  int64                       `json:"total"`
	Limit  int                         `json:"limit"`
	Offset int                         `json:"offset"`
}

type PartArrivalCreateItemRequest struct {
	PartID   int64    `json:"part_id" binding:"gt=0"`
	Quantity int64    `json:"quantity" binding:"gt=0"`
	Price    *float64 `json:"price,omitempty" binding:"omitempty,gte=0"`
}

type PartArrivalCreateRequest struct {
	DocumentNumber string                         `json:"document_number" binding:"required"`
	ArrivalDate    string                         `json:"arrival_date" binding:"required" example:"2026-05-15"`
	Comment        *string                        `json:"comment,omitempty"`
	Items          []PartArrivalCreateItemRequest `json:"items" binding:"required"`
}

type PartArrivalItemResponse struct {
	ID          int64    `json:"id"`
	PartID      int64    `json:"part_id"`
	PartCode    string   `json:"part_code"`
	PartName    string   `json:"part_name"`
	Quantity    int64    `json:"quantity"`
	Price       *float64 `json:"price,omitempty"`
	TotalAmount *float64 `json:"total_amount,omitempty"`
}

type PartArrivalResponse struct {
	ID             int64                     `json:"id"`
	DocumentNumber string                    `json:"document_number"`
	ArrivalDate    string                    `json:"arrival_date"`
	Status         string                    `json:"status"`
	Comment        *string                   `json:"comment,omitempty"`
	CreatedBy      *string                   `json:"created_by,omitempty"`
	AcceptedBy     *string                   `json:"accepted_by,omitempty"`
	AcceptedAt     *time.Time                `json:"accepted_at,omitempty"`
	Items          []PartArrivalItemResponse `json:"items"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

type PartArrivalListResponse struct {
	Items  []PartArrivalResponse `json:"items"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

type PartArrivalSummaryResponse struct {
	Total         int64 `json:"total"`
	DraftCount    int64 `json:"draft_count"`
	AcceptedCount int64 `json:"accepted_count"`
	AcceptedItems int64 `json:"accepted_items"`
}

type VehiclePartInstallationHistoryItem struct {
	EventType     string    `json:"event_type"`
	Vehicle       *string   `json:"vehicle,omitempty"`
	PartRequestID *int64    `json:"part_request_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	Actor         *string   `json:"actor,omitempty"`
}

type VehiclePartInstallationHistoryResponse struct {
	Items  []VehiclePartInstallationHistoryItem `json:"items"`
	Total  int64                                `json:"total"`
	Limit  int                                  `json:"limit"`
	Offset int                                  `json:"offset"`
}
