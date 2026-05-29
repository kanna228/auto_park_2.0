package user_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/user_module/handlers"
	"auto_park/modules/user_module/repository"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	userRepo := repository.NewUserRepo(db, cfg.Schemas.User)
	driverRepo := repository.NewDriverRepo(db, cfg.Schemas.User)
	mechanicShiftRepo := repository.NewMechanicShiftRepo(db)
	driverShiftRepo := repository.NewDriverShiftRepo(db)

	mailer := service.NewSMTPMailer(cfg)
	driverPhotoStorage := service.NewDriverPhotoStorage(cfg.Uploads.RootDir)

	authSvc := service.NewAuthService(cfg, userRepo)
	userSvc := service.NewUserService(cfg, userRepo, mailer)
	usersReadSvc := service.NewUsersReadService(userRepo)
	updSvc := service.NewUsersUpdateService(userRepo)
	delSvc := service.NewUsersDeleteService(userRepo)
	driversSvc := service.NewDriversService(driverRepo, driverPhotoStorage, mailer)
	mechanicShiftSvc := service.NewMechanicShiftService(mechanicShiftRepo)
	driverShiftSvc := service.NewDriverShiftService(driverShiftRepo)

	authH := handlers.NewAuthHandler(cfg, authSvc)
	usersH := handlers.NewUsersHandler(cfg, userSvc)
	usersReadH := handlers.NewUsersReadHandler(usersReadSvc)
	updH := handlers.NewUsersUpdateHandler(updSvc)
	delH := handlers.NewUsersDeleteHandler(delSvc)
	driversH := handlers.NewDriversHandler(driversSvc)
	mechanicShiftH := handlers.NewMechanicShiftHandler(mechanicShiftSvc)
	driverShiftH := handlers.NewDriverShiftHandler(driverShiftSvc)

	api := r.Group("/api/users")
	api.POST("/login", authH.Login)

	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("/me", usersReadH.Me)
		protected.GET("", usersReadH.ListUsers)

		adminUsersCreate := protected.Group("")
		adminUsersCreate.Use(middleware.RequireRoles(1))
		{
			adminUsersCreate.POST("", usersH.CreateUser)
		}

		mechanicShifts := protected.Group("/mechanic-shifts")
		mechanicShifts.Use(middleware.RequireRoles(1, 2, 3, 4))
		{
			mechanicShifts.POST("", mechanicShiftH.Create)
			mechanicShifts.GET("", mechanicShiftH.List)
			mechanicShifts.GET("/:id", mechanicShiftH.GetByID)
			mechanicShifts.PUT("/:id", mechanicShiftH.Update)
			mechanicShifts.PATCH("/:id/activity", mechanicShiftH.UpdateActivity)
			mechanicShifts.DELETE("/:id", mechanicShiftH.Delete)
		}

		driverShifts := protected.Group("/driver-shifts")
		driverShifts.Use(middleware.RequireRoles(1, 2, 3))
		{
			driverShifts.POST("", driverShiftH.Create)
			driverShifts.GET("", driverShiftH.List)
			driverShifts.GET("/:id", driverShiftH.GetByID)
			driverShifts.PUT("/:id", driverShiftH.Update)
			driverShifts.PATCH("/:id/activity", driverShiftH.UpdateActivity)
			driverShifts.DELETE("/:id", driverShiftH.Delete)
		}

		protected.GET("/driver-statuses", driversH.ListStatuses)

		drivers := protected.Group("/drivers")
		drivers.GET("", driversH.List)
		drivers.GET("/:id", driversH.GetByID)

		adminOnlyDrivers := drivers.Group("")
		adminOnlyDrivers.Use(middleware.RequireRoles(1, 2, 3))
		{
			adminOnlyDrivers.POST("", driversH.Create)
			adminOnlyDrivers.PUT("/:id", driversH.Update)
			adminOnlyDrivers.PATCH("/:id/status", driversH.UpdateStatus)
			adminOnlyDrivers.DELETE("/:id", driversH.Delete)
			adminOnlyDrivers.POST("/:id/vehicles", driversH.AssignVehicle)
			adminOnlyDrivers.DELETE("/:id/vehicles/:vehicle_id", driversH.UnassignVehicle)

			adminOnlyDrivers.POST("/:id/photo", driversH.UploadPhoto)
			adminOnlyDrivers.PUT("/:id/photo", driversH.UploadPhoto)
			adminOnlyDrivers.DELETE("/:id/photo", driversH.DeletePhoto)
		}

		protected.GET("/:id", usersReadH.GetUserByID)

		adminUsersByID := protected.Group("")
		adminUsersByID.Use(middleware.RequireRoles(1))
		{
			adminUsersByID.PUT("/:id", updH.UpdateUser)
			adminUsersByID.DELETE("/:id", delH.DeleteUser)
		}
	}

	driversAPI := r.Group("/api/drivers")
	driversAPI.Use(middleware.AuthJWT(cfg))
	{
		driversAPI.GET("", driversH.List)
		driversAPI.GET("/:id", driversH.GetByID)

		driversWriteAPI := driversAPI.Group("")
		driversWriteAPI.Use(middleware.RequireRoles(1, 2, 3))
		{
			driversWriteAPI.POST("", driversH.Create)
			driversWriteAPI.PUT("/:id", driversH.Update)
			driversWriteAPI.PATCH("/:id/status", driversH.UpdateStatus)
			driversWriteAPI.DELETE("/:id", driversH.Delete)
		}
	}
}
