package dashboard_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/dashboard_module/handlers"
	"auto_park/modules/dashboard_module/repository"
	"auto_park/modules/dashboard_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	repo := repository.NewDashboardRepository(db)
	svc := service.NewDashboardService(repo)
	h := handlers.NewDashboardHandler(svc)

	api := r.Group("/api/dashboard")
	api.Use(middleware.AuthJWT(cfg))
	api.Use(middleware.RequireRoles(1, 2, 5))
	{
		api.GET("/stats", h.GetStats)
		api.GET("/stats/export/excel", h.ExportStatsExcel)
		api.GET("/stats/export/pdf", h.ExportStatsPDF)
	}

	analytics := r.Group("/api/analytics")
	analytics.Use(middleware.AuthJWT(cfg))
	analytics.Use(middleware.RequireRoles(1, 2, 5))
	{
		analytics.GET("/dashboard", h.GetAnalyticsDashboard)
		analytics.GET("/export", h.ExportAnalytics)
	}
}
