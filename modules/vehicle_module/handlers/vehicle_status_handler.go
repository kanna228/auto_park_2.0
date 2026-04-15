package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListVehicleStatuses godoc
// @Summary      List vehicle statuses
// @Description  Returns all available vehicle statuses.
// @Tags         Vehicles
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} VehicleStatusListResponseWrap
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicle-statuses [get]
func (h *VehicleHandler) ListVehicleStatuses(c *gin.Context) {
	resp, err := h.svc.ListVehicleStatuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
