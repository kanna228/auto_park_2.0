package dto

type VehicleTireCreateRequest struct {
	Position     string `json:"position" binding:"required"`
	Type         string `json:"type" binding:"required"`
	MileageKM    int64  `json:"mileage_km"`
	MaxMileageKM int64  `json:"max_mileage_km" binding:"required"`
	InstalledAt  string `json:"installed_at,omitempty"`
}

type VehicleTireUpdateRequest struct {
	Position     *string `json:"position,omitempty"`
	Type         *string `json:"type,omitempty"`
	MileageKM    *int64  `json:"mileage_km,omitempty"`
	MaxMileageKM *int64  `json:"max_mileage_km,omitempty"`
	InstalledAt  *string `json:"installed_at,omitempty"`
}

type VehicleTireResponse struct {
	ID               int64  `json:"id"`
	VehicleID        int64  `json:"vehicle_id"`
	Position         string `json:"position"`
	Type             string `json:"type"`
	MileageKM        int64  `json:"mileage_km"`
	MaxMileageKM     int64  `json:"max_mileage_km"`
	RemainingPercent int64  `json:"remaining_percent"`
	InstalledAt      string `json:"installed_at"`
}
