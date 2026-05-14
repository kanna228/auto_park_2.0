package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

type TirePlaceHandler struct {
	svc service.TirePlaceService
}

func NewTirePlaceHandler(svc service.TirePlaceService) *TirePlaceHandler {
	return &TirePlaceHandler{svc: svc}
}

// CreateTirePlace godoc
// @Summary      Create tire place
// @Description  Creates a new tire place.
// @Tags         Tire Places
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.TirePlaceCreateRequest true "Tire place create payload"
// @Success      201 {object} TirePlaceCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tire-places [post]
func (h *TirePlaceHandler) CreateTirePlace(c *gin.Context) {
	var req dto.TirePlaceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}

// GetTirePlaceByID godoc
// @Summary      Get tire place by ID
// @Description  Returns tire place by id.
// @Tags         Tire Places
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tire place ID"
// @Success      200 {object} TirePlaceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tire-places/{id} [get]
func (h *TirePlaceHandler) GetTirePlaceByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire place not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// ListTirePlaces godoc
// @Summary      List tire places
// @Description  Returns tire places with filters, pagination and sorting.
// @Tags         Tire Places
// @Produce      json
// @Security     BearerAuth
// @Param        name query string false "Filter by tire place name"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by: id, name"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} TirePlaceListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tire-places [get]
func (h *TirePlaceHandler) ListTirePlaces(c *gin.Context) {
	q := dto.TirePlaceListQuery{
		Name:   c.Query("name"),
		SortBy: c.Query("sort_by"),
		Order:  c.Query("order"),
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

// UpdateTirePlace godoc
// @Summary      Update tire place
// @Description  Updates tire place by id.
// @Tags         Tire Places
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tire place ID"
// @Param        payload body dto.TirePlaceUpdateRequest true "Tire place update payload"
// @Success      200 {object} TirePlaceUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tire-places/{id} [put]
func (h *TirePlaceHandler) UpdateTirePlace(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.TirePlaceUpdateRequest
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire place not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}

// DeleteTirePlace godoc
// @Summary      Delete tire place
// @Description  Deletes tire place by id.
// @Tags         Tire Places
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tire place ID"
// @Success      200 {object} TirePlaceDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tire-places/{id} [delete]
func (h *TirePlaceHandler) DeleteTirePlace(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	ok, err := h.svc.DeleteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire place not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}
