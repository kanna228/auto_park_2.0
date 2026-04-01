package handlers

import (
	"auto_park/tripsheet_module/dto"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAll godoc
// @Summary      Get all tripsheets
// @Description  Returns all tripsheets with optional filters.
// @Tags         Tripsheets
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id query int false "Vehicle ID"
// @Param        driver_id query int false "Driver ID"
// @Param        date_from query string false "Date from (YYYY-MM-DD)"
// @Param        date_to query string false "Date to (YYYY-MM-DD)"
// @Param        status_id query int false "Status ID"
// @Param        tripsheet_number query string false "Tripsheet number"
// @Param        vehicle_plate_number query string false "Vehicle plate number"
// @Param        vehicle_brand query string false "Vehicle brand"
// @Param        driver_last_name query string false "Driver last name"
// @Param        driver_first_name query string false "Driver first name"
// @Param        driver_middle_name query string false "Driver middle name"
// @Success      200 {object} TripsheetListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /api/tripsheet [get]
func (h *TripsheetHandler) GetAll(c *gin.Context) {
	var filter dto.TripsheetFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	data, total, err := h.svc.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   data,
		"total":   total,
	})
}

// GetByID godoc
// @Summary      Get tripsheet by ID
// @Description  Returns a single tripsheet by its ID.
// @Tags         Tripsheets
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tripsheet ID"
// @Success      200 {object} TripsheetResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/tripsheet/{id} [get]
func (h *TripsheetHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	data, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "tripsheet not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
