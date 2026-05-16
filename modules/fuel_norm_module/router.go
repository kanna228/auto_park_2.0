package fuel_norm_module

import (
	"database/sql"

	"auto_park/internal/config"
	"auto_park/middleware"
	"auto_park/modules/fuel_norm_module/handlers"
	"auto_park/modules/fuel_norm_module/repository"
	"auto_park/modules/fuel_norm_module/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	repo := repository.NewFuelNormRepository(db)
	svc := service.NewFuelNormService(repo)
	h := handlers.NewFuelNormHandler(svc)

	api := r.Group("/api/fuel/norms")
	api.Use(middleware.AuthJWT(cfg))
	{
		api.GET("", h.ListFuelNorms)
		api.GET("/:id", h.GetFuelNormByID)

		writeGroup := api.Group("")
		writeGroup.Use(middleware.RequireRoles(1, 2, 3))
		{
			writeGroup.POST("", h.CreateFuelNorm)
			writeGroup.PUT("/:id", h.UpdateFuelNorm)
		}

		deleteGroup := api.Group("")
		deleteGroup.Use(middleware.RequireRoles(1))
		{
			deleteGroup.DELETE("/:id", h.DeleteFuelNorm)
		}
	}
}
