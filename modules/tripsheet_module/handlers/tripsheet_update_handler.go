package handlers

import (
	"auto_park/modules/tripsheet_module/dto"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary      Update tripsheet
// @Description  Updates an existing tripsheet by ID.
// @Tags         Tripsheets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tripsheet ID"
// @Param        request body dto.UpdateTripsheetRequest true "Tripsheet update request"
// @Success      200 {object} TripsheetCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/tripsheet/{id} [put]
func (h *TripsheetHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	var req dto.UpdateTripsheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "tripsheet not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
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
