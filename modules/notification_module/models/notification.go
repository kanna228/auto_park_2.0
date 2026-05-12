package models

import (
	"encoding/json"
	"time"
)

type NotificationType struct {
	ID        int64
	Code      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Notification struct {
	ID                 int64
	UserID             int64
	NotificationTypeID int64
	TypeCode           string
	TypeName           string
	Title              string
	Message            string
	Context            json.RawMessage
	IsReaded           bool
	ReadAt             *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
