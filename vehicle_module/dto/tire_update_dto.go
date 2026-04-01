package dto

type TireUpdateRequest struct {
	PlaceID   int64  `json:"place_id" binding:"required"`
	VehicleID *int64 `json:"vehicle_id,omitempty"`
	Tire      string `json:"tire" binding:"required"`
	Mileage   int64  `json:"mileage" binding:"required"`
	MaxUsage  int64  `json:"max_usage" binding:"required"`
}
