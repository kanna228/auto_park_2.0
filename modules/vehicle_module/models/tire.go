package models

import "time"

type Tire struct {
	ID          int64     `json:"id"`
	PlaceID     int64     `json:"place_id"`
	PlaceName   string    `json:"place_name"`
	VehicleID   *int64    `json:"vehicle_id,omitempty"`
	Tire        string    `json:"tire"`
	Mileage     int64     `json:"mileage"`
	MaxUsage    int64     `json:"max_usage"`
	InstalledAt time.Time `json:"installed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type TirePlace struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
