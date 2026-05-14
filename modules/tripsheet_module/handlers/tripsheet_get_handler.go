package handlers

import (
	"auto_park/modules/tripsheet_module/dto"
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
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Param        sort_by query string false "Sort by: tripsheet_date, created_at, updated_at, id, vehicle_id, driver_id, status_id"
// @Param        order query string false "Sort order: asc or desc"
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

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   data,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
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
