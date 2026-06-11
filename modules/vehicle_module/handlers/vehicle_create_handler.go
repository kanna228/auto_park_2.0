package handlers

import (
	"net/http"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	svc      service.VehicleService
	auditSvc *auditlogservice.Service
}

func NewVehicleHandler(svc service.VehicleService, auditSvc *auditlogservice.Service) *VehicleHandler {
	return &VehicleHandler{svc: svc, auditSvc: auditSvc}
}

// POST /api/vehicles

// CreateVehicle godoc
// @Summary      Create vehicle
// @Description  Creates a new vehicle.
// @Tags         Vehicles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.VehicleCreateRequest true "Vehicle create payload"
// @Success      201 {object} CreateVehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles [post]
func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	var req dto.VehicleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body (check required fields and date format)",
		})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"success",
		"vehicle",
		"",
		"created",
		auditlog.Actor(middleware.CurrentEmail(c), 0),
		auditlog.Message("id", id, "board", req.BoardNumber, "plate", req.StateNumber, "brand_model", req.BrandModel),
	)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": dto.VehicleCreateResponse{
			ID: id,
		},
	})
}
