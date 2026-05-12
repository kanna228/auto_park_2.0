package dto

import "time"

type NotificationResponse struct {
	ID                 int64          `json:"id" example:"1"`
	UserID             int64          `json:"user_id" example:"10"`
	NotificationTypeID int64          `json:"notification_type_id" example:"1"`
	TypeCode           string         `json:"type_code" example:"part_request_created"`
	TypeName           string         `json:"type_name" example:"Создана заявка на деталь"`
	Title              string         `json:"title" example:"Новая заявка на деталь"`
	Message            string         `json:"message" example:"Механик создал новую заявку"`
	Context            map[string]any `json:"context"`
	DedupKey           *string        `json:"dedup_key,omitempty" example:"vehicle_part_replacement:7_days:vpi:1:date:2026-05-19"`
	IsReaded           bool           `json:"is_readed" example:"false"`
	ReadAt             *time.Time     `json:"read_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type NotificationListResponse struct {
	Items  []NotificationResponse `json:"items"`
	Total  int64                  `json:"total" example:"1"`
	Limit  int                    `json:"limit" example:"50"`
	Offset int                    `json:"offset" example:"0"`
}

type NotificationUnreadCountResponse struct {
	Count int64 `json:"count" example:"5"`
}

type NotificationMarkAllReadResponse struct {
	Updated int64 `json:"updated" example:"3"`
}

type WebSocketMessage struct {
	Event string `json:"event" example:"notification.created"`
	Data  any    `json:"data"`
}
