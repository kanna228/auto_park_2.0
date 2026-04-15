package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/vehicle_module/dto"

	"github.com/gin-gonic/gin"
)

// UpdateTire godoc
// @Summary      Update tire
// @Description  Updates tire by id.
// @Tags         Tires
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tire ID"
// @Param        payload body dto.TireUpdateRequest true "Tire update payload"
// @Success      200 {object} TireUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tires/{id} [put]
func (h *TireHandler) UpdateTire(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.TireUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	ok, err := h.svc.UpdateByID(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}
