package dto

import "time"

type AuditLogResponse struct {
	ID         int64     `json:"id" example:"1"`
	Level      string    `json:"level" example:"success"`
	Function   string    `json:"function" example:"request"`
	FromStatus *string   `json:"from_status" example:"Processing"`
	ToStatus   *string   `json:"to_status" example:"Ready"`
	Actor      string    `json:"actor" example:"username_skl"`
	Message    *string   `json:"message" example:"status changed"`
	CreatedAt  time.Time `json:"created_at" example:"2026-05-16T09:44:00Z"`
}

type AuditLogListQuery struct {
	Function string
	Level    string
	Date     string
	Search   string
	Limit    int
	Offset   int
}

type AuditLogListResponse struct {
	Items  []AuditLogResponse `json:"items"`
	Total  int64              `json:"total" example:"1"`
	Limit  int                `json:"limit" example:"50"`
	Offset int                `json:"offset" example:"0"`
}
