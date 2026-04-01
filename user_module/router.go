package user_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/user_module/handlers"
	"auto_park/user_module/repository"
	"auto_park/user_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	userRepo := repository.NewUserRepo(db, cfg.Schemas.User)
	driverRepo := repository.NewDriverRepo(db, cfg.Schemas.User)

	mailer := service.NewSMTPMailer(cfg)

	authSvc := service.NewAuthService(cfg, userRepo)
	userSvc := service.NewUserService(cfg, userRepo, mailer)
	usersReadSvc := service.NewUsersReadService(userRepo)
	updSvc := service.NewUsersUpdateService(userRepo)
	delSvc := service.NewUsersDeleteService(userRepo)
	driversSvc := service.NewDriversService(driverRepo)

	authH := handlers.NewAuthHandler(cfg, authSvc)
	usersH := handlers.NewUsersHandler(cfg, userSvc)
	usersReadH := handlers.NewUsersReadHandler(usersReadSvc)
	updH := handlers.NewUsersUpdateHandler(updSvc)
	delH := handlers.NewUsersDeleteHandler(delSvc)
	driversH := handlers.NewDriversHandler(driversSvc)

	api := r.Group("/api/users")
	api.POST("/login", authH.Login)

	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		protected.GET("", usersReadH.ListUsers)
		protected.GET("/:id", usersReadH.GetUserByID)

		adminUsers := protected.Group("")
		adminUsers.Use(middleware.RequireRoles(1))
		{
			adminUsers.POST("", usersH.CreateUser)
			adminUsers.PUT("/:id", updH.UpdateUser)
			adminUsers.DELETE("/:id", delH.DeleteUser)
		}

		drivers := protected.Group("/drivers")
		drivers.GET("", driversH.List)
		drivers.GET("/:id", driversH.GetByID)

		adminOnly := drivers.Group("")
		adminOnly.Use(middleware.RequireRoles(1, 2, 3))
		{
			adminOnly.POST("", driversH.Create)
			adminOnly.PUT("/:id", driversH.Update)
			adminOnly.DELETE("/:id", driversH.Delete)
		}
	}
}
