package handlers

import "auto_park/modules/notification_module/dto"

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}

type NotificationResponseWrap struct {
	Success bool                     `json:"success" example:"true"`
	Data    dto.NotificationResponse `json:"data"`
}

type NotificationListResponseWrap struct {
	Success bool                         `json:"success" example:"true"`
	Data    dto.NotificationListResponse `json:"data"`
}

type NotificationUnreadCountResponseWrap struct {
	Success bool                                `json:"success" example:"true"`
	Data    dto.NotificationUnreadCountResponse `json:"data"`
}

type NotificationMarkAllReadResponseWrap struct {
	Success bool                                `json:"success" example:"true"`
	Data    dto.NotificationMarkAllReadResponse `json:"data"`
}
