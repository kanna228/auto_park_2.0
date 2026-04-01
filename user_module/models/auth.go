// internal/models/auth.go
package models

import "time"

type UserAuth struct {
	ID           int64
	Email        string
	PassHash     string
	IIN          string
	RoleID       int64
	FirstName    string
	LastName     string
	MiddleName   *string
	SessionToken *string
	LastSeen     *time.Time
}

type AuthResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID    int64     `json:"user_id" example:"12"`
	Email     string    `json:"email" example:"admin@auto-park.kz"`
	RoleID    int64     `json:"role_id" example:"1"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
	ExpiresAt int64     `json:"expires_at" example:"1739999999"`
}
