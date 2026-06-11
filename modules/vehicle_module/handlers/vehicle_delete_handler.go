package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	"auto_park/modules/vehicle_module/dto"

	"github.com/gin-gonic/gin"
)

// DELETE /api/vehicles/:id

// DeleteVehicle godoc
// @Summary      Delete vehicle
// @Description  Deletes vehicle by id, unassigns all tires from it, and cascades deletion of related tripsheets and tripsheet trips.
// @Tags         Vehicles
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle ID"
// @Success      200 {object} DeleteVehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/{id} [delete]
func (h *VehicleHandler) DeleteVehicle(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	vehicle, _ := h.svc.GetByID(c.Request.Context(), id, true)

	ok, err := h.svc.DeleteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "vehicle not found",
		})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"warning",
		"vehicle",
		"active",
		"deleted",
		auditlog.Actor(middleware.CurrentEmail(c), 0),
		auditlog.Message(
			"id", id,
			"board", vehicleAuditBoard(vehicle),
			"plate", vehicleAuditPlate(vehicle),
			"brand_model", vehicleAuditBrand(vehicle),
			"vin", vehicleAuditVIN(vehicle),
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}

func vehicleAuditBoard(v *dto.VehicleResponse) string {
	if v == nil {
		return ""
	}
	return v.BoardNumber
}

func vehicleAuditPlate(v *dto.VehicleResponse) string {
	if v == nil {
		return ""
	}
	return v.StateNumber
}

func vehicleAuditBrand(v *dto.VehicleResponse) string {
	if v == nil {
		return ""
	}
	return v.BrandModel
}

func vehicleAuditVIN(v *dto.VehicleResponse) string {
	if v == nil {
		return ""
	}
	return v.VIN
}
