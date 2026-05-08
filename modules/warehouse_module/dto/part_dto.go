package dto

import "time"

type PartCreateRequest struct {
	PartID        string  `json:"part_id" binding:"required" example:"BRK-PAD-001"`
	Name          string  `json:"name" binding:"required" example:"Brake Pad Front"`
	StartQuantity int64   `json:"start_quantity" binding:"required" example:"50"`
	Category      string  `json:"category" binding:"required" example:"brake_system"`
	Dimensions    *string `json:"dimensions,omitempty" example:"120x45x18 mm"`
	Manufacturer  *string `json:"manufacturer,omitempty" example:"Bosch"`
	IsConsumable  bool    `json:"is_consumable" example:"false"`
}

type PartUpdateRequest struct {
	Name         string  `json:"name" binding:"required" example:"Brake Pad Front"`
	Quantity     int64   `json:"quantity" binding:"required" example:"75"`
	Category     string  `json:"category" binding:"required" example:"brake_system"`
	Dimensions   *string `json:"dimensions,omitempty" example:"120x45x18 mm"`
	Manufacturer *string `json:"manufacturer,omitempty" example:"Bosch"`
	IsConsumable bool    `json:"is_consumable" example:"false"`
}

type PartResponse struct {
	ID           int64     `json:"id" example:"1"`
	PartID       string    `json:"part_id" example:"BRK-PAD-001"`
	Name         string    `json:"name" example:"Brake Pad Front"`
	Quantity     int64     `json:"quantity" example:"50"`
	Category     string    `json:"category" example:"brake_system"`
	Dimensions   *string   `json:"dimensions,omitempty" example:"120x45x18 mm"`
	Manufacturer *string   `json:"manufacturer,omitempty" example:"Bosch"`
	IsConsumable bool      `json:"is_consumable" example:"false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

type PartListResponse struct {
	Items  []PartResponse `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}
