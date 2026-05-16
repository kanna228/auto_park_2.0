package maintenance_schedule_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/maintenance_schedule_module/handlers"
	"auto_park/modules/maintenance_schedule_module/repository"
	"auto_park/modules/maintenance_schedule_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	repo := repository.NewMaintenanceScheduleRepository(db)
	svc := service.NewMaintenanceScheduleService(repo)
	h := handlers.NewMaintenanceScheduleHandler(svc)

	api := r.Group("/api/maintenance-schedules")
	api.Use(middleware.AuthJWT(cfg))
	{
		readGroup := api.Group("")
		readGroup.Use(middleware.RequireRoles(1, 3, 4))
		{
			readGroup.GET("", h.ListMaintenanceSchedules)
			readGroup.GET("/:id", h.GetMaintenanceScheduleByID)
		}

		writeGroup := api.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 3))
		{
			writeGroup.POST("", h.CreateMaintenanceSchedule)
			writeGroup.PUT("/:id", h.UpdateMaintenanceSchedule)
			writeGroup.DELETE("/:id", h.DeleteMaintenanceSchedule)
		}
	}
}
