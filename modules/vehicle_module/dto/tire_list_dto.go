package dto

type TireListQuery struct {
	VehicleID *int64
	PlaceID   *int64
	Tire      string
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type TireListResponse struct {
	Items  []TireResponse `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}
