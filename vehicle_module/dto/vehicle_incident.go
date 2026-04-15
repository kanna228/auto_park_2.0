package dto

import "time"

type VehicleIncidentDriverInfo struct {
	ID         int64   `json:"id"`
	IIN        string  `json:"iin"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty"`
}

type VehicleIncidentMechanicInfo struct {
	ID         int64   `json:"id"`
	IIN        string  `json:"iin"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      string  `json:"email"`
	RoleID     int64   `json:"role_id"`
	RoleName   string  `json:"role_name"`
}

type VehicleIncidentHistoryItem struct {
	ID               int64  `json:"id"`
	IncidentTypeID   int64  `json:"incident_type_id"`
	IncidentTypeName string `json:"incident_type_name"`

	VehicleID int64 `json:"vehicle_id"`

	TripsheetID     *int64  `json:"tripsheet_id,omitempty"`
	TripsheetNumber *string `json:"tripsheet_number,omitempty"`

	Date     time.Time `json:"date"`
	Time     string    `json:"time"`
	Location string    `json:"location"`
	Text     string    `json:"text"`

	Driver   VehicleIncidentDriverInfo   `json:"driver"`
	Mechanic VehicleIncidentMechanicInfo `json:"mechanic"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
