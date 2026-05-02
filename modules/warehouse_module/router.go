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

const (
	roleAdmin            int64 = 1
	roleDutyMechanic     int64 = 4
	roleWarehouseManager int64 = 5
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config, db *sql.DB) {
	partRepo := repository.NewPartRepository(db)
	partSvc := service.NewPartService(partRepo)
	partHandler := handlers.NewPartHandler(partSvc)

	partRequestRepo := repository.NewPartRequestRepository(db)
	partRequestSvc := service.NewPartRequestService(partRequestRepo)
	partRequestHandler := handlers.NewPartRequestHandler(partRequestSvc)

	api := r.Group("/api/warehouse")
	protected := api.Group("")
	protected.Use(middleware.AuthJWT(cfg))
	{
		readGroup := protected.Group("/parts")
		readGroup.Use(middleware.RequireRoles(roleAdmin, roleWarehouseManager))
		{
			readGroup.GET("", partHandler.ListParts)
			readGroup.GET("/:id", partHandler.GetPartByID)
		}

		writeGroup := protected.Group("/parts")
		writeGroup.Use(middleware.RequireRoles(roleWarehouseManager))
		{
			writeGroup.POST("", partHandler.CreatePart)
			writeGroup.PUT("/:id", partHandler.UpdatePart)
		}

		deleteGroup := protected.Group("/parts")
		deleteGroup.Use(middleware.RequireRoles(roleAdmin))
		{
			deleteGroup.DELETE("/:id", partHandler.DeletePart)
		}

		partRequestReadGroup := protected.Group("")
		partRequestReadGroup.Use(middleware.RequireRoles(roleAdmin, roleDutyMechanic, roleWarehouseManager))
		{
			partRequestReadGroup.GET("/part-request-statuses", partRequestHandler.ListPartRequestStatuses)
			partRequestReadGroup.GET("/part-requests", partRequestHandler.ListPartRequests)
			partRequestReadGroup.GET("/part-requests/:id", partRequestHandler.GetPartRequestByID)
		}

		partRequestCreateGroup := protected.Group("/part-requests")
		partRequestCreateGroup.Use(middleware.RequireRoles(roleDutyMechanic))
		{
			partRequestCreateGroup.POST("", partRequestHandler.CreatePartRequest)
		}

		partRequestUpdateGroup := protected.Group("/part-requests")
		partRequestUpdateGroup.Use(middleware.RequireRoles(roleAdmin, roleDutyMechanic, roleWarehouseManager))
		{
			partRequestUpdateGroup.PUT("/:id", partRequestHandler.UpdatePartRequest)
			partRequestUpdateGroup.DELETE("/:id", partRequestHandler.DeletePartRequest)
		}

		partRequestStatusGroup := protected.Group("/part-requests")
		partRequestStatusGroup.Use(middleware.RequireRoles(roleAdmin, roleWarehouseManager))
		{
			partRequestStatusGroup.PATCH("/:id/status", partRequestHandler.UpdatePartRequestStatus)
		}
	}
}
