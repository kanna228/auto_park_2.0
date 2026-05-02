package warehouse_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/warehouse_module/handlers"
	"auto_park/modules/warehouse_module/repository"
	"auto_park/modules/warehouse_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	partRepo := repository.NewPartRepository(db)
	partSvc := service.NewPartService(partRepo)
	partHandler := handlers.NewPartHandler(partSvc)

	api := r.Group("/api/warehouse")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		readGroup := protected.Group("/parts")
		readGroup.Use(middleware.RequireRoles(1, 5))
		{
			readGroup.GET("", partHandler.ListParts)
			readGroup.GET("/:id", partHandler.GetPartByID)
		}

		writeGroup := protected.Group("/parts")
		writeGroup.Use(middleware.RequireRoles(5))
		{
			writeGroup.POST("", partHandler.CreatePart)
			writeGroup.PUT("/:id", partHandler.UpdatePart)
		}

		deleteGroup := protected.Group("/parts")
		deleteGroup.Use(middleware.RequireRoles(1))
		{
			deleteGroup.DELETE("/:id", partHandler.DeletePart)
		}
	}
}
