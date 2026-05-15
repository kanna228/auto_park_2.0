package models

import "time"

type Part struct {
	ID               int64
	PartID           string
	Name             string
	Quantity         int64
	MinStockQuantity int64
	Unit             string
	Price            float64
	Category         string
	Dimensions       *string
	Manufacturer     *string
	IsConsumable     bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
