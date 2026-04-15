package incident_module

import (
	"database/sql"

	"auto_park/incident_module/handlers"
	"auto_park/incident_module/repository"
	"auto_park/incident_module/service"
	"auto_park/internal/config"
	"auto_park/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	incidentRepo := repository.NewIncidentRepository(db)
	incidentSvc := service.NewIncidentService(incidentRepo)
	incidentHandler := handlers.NewIncidentHandler(incidentSvc)

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
