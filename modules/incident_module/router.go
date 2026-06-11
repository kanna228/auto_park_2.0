package incident_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/incident_module/handlers"
	"auto_park/modules/incident_module/repository"
	"auto_park/modules/incident_module/service"
	notificationservice "auto_park/modules/notification_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB, notifySvc notificationservice.NotificationService, auditSvc *auditlogservice.Service) {
	incidentRepo := repository.NewIncidentRepository(db)
	incidentSvc := service.NewIncidentService(incidentRepo, notifySvc)
	incidentHandler := handlers.NewIncidentHandler(incidentSvc, auditSvc)

	api := r.Group("/api/incidents")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("/types", incidentHandler.ListIncidentTypes)
		protected.GET("", incidentHandler.GetAll)
		protected.GET("/:id", incidentHandler.GetByID)

		writeGroup := protected.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			writeGroup.POST("", incidentHandler.Create)
			writeGroup.PUT("/:id", incidentHandler.Update)
			writeGroup.DELETE("/:id", incidentHandler.Delete)
		}
	}
}
