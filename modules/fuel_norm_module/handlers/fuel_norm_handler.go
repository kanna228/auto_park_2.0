package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/fuel_norm_module/dto"
	"auto_park/modules/fuel_norm_module/repository"
	"auto_park/modules/fuel_norm_module/service"

	"github.com/gin-gonic/gin"
)

type FuelNormHandler struct {
	svc service.FuelNormService
}

func NewFuelNormHandler(svc service.FuelNormService) *FuelNormHandler {
	return &FuelNormHandler{svc: svc}
}

// ListFuelNorms godoc
// @Summary      List fuel norms
// @Description  Returns paginated fuel consumption norms.
// @Tags         Fuel Norms
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} FuelNormListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/fuel/norms [get]
func (h *FuelNormHandler) ListFuelNorms(c *gin.Context) {
	limit, ok := parseIntQuery(c, "limit", false)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}

	resp, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// GetFuelNormByID godoc
// @Summary      Get fuel norm by ID
// @Description  Returns one fuel consumption norm by ID.
// @Tags         Fuel Norms
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Fuel norm ID"
// @Success      200 {object} FuelNormResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/fuel/norms/{id} [get]
func (h *FuelNormHandler) GetFuelNormByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "fuel norm not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// CreateFuelNorm godoc
// @Summary      Create fuel norm
// @Description  Creates fuel consumption norm for a vehicle. Vehicle can have only one norm.
// @Tags         Fuel Norms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.FuelNormRequest true "Fuel norm payload"
// @Success      201 {object} FuelNormResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/fuel/norms [post]
func (h *FuelNormHandler) CreateFuelNorm(c *gin.Context) {
	var req dto.FuelNormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		writeFuelNormError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

// UpdateFuelNorm godoc
// @Summary      Update fuel norm
// @Description  Updates fuel consumption norm by ID.
// @Tags         Fuel Norms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Fuel norm ID"
// @Param        payload body dto.FuelNormRequest true "Fuel norm payload"
// @Success      200 {object} FuelNormResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/fuel/norms/{id} [put]
func (h *FuelNormHandler) UpdateFuelNorm(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.FuelNormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		writeFuelNormError(c, err)
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "fuel norm not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// DeleteFuelNorm godoc
// @Summary      Delete fuel norm
// @Description  Deletes fuel consumption norm by ID.
// @Tags         Fuel Norms
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Fuel norm ID"
// @Success      200 {object} FuelNormDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/fuel/norms/{id} [delete]
func (h *FuelNormHandler) DeleteFuelNorm(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deleted, err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "fuel norm not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return 0, false
	}
	return id, true
}

func parseIntQuery(c *gin.Context, key string, allowZero bool) (int, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}
	return n, true
}

func writeFuelNormError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrFuelNormDuplicateVehicle):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "fuel norm for vehicle already exists"})
	case errors.Is(err, repository.ErrFuelNormVehicleNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "vehicle not found"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
