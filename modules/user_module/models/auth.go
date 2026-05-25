// internal/models/auth.go
package models

import "time"

type UserAuth struct {
	ID           int64
	AccountType  string
	Email        string
	PassHash     string
	IIN          string
	RoleID       int64
	RoleName     string
	FirstName    string
	LastName     string
	MiddleName   *string
	DriverID     *int64
	SessionToken *string
	LastSeen     *time.Time
}

type AuthResponse struct {
	Token       string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID      int64     `json:"user_id" example:"12"`
	AccountType string    `json:"account_type" example:"user"`
	DriverID    *int64    `json:"driver_id,omitempty" example:"42"`
	Email       string    `json:"email" example:"admin@auto-park.kz"`
	RoleID      int64     `json:"role_id" example:"1"`
	RoleName    string    `json:"role_name" example:"admin"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
	ExpiresAt   int64     `json:"expires_at" example:"1739999999"`
}
