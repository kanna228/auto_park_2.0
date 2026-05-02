package models

import "time"

type PartRequestStatus struct {
	ID   int64
	Code string
	Name string
}

type PartRequest struct {
	ID              int64
	PartID          int64
	PartCatalogCode string
	PartName        string
	PartCategory    string
	Quantity        int64
	MechanicComment string
	StatusID        int64
	StatusCode      string
	StatusName      string
	AuthorUserID    int64
	AuthorEmail     *string
	AuthorFullName  *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
