package dto

type VehicleDocumentCreateRequest struct {
	Type      string `json:"type" binding:"required"`
	Number    string `json:"number" binding:"required"`
	ValidFrom string `json:"valid_from" binding:"required"`
	ValidTo   string `json:"valid_to" binding:"required"`
}

type VehicleDocumentResponse struct {
	ID        int64  `json:"id"`
	VehicleID int64  `json:"vehicle_id"`
	Type      string `json:"type"`
	Number    string `json:"number"`
	ValidFrom string `json:"valid_from"`
	ValidTo   string `json:"valid_to"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
