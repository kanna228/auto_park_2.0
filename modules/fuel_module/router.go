package fuel_module

import (
	"database/sql"
	"net/http"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/fuel_module/handlers"
	"auto_park/modules/fuel_module/repository"
	"auto_park/modules/fuel_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	r.GET("/api/fuel/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "service": "fuel_module", "db": "down", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "fuel_module", "db": "up"})
	})

	repo := repository.NewFuelRefillRepository(db)
	reportRepo := repository.NewFuelReportRepository(db)
	svc := service.NewFuelRefillService(repo)
	reportSvc := service.NewFuelReportService(reportRepo)
	h := handlers.NewFuelRefillHandler(svc)
	reportH := handlers.NewFuelReportHandler(reportSvc)

	api := r.Group("/api/fuel")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("/refills", h.GetAll)
		protected.GET("/refills/:id", h.GetByID)
		protected.GET("/refills/by-tripsheet/:tripsheet_id", h.GetAllByTripsheetID)
		protected.GET("/refills/by-vehicle/:vehicle_id", h.GetAllByVehicleID)

		reportGroup := protected.Group("/reports")
		reportGroup.Use(middleware.RequireRoles(1, 2))
		{
			reportGroup.GET("/driver/:driver_id", reportH.DownloadDriverReport)
			reportGroup.GET("/vehicle/:vehicle_id", reportH.DownloadVehicleReport)
			reportGroup.GET("/tripsheet/:tripsheet_id", reportH.DownloadTripsheetReport)
		}

		writeGroup := protected.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			writeGroup.POST("/refills", h.Create)
			writeGroup.PUT("/refills/:id", h.Update)
			writeGroup.DELETE("/refills/:id", h.Delete)
		}
	}
}
