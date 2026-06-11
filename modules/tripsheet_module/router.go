package tripsheet_module

import (
	"database/sql"
	"net/http"

	"auto_park/internal/config"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/tripsheet_module/handlers"
	"auto_park/modules/tripsheet_module/repository"
	"auto_park/modules/tripsheet_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB, auditSvc *auditlogservice.Service) {
	r.GET("/api/tripsheet/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "service": "tripsheet_module", "db": "down", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "tripsheet_module", "db": "up"})
	})

	tripsheetRepo := repository.NewTripsheetRepo(db, cfg.Schemas.Tripsheet)
	tripsheetTripRepo := repository.NewTripsheetTripRepo(db, cfg.Schemas.Tripsheet)
	waybillRoutePointRepo := repository.NewWaybillRoutePointRepository(db)
	tripsheetSvc := service.NewTripsheetService(tripsheetRepo)
	tripsheetTripSvc := service.NewTripsheetTripService(tripsheetTripRepo)
	waybillRoutePointSvc := service.NewWaybillRoutePointService(waybillRoutePointRepo)
	tripsheetH := handlers.NewTripsheetHandler(tripsheetSvc, auditSvc)
	tripsheetTripH := handlers.NewTripsheetTripHandler(tripsheetTripSvc)
	waybillRoutePointH := handlers.NewWaybillRoutePointHandler(waybillRoutePointSvc)

	api := r.Group("/api/tripsheet")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("", tripsheetH.GetAll)
		protected.GET("/:id", tripsheetH.GetByID)
		protected.GET("/:id/routes", waybillRoutePointH.List)
		protected.GET("/trips", tripsheetTripH.GetAll)
		protected.GET("/trips/:id", tripsheetTripH.GetByID)
		protected.GET("/trips/by-tripsheet/:tripsheet_id", tripsheetTripH.GetAllByTripsheetID)

		writeGroup := protected.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			writeGroup.POST("", tripsheetH.Create)
			writeGroup.PUT("/:id", tripsheetH.Update)
			writeGroup.DELETE("/:id", tripsheetH.Delete)
			writeGroup.POST("/:id/routes", waybillRoutePointH.Create)
			writeGroup.PUT("/:id/routes/:rid", waybillRoutePointH.Update)
			writeGroup.DELETE("/:id/routes/:rid", waybillRoutePointH.Delete)
			writeGroup.POST("/trips", tripsheetTripH.Create)
			writeGroup.PUT("/trips/:id", tripsheetTripH.Update)
			writeGroup.DELETE("/trips/:id", tripsheetTripH.Delete)
		}
	}

	waybills := r.Group("/api/waybills")
	waybills.Use(middleware.AuthJWT(cfg))
	{
		waybills.GET("/:id/routes", waybillRoutePointH.List)

		waybillWriteGroup := waybills.Group("")
		waybillWriteGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			waybillWriteGroup.POST("/:id/routes", waybillRoutePointH.Create)
			waybillWriteGroup.PUT("/:id/routes/:rid", waybillRoutePointH.Update)
			waybillWriteGroup.DELETE("/:id/routes/:rid", waybillRoutePointH.Delete)
		}
	}
}
