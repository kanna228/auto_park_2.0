package dto

import "time"

type CreateMechanicShiftRequest struct {
	UserID    *int64  `json:"user_id,omitempty" example:"4"`
	ShiftDate string  `json:"shift_date" binding:"required" example:"2026-05-12"`
	TimeFrom  string  `json:"time_from" binding:"required" example:"09:00"`
	TimeTo    *string `json:"time_to,omitempty" example:"18:00"`
	Comment   *string `json:"comment,omitempty" example:"Morning garage inspection shift"`
	IsActive  *bool   `json:"is_active,omitempty" example:"true"`
}

type UpdateMechanicShiftRequest struct {
	UserID    *int64  `json:"user_id,omitempty" example:"4"`
	ShiftDate *string `json:"shift_date,omitempty" example:"2026-05-12"`
	TimeFrom  *string `json:"time_from,omitempty" example:"09:00"`
	TimeTo    *string `json:"time_to,omitempty" example:"18:00"`
	Comment   *string `json:"comment,omitempty" example:"Updated shift note"`
	IsActive  *bool   `json:"is_active,omitempty" example:"true"`
}

type UpdateMechanicShiftActivityRequest struct {
	IsActive bool `json:"is_active" example:"false"`
}

type MechanicShiftListQuery struct {
	UserID   int64
	DateFrom string
	DateTo   string
	IsActive *bool
	Limit    int
	Offset   int
	SortBy   string
	Order    string
}

type MechanicShiftUserBriefResponse struct {
	ID         int64   `json:"id" example:"4"`
	Email      string  `json:"email" example:"mechanic@example.com"`
	FirstName  string  `json:"first_name" example:"Ivan"`
	LastName   string  `json:"last_name" example:"Ivanov"`
	MiddleName *string `json:"middle_name,omitempty" example:"Ivanovich"`
	RoleID     int64   `json:"role_id" example:"4"`
	RoleName   string  `json:"role_name" example:"duty_mechanic"`
}

type MechanicShiftResponse struct {
	ID        int64                          `json:"id" example:"1"`
	UserID    int64                          `json:"user_id" example:"4"`
	User      MechanicShiftUserBriefResponse `json:"user"`
	ShiftDate string                         `json:"shift_date" example:"2026-05-12"`
	TimeFrom  string                         `json:"time_from" example:"09:00"`
	TimeTo    *string                        `json:"time_to,omitempty" example:"18:00"`
	Comment   *string                        `json:"comment,omitempty" example:"Morning garage inspection shift"`
	IsActive  bool                           `json:"is_active" example:"true"`
	CreatedAt time.Time                      `json:"created_at"`
	UpdatedAt time.Time                      `json:"updated_at"`
}

type MechanicShiftListResponse struct {
	Items  []MechanicShiftResponse `json:"items"`
	Total  int64                   `json:"total" example:"1"`
	Limit  int                     `json:"limit" example:"50"`
	Offset int                     `json:"offset" example:"0"`
}
