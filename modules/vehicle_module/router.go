package vehicle_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/vehicle_module/handlers"
	"auto_park/modules/vehicle_module/repository"
	"auto_park/modules/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB, auditSvc *auditlogservice.Service) {
	vehicleRepo := repository.NewVehicleRepository(db)
	vehiclePhotoStorage := service.NewVehiclePhotoStorage(cfg.Uploads.RootDir)

	insuranceRepo := repository.NewInsuranceRepository(db)
	technicalInspectionRepo := repository.NewTechnicalInspectionRepository(db)
	vehicleDocumentRepo := repository.NewVehicleDocumentRepository(db)

	vehicleSvc := service.NewVehicleService(
		vehicleRepo,
		vehiclePhotoStorage,
		insuranceRepo,
		technicalInspectionRepo,
	)
	vehicleHandler := handlers.NewVehicleHandler(vehicleSvc, auditSvc)

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
	vehicleDocumentSvc := service.NewVehicleDocumentService(vehicleDocumentRepo)
	vehicleDocumentHandler := handlers.NewVehicleDocumentHandler(vehicleDocumentSvc)

	api := r.Group("/api/vehicles")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("", vehicleHandler.ListVehicles)

		// Static routes must stay before /:id routes.
		protected.GET("/statuses", vehicleHandler.ListVehicleStatuses)
		protected.GET("/status-history", vehicleHandler.ListVehicleStatusHistory)
		protected.GET("/status-history/:id", vehicleHandler.GetVehicleStatusHistoryByID)

		protected.GET("/:id/tires", tireHandler.ListVehicleTires)
		protected.GET("/:id/documents", vehicleDocumentHandler.List)

		protected.GET("/:id", vehicleHandler.GetVehicleByID)

		protected.GET("/tires", tireHandler.ListTires)
		protected.GET("/tires/:id", tireHandler.GetTireByID)

		protected.GET("/tire-places", tirePlaceHandler.ListTirePlaces)
		protected.GET("/tire-places/:id", tirePlaceHandler.GetTirePlaceByID)
		protected.GET("/vehicle-tires/:vehicle_id", tireHandler.GetTiresByVehicleID)

		protected.GET("/insurance", insuranceHandler.ListInsurance)
		protected.GET("/insurance/:id", insuranceHandler.GetInsuranceByID)

		protected.GET("/technical-inspections", technicalInspectionHandler.ListTechnicalInspections)
		protected.GET("/technical-inspections/:id", technicalInspectionHandler.GetTechnicalInspectionByID)

		writeGroup := protected.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			writeGroup.POST("", vehicleHandler.CreateVehicle)

			writeGroup.POST("/tires", tireHandler.CreateTire)
			writeGroup.PUT("/tires/:id", tireHandler.UpdateTire)
			writeGroup.DELETE("/tires/:id", tireHandler.DeleteTire)

			writeGroup.POST("/tire-places", tirePlaceHandler.CreateTirePlace)
			writeGroup.PUT("/tire-places/:id", tirePlaceHandler.UpdateTirePlace)
			writeGroup.DELETE("/tire-places/:id", tirePlaceHandler.DeleteTirePlace)

			writeGroup.POST("/insurance", insuranceHandler.CreateInsurance)
			writeGroup.PUT("/insurance/:id", insuranceHandler.UpdateInsurance)
			writeGroup.DELETE("/insurance/:id", insuranceHandler.DeleteInsurance)
			writeGroup.POST("/insurance/:id/file", insuranceHandler.UploadInsuranceFile)
			writeGroup.DELETE("/insurance/:id/file", insuranceHandler.DeleteInsuranceFile)

			writeGroup.POST("/technical-inspections", technicalInspectionHandler.CreateTechnicalInspection)
			writeGroup.PUT("/technical-inspections/:id", technicalInspectionHandler.UpdateTechnicalInspection)
			writeGroup.DELETE("/technical-inspections/:id", technicalInspectionHandler.DeleteTechnicalInspection)
			writeGroup.POST("/technical-inspections/:id/file", technicalInspectionHandler.UploadTechnicalInspectionFile)
			writeGroup.DELETE("/technical-inspections/:id/file", technicalInspectionHandler.DeleteTechnicalInspectionFile)

			writeGroup.PUT("/:id", vehicleHandler.UpdateVehicle)
			writeGroup.DELETE("/:id", vehicleHandler.DeleteVehicle)

			writeGroup.POST("/:id/photo", vehicleHandler.UploadVehiclePhoto)
			writeGroup.PUT("/:id/photo", vehicleHandler.UpdateVehiclePhoto)
			writeGroup.DELETE("/:id/photo", vehicleHandler.DeleteVehiclePhoto)

			writeGroup.POST("/:id/tires", tireHandler.CreateVehicleTire)
			writeGroup.PUT("/:id/tires/:tid", tireHandler.UpdateVehicleTire)
			writeGroup.DELETE("/:id/tires/:tid", tireHandler.DeleteVehicleTire)

			writeGroup.POST("/:id/documents", vehicleDocumentHandler.Create)
			writeGroup.DELETE("/:id/documents/:doc_id", vehicleDocumentHandler.Delete)
		}
	}
}
