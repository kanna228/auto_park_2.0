package models

import "time"

type Insurance struct {
	ID        int64     `json:"id"`
	VehicleID int64     `json:"vehicle_id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	FilePath  string    `json:"file_path,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
