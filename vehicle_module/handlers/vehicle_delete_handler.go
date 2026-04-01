package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// DELETE /api/vehicles/:id

// DeleteVehicle godoc
// @Summary      Delete vehicle
// @Description  Deletes vehicle by id.
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	ok, err := h.svc.DeleteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle not found"})
		return
	}

	// можно 200 с JSON (как у тебя стиль), либо 204 No Content
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}
