package dto

type VehicleListQuery struct {
	Search      string
	BoardNumber string
	StateNumber string
	VIN         string
	BrandModel  string

	StatusID   *int64
	StatusName string

	ManufactureYearFrom *int
	ManufactureYearTo   *int

	DriverID *int64

	Limit  int
	Offset int

	SortBy          string
	Order           string
	IncludeArchived bool
}

type VehicleListResponse struct {
	Items  []VehicleResponse `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
