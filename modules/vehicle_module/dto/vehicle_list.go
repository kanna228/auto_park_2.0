package dto

type VehicleListQuery struct {
	BoardNumber string
	StateNumber string
	VIN         string
	BrandModel  string

	StatusID *int64

	ManufactureYearFrom *int
	ManufactureYearTo   *int

	DriverID *int64

	Limit  int
	Offset int

	SortBy string
	Order  string
}

type VehicleListResponse struct {
	Items  []VehicleResponse `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
