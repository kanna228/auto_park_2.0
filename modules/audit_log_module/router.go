package audit_log_module

import (
	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/audit_log_module/handlers"
	"auto_park/modules/audit_log_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, svc *service.Service) {
	h := handlers.NewAuditLogHandler(svc)

	api := r.Group("/api/audit-logs")
	api.Use(middleware.AuthJWT(cfg), middleware.RequireRoles(1))
	{
		api.GET("", h.ListAuditLogs)
		api.GET("/export", h.ExportAuditLogs)
	}
}
