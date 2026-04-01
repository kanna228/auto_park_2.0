package dto

type VehicleListQuery struct {
	// фильтры
	BoardNumber string
	StateNumber string
	VIN         string
	BrandModel  string

	ManufactureYearFrom *int
	ManufactureYearTo   *int

	DriverID *int64 // искать WHERE $x = ANY(drivers_ids)

	// пагинация
	Limit  int
	Offset int

	// сортировка (простая/безопасная)
	SortBy string // "id", "board_number", "state_number", "manufacture_year", "mileage", "created_at"
	Order  string // "asc" | "desc"
}

type VehicleListResponse struct {
	Items  []VehicleResponse `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
