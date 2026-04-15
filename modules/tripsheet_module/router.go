package tripsheet_module

import (
	"database/sql"
	"net/http"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/tripsheet_module/handlers"
	"auto_park/modules/tripsheet_module/repository"
	"auto_park/modules/tripsheet_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	r.GET("/api/tripsheet/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "service": "tripsheet_module", "db": "down", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "tripsheet_module", "db": "up"})
	})

	tripsheetRepo := repository.NewTripsheetRepo(db, cfg.Schemas.Tripsheet)
	tripsheetTripRepo := repository.NewTripsheetTripRepo(db, cfg.Schemas.Tripsheet)
	tripsheetSvc := service.NewTripsheetService(tripsheetRepo)
	tripsheetTripSvc := service.NewTripsheetTripService(tripsheetTripRepo)
	tripsheetH := handlers.NewTripsheetHandler(tripsheetSvc)
	tripsheetTripH := handlers.NewTripsheetTripHandler(tripsheetTripSvc)

	api := r.Group("/api/tripsheet")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("", tripsheetH.GetAll)
		protected.GET("/:id", tripsheetH.GetByID)
		protected.GET("/trips", tripsheetTripH.GetAll)
		protected.GET("/trips/:id", tripsheetTripH.GetByID)
		protected.GET("/trips/by-tripsheet/:tripsheet_id", tripsheetTripH.GetAllByTripsheetID)

		writeGroup := protected.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			writeGroup.POST("", tripsheetH.Create)
			writeGroup.PUT("/:id", tripsheetH.Update)
			writeGroup.DELETE("/:id", tripsheetH.Delete)
			writeGroup.POST("/trips", tripsheetTripH.Create)
			writeGroup.PUT("/trips/:id", tripsheetTripH.Update)
			writeGroup.DELETE("/trips/:id", tripsheetTripH.Delete)
		}
	}
}
