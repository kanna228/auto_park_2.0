package dto

type CreateFuelRefillRequest struct {
	TripsheetID int64   `json:"tripsheet_id" example:"1"`
	VehicleID   int64   `json:"vehicle_id" example:"12"`
	FuelAmount  float64 `json:"fuel_amount" example:"25.5"`
	Date        string  `json:"date" example:"2026-04-15"`
	Time        string  `json:"time" example:"14:30:00"`
	Location    *string `json:"location,omitempty" example:"Helios, Astana"`
}

type UpdateFuelRefillRequest struct {
	TripsheetID int64   `json:"tripsheet_id" example:"1"`
	VehicleID   int64   `json:"vehicle_id" example:"12"`
	FuelAmount  float64 `json:"fuel_amount" example:"30"`
	Date        string  `json:"date" example:"2026-04-15"`
	Time        string  `json:"time" example:"17:45:00"`
	Location    *string `json:"location,omitempty" example:"Sinooil, Astana"`
}

type FuelRefillResponse struct {
	ID          int64   `json:"id" example:"1"`
	TripsheetID int64   `json:"tripsheet_id" example:"1"`
	VehicleID   int64   `json:"vehicle_id" example:"12"`
	FuelAmount  float64 `json:"fuel_amount" example:"25.5"`
	Date        string  `json:"date" example:"2026-04-15"`
	Time        string  `json:"time" example:"14:30:00"`
	Location    *string `json:"location,omitempty" example:"Helios, Astana"`
	CreatedAt   string  `json:"created_at" example:"2026-04-15T14:30:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2026-04-15T14:30:00Z"`
}

type FuelRefillFilter struct {
	TripsheetID *int64  `form:"tripsheet_id" example:"1"`
	VehicleID   *int64  `form:"vehicle_id" example:"12"`
	DriverID    *int64  `form:"driver_id" example:"55"`
	DateFrom    *string `form:"date_from" example:"2026-04-01"`
	DateTo      *string `form:"date_to" example:"2026-04-30"`
	Limit       int     `form:"limit" example:"50"`
	Offset      int     `form:"offset" example:"0"`
	SortBy      string  `form:"sort_by" example:"date"`
	Order       string  `form:"order" example:"desc"`
}
