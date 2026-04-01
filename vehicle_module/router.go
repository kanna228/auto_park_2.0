package vehicle_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/vehicle_module/handlers"
	"auto_park/vehicle_module/repository"
	"auto_park/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	vehicleRepo := repository.NewVehicleRepository(db)
	vehicleSvc := service.NewVehicleService(vehicleRepo)
	vehicleHandler := handlers.NewVehicleHandler(vehicleSvc)

	tireRepo := repository.NewTireRepository(db)
	tireSvc := service.NewTireService(tireRepo)
	tireHandler := handlers.NewTireHandler(tireSvc)

	tirePlaceRepo := repository.NewTirePlaceRepository(db)
	tirePlaceSvc := service.NewTirePlaceService(tirePlaceRepo)
	tirePlaceHandler := handlers.NewTirePlaceHandler(tirePlaceSvc)

	api := r.Group("/api/vehicles")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.POST("", vehicleHandler.CreateVehicle)
		protected.GET("/:id", vehicleHandler.GetVehicleByID)
		protected.GET("", vehicleHandler.ListVehicles)
		protected.PUT("/:id", vehicleHandler.UpdateVehicle)
		protected.DELETE("/:id", vehicleHandler.DeleteVehicle)

		protected.POST("/tires", tireHandler.CreateTire)
		protected.GET("/tires/:id", tireHandler.GetTireByID)
		protected.GET("/tires", tireHandler.ListTires)
		protected.PUT("/tires/:id", tireHandler.UpdateTire)
		protected.DELETE("/tires/:id", tireHandler.DeleteTire)

		protected.POST("/tire-places", tirePlaceHandler.CreateTirePlace)
		protected.GET("/tire-places/:id", tirePlaceHandler.GetTirePlaceByID)
		protected.GET("/tire-places", tirePlaceHandler.ListTirePlaces)
		protected.PUT("/tire-places/:id", tirePlaceHandler.UpdateTirePlace)
		protected.DELETE("/tire-places/:id", tirePlaceHandler.DeleteTirePlace)
		protected.GET("/vehicle-tires/:vehicle_id", tireHandler.GetTiresByVehicleID)
	}
}
