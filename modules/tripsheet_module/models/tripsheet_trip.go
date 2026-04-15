package models

import "time"

type TripsheetTrip struct {
	ID               int64
	TripsheetID      int64
	RouteDescription string
	StartTime        *time.Time
	EndTime          *time.Time
	DistancePassed   int
	StatusID         int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateTripsheetTripInput struct {
	TripsheetID      int64
	RouteDescription string
	StartTime        *time.Time
	EndTime          *time.Time
	DistancePassed   int
	StatusID         int64
}

type UpdateTripsheetTripInput struct {
	ID               int64
	TripsheetID      int64
	RouteDescription string
	StartTime        *time.Time
	EndTime          *time.Time
	DistancePassed   int
	StatusID         int64
}
