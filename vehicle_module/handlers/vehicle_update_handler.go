package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/vehicle_module/dto"

	"github.com/gin-gonic/gin"
)

// UpdateVehicle godoc
// @Summary      Update vehicle
// @Description  Updates vehicle by id. PUT updates all fields.
// @Tags         Vehicles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Vehicle ID"
// @Param        payload body dto.VehicleUpdateRequest true "Vehicle update payload"
// @Success      200 {object} UpdateVehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/{id} [put]
func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.VehicleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body (check required fields and date format)",
		})
		return
	}

	ok, err := h.svc.UpdateByID(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}
