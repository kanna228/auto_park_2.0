package models

import "time"

type PurchaseRequest struct {
	ID                  int64
	PartID              int64
	PartCatalogCode     string
	PartName            string
	PartCategory        string
	Quantity            int64
	Status              string
	SourcePartRequestID *int64
	Comment             *string
	CreatedByUserID     int64
	CreatedByEmail      *string
	CreatedByFullName   *string
	ConfirmedByUserID   *int64
	ConfirmedByEmail    *string
	ConfirmedByFullName *string
	ConfirmedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
