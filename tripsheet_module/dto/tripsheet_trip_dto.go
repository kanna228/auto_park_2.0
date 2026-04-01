package dto

// CreateTripsheetTripRequest represents request body for creating a tripsheet trip.
type CreateTripsheetTripRequest struct {
	TripsheetID      int64   `json:"tripsheet_id" example:"1"`
	RouteDescription string  `json:"route_description" example:"Warehouse - Airport"`
	StartTime        *string `json:"start_time,omitempty" example:"2026-03-19T08:00:00Z"`
	EndTime          *string `json:"end_time,omitempty" example:"2026-03-19T10:00:00Z"`
	DistancePassed   *int    `json:"distance_passed,omitempty" example:"25"`
	StatusID         int64   `json:"status_id" example:"1"`
}

// UpdateTripsheetTripRequest represents request body for updating a tripsheet trip.
type UpdateTripsheetTripRequest struct {
	TripsheetID      int64   `json:"tripsheet_id" example:"1"`
	RouteDescription string  `json:"route_description" example:"Warehouse - Airport - Base"`
	StartTime        *string `json:"start_time,omitempty" example:"2026-03-19T08:00:00Z"`
	EndTime          *string `json:"end_time,omitempty" example:"2026-03-19T11:00:00Z"`
	DistancePassed   *int    `json:"distance_passed,omitempty" example:"40"`
	StatusID         int64   `json:"status_id" example:"2"`
}

// TripsheetTripResponse represents tripsheet trip response.
type TripsheetTripResponse struct {
	ID               int64   `json:"id" example:"1"`
	TripsheetID      int64   `json:"tripsheet_id" example:"1"`
	RouteDescription string  `json:"route_description" example:"Warehouse - Airport"`
	StartTime        *string `json:"start_time,omitempty" example:"2026-03-19T08:00:00Z"`
	EndTime          *string `json:"end_time,omitempty" example:"2026-03-19T10:00:00Z"`
	DistancePassed   int     `json:"distance_passed" example:"25"`
	StatusID         int64   `json:"status_id" example:"1"`
	CreatedAt        string  `json:"created_at" example:"2026-03-19T08:00:00Z"`
	UpdatedAt        string  `json:"updated_at" example:"2026-03-19T08:00:00Z"`
}

// TripsheetTripFilter represents filters for listing tripsheet trips.
type TripsheetTripFilter struct {
	TripsheetID *int64  `form:"tripsheet_id" example:"1"`
	StatusID    *int64  `form:"status_id" example:"1"`
	DateFrom    *string `form:"date_from" example:"2026-03-01"`
	DateTo      *string `form:"date_to" example:"2026-03-31"`
}

// TripsheetTripListResponse represents list response for tripsheet trips.
type TripsheetTripListResponse struct {
	Items []TripsheetTripResponse `json:"items"`
	Total int                     `json:"total" example:"1"`
}
