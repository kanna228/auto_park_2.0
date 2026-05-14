package handlers

import (
	"auto_park/modules/vehicle_module/dto"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetTireByID godoc
// @Summary      Get tire by ID
// @Description  Returns tire by id.
// @Tags         Tires
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tire ID"
// @Success      200 {object} TireResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tires/{id} [get]
func (h *TireHandler) GetTireByID(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// ListTires godoc
// @Summary      List tires
// @Description  Returns list of tires with filters, pagination and sorting.
// @Tags         Tires
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id query int false "Filter by vehicle ID"
// @Param        place_id   query int false "Filter by place ID"
// @Param        tire       query string false "Filter by tire name/code"
// @Param        limit      query int false "Limit" default(50)
// @Param        offset     query int false "Offset" default(0)
// @Param        sort_by    query string false "Sort by field"
// @Param        order      query string false "Sort order: asc or desc"
// @Success      200 {object} TireListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tires [get]
func (h *TireHandler) ListTires(c *gin.Context) {
	q := dto.TireListQuery{
		Tire:   c.Query("tire"),
		SortBy: c.Query("sort_by"),
		Order:  c.Query("order"),
	}

	if s := strings.TrimSpace(c.Query("vehicle_id")); s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid vehicle_id"})
			return
		}
		q.VehicleID = &n
	}

	if s := strings.TrimSpace(c.Query("place_id")); s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid place_id"})
			return
		}
		q.PlaceID = &n
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// GetTiresByVehicleID godoc
// @Summary      Get tires by vehicle ID
// @Description  Returns all tires attached to the specified vehicle.
// @Tags         Tires
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id path int true "Vehicle ID"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} TireListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/vehicle-tires/{vehicle_id} [get]
func (h *TireHandler) GetTiresByVehicleID(c *gin.Context) {
	vehicleIDStr := strings.TrimSpace(c.Param("vehicle_id"))
	vehicleID, err := strconv.ParseInt(vehicleIDStr, 10, 64)
	if err != nil || vehicleID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid vehicle_id",
		})
		return
	}

	limit := 50
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid limit"})
			return
		}
		limit = n
	}

	offset := 0
	if s := strings.TrimSpace(c.Query("offset")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid offset"})
			return
		}
		offset = n
	}

	resp, err := h.svc.GetByVehicleID(c.Request.Context(), vehicleID, limit, offset)
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
