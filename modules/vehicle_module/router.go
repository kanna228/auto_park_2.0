package vehicle_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/vehicle_module/handlers"
	"auto_park/modules/vehicle_module/repository"
	"auto_park/modules/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	vehicleRepo := repository.NewVehicleRepository(db)
	vehiclePhotoStorage := service.NewVehiclePhotoStorage(cfg.Uploads.RootDir)

	insuranceRepo := repository.NewInsuranceRepository(db)
	technicalInspectionRepo := repository.NewTechnicalInspectionRepository(db)

	vehicleSvc := service.NewVehicleService(
		vehicleRepo,
		vehiclePhotoStorage,
		insuranceRepo,
		technicalInspectionRepo,
	)
	vehicleHandler := handlers.NewVehicleHandler(vehicleSvc)

	tireRepo := repository.NewTireRepository(db)
	tireSvc := service.NewTireService(tireRepo)
	tireHandler := handlers.NewTireHandler(tireSvc)

	tirePlaceRepo := repository.NewTirePlaceRepository(db)
	tirePlaceSvc := service.NewTirePlaceService(tirePlaceRepo)
	tirePlaceHandler := handlers.NewTirePlaceHandler(tirePlaceSvc)

	documentStorage := service.NewVehicleDocumentStorage(cfg.Uploads.RootDir)

	insuranceSvc := service.NewInsuranceService(insuranceRepo, documentStorage)
	insuranceHandler := handlers.NewInsuranceHandler(insuranceSvc)

	technicalInspectionSvc := service.NewTechnicalInspectionService(technicalInspectionRepo, documentStorage)
	technicalInspectionHandler := handlers.NewTechnicalInspectionHandler(technicalInspectionSvc)

	api := r.Group("/api/vehicles")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.POST("", vehicleHandler.CreateVehicle)
		protected.GET("/:id", vehicleHandler.GetVehicleByID)
		protected.GET("", vehicleHandler.ListVehicles)
		protected.PUT("/:id", vehicleHandler.UpdateVehicle)
		protected.DELETE("/:id", vehicleHandler.DeleteVehicle)
		protected.GET("/statuses", vehicleHandler.ListVehicleStatuses)

		protected.POST("/:id/photo", vehicleHandler.UploadVehiclePhoto)
		protected.PUT("/:id/photo", vehicleHandler.UpdateVehiclePhoto)
		protected.DELETE("/:id/photo", vehicleHandler.DeleteVehiclePhoto)

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

		protected.POST("/insurance", insuranceHandler.CreateInsurance)
		protected.GET("/insurance/:id", insuranceHandler.GetInsuranceByID)
		protected.GET("/insurance", insuranceHandler.ListInsurance)
		protected.PUT("/insurance/:id", insuranceHandler.UpdateInsurance)
		protected.DELETE("/insurance/:id", insuranceHandler.DeleteInsurance)
		protected.POST("/insurance/:id/file", insuranceHandler.UploadInsuranceFile)
		protected.DELETE("/insurance/:id/file", insuranceHandler.DeleteInsuranceFile)

		protected.POST("/technical-inspections", technicalInspectionHandler.CreateTechnicalInspection)
		protected.GET("/technical-inspections/:id", technicalInspectionHandler.GetTechnicalInspectionByID)
		protected.GET("/technical-inspections", technicalInspectionHandler.ListTechnicalInspections)
		protected.PUT("/technical-inspections/:id", technicalInspectionHandler.UpdateTechnicalInspection)
		protected.DELETE("/technical-inspections/:id", technicalInspectionHandler.DeleteTechnicalInspection)
		protected.POST("/technical-inspections/:id/file", technicalInspectionHandler.UploadTechnicalInspectionFile)
		protected.DELETE("/technical-inspections/:id/file", technicalInspectionHandler.DeleteTechnicalInspectionFile)
	}
}
