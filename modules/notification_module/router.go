package notification_module

import (
	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/notification_module/handlers"
	notificationservice "auto_park/modules/notification_module/service"
	"auto_park/modules/notification_module/websocket"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, svc notificationservice.NotificationService, hub *websocket.Hub) {
	h := handlers.NewNotificationHandler(svc, hub)

	api := r.Group("/api/notifications")
	api.Use(middleware.AuthJWT(cfg))
	api.Use(middleware.RequireRoles(1, 2, 3, 4, 5))
	{
		api.GET("/ws", h.WebSocket)
		api.GET("", h.List)
		api.GET("/unread", h.ListUnread)
		api.GET("/unread/count", h.CountUnread)
		api.GET("/unread-count", h.CountUnreadPlain)
		api.PATCH("/:id/read", h.MarkAsRead)
		api.PATCH("/read-all", h.MarkAllAsRead)
		api.PATCH("/read-by-type", h.MarkAllAsReadByType)
		api.PUT("/:id/read", h.MarkAsRead)
		api.PUT("/read-all", h.MarkAllAsRead)
	}
}
