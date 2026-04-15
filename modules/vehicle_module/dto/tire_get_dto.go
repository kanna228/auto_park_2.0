package dto

import "time"

type TireResponse struct {
	ID             int64     `json:"id"`
	PlaceID        int64     `json:"place_id"`
	PlaceName      string    `json:"place_name"`
	VehicleID      *int64    `json:"vehicle_id,omitempty"`
	Tire           string    `json:"tire"`
	Mileage        int64     `json:"mileage"`
	MaxUsage       int64     `json:"max_usage"`
	RemainingUsage int64     `json:"remaining_usage"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
