package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"auto_park/middleware"
	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/service"

	"github.com/gin-gonic/gin"
)

type TripsheetTripHandler struct {
	svc service.TripsheetTripService
}

func NewTripsheetTripHandler(svc service.TripsheetTripService) *TripsheetTripHandler {
	return &TripsheetTripHandler{
		svc: svc,
	}
}

// Create godoc
// @Summary      Create tripsheet trip
// @Description  Creates a new trip inside a tripsheet.
// @Tags         Tripsheet Trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateTripsheetTripRequest true "Tripsheet trip create request"
// @Success      201 {object} TripsheetTripResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/tripsheet/trips [post]
func (h *TripsheetTripHandler) Create(c *gin.Context) {
	var req dto.CreateTripsheetTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    resp,
	})
}

// Update godoc
// @Summary      Update tripsheet trip
// @Description  Updates a trip inside a tripsheet by ID.
// @Tags         Tripsheet Trips
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tripsheet Trip ID"
// @Param        request body dto.UpdateTripsheetTripRequest true "Tripsheet trip update request"
// @Success      200 {object} TripsheetTripResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/tripsheet/trips/{id} [put]
func (h *TripsheetTripHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	var req dto.UpdateTripsheetTripRequest
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
				"error":   "tripsheet trip not found",
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

// Delete godoc
// @Summary      Delete tripsheet trip
// @Description  Deletes a tripsheet trip by ID.
// @Tags         Tripsheet Trips
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tripsheet Trip ID"
// @Success      200 {object} SuccessMessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/tripsheet/trips/{id} [delete]
func (h *TripsheetTripHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	err = h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "tripsheet trip not found",
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
		"message": "tripsheet trip deleted successfully",
	})
}

// GetByID godoc
// @Summary      Get tripsheet trip by ID
// @Description  Returns a single tripsheet trip by its ID.
// @Tags         Tripsheet Trips
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tripsheet Trip ID"
// @Success      200 {object} TripsheetTripResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/tripsheet/trips/{id} [get]
func (h *TripsheetTripHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "tripsheet trip not found",
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
		"data":    resp,
	})
}

// GetAll godoc
// @Summary      Get all tripsheet trips
// @Description  Returns all tripsheet trips with optional filters.
// @Tags         Tripsheet Trips
// @Produce      json
// @Security     BearerAuth
// @Param        tripsheet_id query int false "Tripsheet ID"
// @Param        vehicle_id query int false "Vehicle ID"
// @Param        driver_id query int false "Driver ID"
// @Param        status_id query int false "Status ID"
// @Param        date_from query string false "Date from (YYYY-MM-DD)"
// @Param        date_to query string false "Date to (YYYY-MM-DD)"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Param        sort_by query string false "Sort by: id, tripsheet_id, vehicle_id, driver_id, status_id, start_time, created_at"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} TripsheetTripListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /api/tripsheet/trips [get]
func (h *TripsheetTripHandler) GetAll(c *gin.Context) {
	var filter dto.TripsheetTripFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if middleware.CurrentAccountType(c) == "driver" {
		driverID := middleware.CurrentDriverID(c)
		if driverID <= 0 {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "driver account is not linked"})
			return
		}
		filter.DriverID = &driverID
	}

	items, total, err := h.svc.GetAll(c.Request.Context(), filter)
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
		"items":   items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetAllByTripsheetID godoc
// @Summary      Get all trips by tripsheet ID
// @Description  Returns all trips for a specific tripsheet with optional filters.
// @Tags         Tripsheet Trips
// @Produce      json
// @Security     BearerAuth
// @Param        tripsheet_id path int true "Tripsheet ID"
// @Param        status_id query int false "Status ID"
// @Param        date_from query string false "Date from (YYYY-MM-DD)"
// @Param        date_to query string false "Date to (YYYY-MM-DD)"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Param        sort_by query string false "Sort by: id, tripsheet_id, vehicle_id, driver_id, status_id, start_time, created_at"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} TripsheetTripListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /api/tripsheet/trips/by-tripsheet/{tripsheet_id} [get]
func (h *TripsheetTripHandler) GetAllByTripsheetID(c *gin.Context) {
	tripsheetID, err := strconv.ParseInt(c.Param("tripsheet_id"), 10, 64)
	if err != nil || tripsheetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid tripsheet_id",
		})
		return
	}

	var filter dto.TripsheetTripFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if middleware.CurrentAccountType(c) == "driver" {
		driverID := middleware.CurrentDriverID(c)
		if driverID <= 0 {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "driver account is not linked"})
			return
		}
		filter.DriverID = &driverID
	}

	items, total, err := h.svc.GetAllByTripsheetID(c.Request.Context(), tripsheetID, filter)
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
		"items":   items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
