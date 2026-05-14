package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"auto_park/modules/fuel_module/dto"
	"auto_park/modules/fuel_module/service"

	"github.com/gin-gonic/gin"
)

type FuelRefillHandler struct {
	svc service.FuelRefillService
}

func NewFuelRefillHandler(svc service.FuelRefillService) *FuelRefillHandler {
	return &FuelRefillHandler{svc: svc}
}

// Create godoc
// @Summary      Create fuel refill
// @Description  Creates a new fuel refill record linked to a tripsheet and vehicle.
// @Tags         Fuel Refills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateFuelRefillRequest true "Fuel refill create request"
// @Success      201 {object} FuelRefillResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/fuel/refills [post]
func (h *FuelRefillHandler) Create(c *gin.Context) {
	var req dto.CreateFuelRefillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": resp})
}

// Update godoc
// @Summary      Update fuel refill
// @Description  Updates an existing fuel refill record by ID.
// @Tags         Fuel Refills
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Fuel refill ID"
// @Param        request body dto.UpdateFuelRefillRequest true "Fuel refill update request"
// @Success      200 {object} FuelRefillResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/fuel/refills/{id} [put]
func (h *FuelRefillHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.UpdateFuelRefillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "fuel refill not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// Delete godoc
// @Summary      Delete fuel refill
// @Description  Deletes a fuel refill record by ID.
// @Tags         Fuel Refills
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Fuel refill ID"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/fuel/refills/{id} [delete]
func (h *FuelRefillHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "fuel refill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "fuel refill deleted successfully"})
}

// GetByID godoc
// @Summary      Get fuel refill by ID
// @Description  Returns a fuel refill record by its ID.
// @Tags         Fuel Refills
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Fuel refill ID"
// @Success      200 {object} FuelRefillResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/fuel/refills/{id} [get]
func (h *FuelRefillHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "fuel refill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// GetAll godoc
// @Summary      Get all fuel refills
// @Description  Returns paginated fuel refill records with optional filters.
// @Tags         Fuel Refills
// @Produce      json
// @Security     BearerAuth
// @Param        tripsheet_id query int false "Tripsheet ID"
// @Param        vehicle_id query int false "Vehicle ID"
// @Param        driver_id query int false "Driver ID"
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to query string false "End date (YYYY-MM-DD)"
// @Param        page query int false "Page number"
// @Param        page_size query int false "Page size"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by: date, created_at, vehicle_id, driver_id, tripsheet_id"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} FuelRefillListResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/fuel/refills [get]
func (h *FuelRefillHandler) GetAll(c *gin.Context) {
	var filter dto.FuelRefillFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	items, total, err := h.svc.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"success": true, "items": items, "total": total, "limit": limit, "offset": offset})
}

// GetAllByTripsheetID godoc
// @Summary      Get fuel refills by tripsheet ID
// @Description  Returns all fuel refill records for a specific tripsheet.
// @Tags         Fuel Refills
// @Produce      json
// @Security     BearerAuth
// @Param        tripsheet_id path int true "Tripsheet ID"
// @Success      200 {object} FuelRefillListResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/fuel/refills/by-tripsheet/{tripsheet_id} [get]
func (h *FuelRefillHandler) GetAllByTripsheetID(c *gin.Context) {
	tripsheetID, err := strconv.ParseInt(c.Param("tripsheet_id"), 10, 64)
	if err != nil || tripsheetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid tripsheet_id"})
		return
	}

	var filter dto.FuelRefillFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	items, total, err := h.svc.GetAllByTripsheetID(c.Request.Context(), tripsheetID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"success": true, "items": items, "total": total, "limit": limit, "offset": offset})
}

// GetAllByVehicleID godoc
// @Summary      Get fuel refills by vehicle ID
// @Description  Returns all fuel refill records for a specific vehicle.
// @Tags         Fuel Refills
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id path int true "Vehicle ID"
// @Success      200 {object} FuelRefillListResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/fuel/refills/by-vehicle/{vehicle_id} [get]
func (h *FuelRefillHandler) GetAllByVehicleID(c *gin.Context) {
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil || vehicleID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid vehicle_id"})
		return
	}

	var filter dto.FuelRefillFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	items, total, err := h.svc.GetAllByVehicleID(c.Request.Context(), vehicleID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
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

	c.JSON(http.StatusOK, gin.H{"success": true, "items": items, "total": total, "limit": limit, "offset": offset})
}
