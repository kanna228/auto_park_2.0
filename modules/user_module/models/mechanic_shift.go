package models

import "time"

type MechanicShift struct {
	ID             int64
	UserID         int64
	UserEmail      string
	UserFirstName  string
	UserLastName   string
	UserMiddleName *string
	UserRoleID     int64
	UserRoleName   string
	ShiftDate      time.Time
	TimeFrom       string
	TimeTo         *string
	Comment        *string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
