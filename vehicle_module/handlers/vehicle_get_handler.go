package handlers

import (
	"auto_park/vehicle_module/dto"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GET /api/vehicles/:id

// GetVehicleByID godoc
// @Summary      Get vehicle by ID
// @Description  Returns vehicle by id.
// @Tags         Vehicles
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle ID"
// @Success      200 {object} VehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/{id} [get]
func (h *VehicleHandler) GetVehicleByID(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	v, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "vehicle not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    v,
	})
}

// GET /api/vehicles (list + filters)

// ListVehicles godoc
// @Summary      List vehicles
// @Description  List vehicles with filters, pagination and sorting.
// @Tags         Vehicles
// @Produce      json
// @Security     BearerAuth
// @Param        board_number query string false "Filter by board number"
// @Param        state_number query string false "Filter by state number"
// @Param        vin          query string false "Filter by VIN"
// @Param        brand_model  query string false "Filter by brand/model"
// @Param        year_from    query int    false "Manufacture year from"
// @Param        year_to      query int    false "Manufacture year to"
// @Param        driver_id    query int    false "Filter by driver id (WHERE driver_id = ANY(drivers_ids))"
// @Param        limit        query int    false "Limit"  default(50)
// @Param        offset       query int    false "Offset" default(0)
// @Param        sort_by      query string false "Sort by: id, board_number, state_number, manufacture_year, mileage, created_at"
// @Param        order        query string false "Order: asc or desc"
// @Success      200 {object} VehicleListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles [get]
func (h *VehicleHandler) ListVehicles(c *gin.Context) {
	q := dto.VehicleListQuery{
		BoardNumber: c.Query("board_number"),
		StateNumber: c.Query("state_number"),
		VIN:         c.Query("vin"),
		BrandModel:  c.Query("brand_model"),

		SortBy: c.Query("sort_by"),
		Order:  c.Query("order"),
	}

	// year_from, year_to
	if s := strings.TrimSpace(c.Query("year_from")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			q.ManufactureYearFrom = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid year_from"})
			return
		}
	}
	if s := strings.TrimSpace(c.Query("year_to")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			q.ManufactureYearTo = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid year_to"})
			return
		}
	}

	// driver_id
	if s := strings.TrimSpace(c.Query("driver_id")); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
			q.DriverID = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid driver_id"})
			return
		}
	}

	// limit / offset
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid limit"})
			return
		}
		q.Limit = n
	} else {
		q.Limit = 50
	}

	if s := strings.TrimSpace(c.Query("offset")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid offset"})
			return
		}
		q.Offset = n
	} else {
		q.Offset = 0
	}

	resp, err := h.svc.List(c.Request.Context(), q)
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
