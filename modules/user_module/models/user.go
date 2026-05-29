package models

import "time"

type Role struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type User struct {
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	FirstName    string     `db:"first_name" json:"first_name"`
	LastName     string     `db:"last_name" json:"last_name"`
	MiddleName   *string    `db:"middle_name" json:"middle_name,omitempty"`
	IIN          string     `db:"iin" json:"iin"`
	Phone        *string    `db:"phone" json:"phone,omitempty"`
	PasswordHash string     `db:"password" json:"-"`
	RoleID       int64      `db:"role_id" json:"role_id"`
	SessionToken *string    `db:"session_token" json:"session_token,omitempty"`
	LastSeen     *time.Time `db:"last_seen" json:"last_seen,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

type UserPublic struct {
	ID         int64      `json:"id"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	MiddleName *string    `json:"middle_name,omitempty"`
	IIN        string     `json:"iin"`
	Phone      *string    `json:"phone,omitempty"`
	RoleID     int64      `json:"role_id"`
	RoleName   string     `json:"role_name,omitempty"`
	IsActive   bool       `json:"is_active"`
	IsArchived bool       `json:"is_archived"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	LastSeen   *time.Time `json:"last_seen,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
