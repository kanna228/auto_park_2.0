package maintenance_execution_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/maintenance_execution_module/handlers"
	"auto_park/modules/maintenance_execution_module/repository"
	"auto_park/modules/maintenance_execution_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	repo := repository.NewMaintenanceExecutionRepository(db)
	svc := service.NewMaintenanceExecutionService(repo)
	h := handlers.NewMaintenanceExecutionHandler(svc)

	api := r.Group("/api/maintenance-executions")
	api.Use(middleware.AuthJWT(cfg))
	{
		readGroup := api.Group("")
		readGroup.Use(middleware.RequireRoles(1, 3, 4))
		{
			readGroup.GET("", h.ListMaintenanceExecutions)
			readGroup.GET("/by-schedule/:schedule_id", h.ListMaintenanceExecutionsBySchedule)
			readGroup.GET("/:id", h.GetMaintenanceExecutionByID)
		}

		writeGroup := api.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 4))
		{
			writeGroup.POST("", h.CreateMaintenanceExecution)
			writeGroup.POST("/bulk", h.BulkCreateMaintenanceExecutions)
			writeGroup.PUT("/:id", h.UpdateMaintenanceExecution)
		}

		deleteGroup := api.Group("")
		deleteGroup.Use(middleware.RequireRoles(1))
		{
			deleteGroup.DELETE("/:id", h.DeleteMaintenanceExecution)
		}
	}
}
