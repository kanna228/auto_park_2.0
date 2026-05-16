package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/maintenance_schedule_module/dto"
	"auto_park/modules/maintenance_schedule_module/service"

	"github.com/gin-gonic/gin"
)

type MaintenanceScheduleHandler struct {
	svc service.MaintenanceScheduleService
}

func NewMaintenanceScheduleHandler(svc service.MaintenanceScheduleService) *MaintenanceScheduleHandler {
	return &MaintenanceScheduleHandler{svc: svc}
}

// ListMaintenanceSchedules godoc
// @Summary      List maintenance schedules
// @Description  Returns paginated maintenance schedules with optional draft filter.
// @Tags         Maintenance Schedules
// @Produce      json
// @Security     BearerAuth
// @Param        is_draft query bool false "Filter by draft state"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} MaintenanceScheduleListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-schedules [get]
func (h *MaintenanceScheduleHandler) ListMaintenanceSchedules(c *gin.Context) {
	q, ok := parseListQuery(c)
	if !ok {
		return
	}
	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// GetMaintenanceScheduleByID godoc
// @Summary      Get maintenance schedule by ID
// @Description  Returns one maintenance schedule by ID.
// @Tags         Maintenance Schedules
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Maintenance schedule ID"
// @Success      200 {object} MaintenanceScheduleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-schedules/{id} [get]
func (h *MaintenanceScheduleHandler) GetMaintenanceScheduleByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "maintenance schedule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// CreateMaintenanceSchedule godoc
// @Summary      Create maintenance schedule
// @Description  Creates draft or published maintenance schedule.
// @Tags         Maintenance Schedules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.MaintenanceScheduleRequest true "Maintenance schedule payload"
// @Success      201 {object} MaintenanceScheduleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-schedules [post]
func (h *MaintenanceScheduleHandler) CreateMaintenanceSchedule(c *gin.Context) {
	var req dto.MaintenanceScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

// UpdateMaintenanceSchedule godoc
// @Summary      Update maintenance schedule
// @Description  Updates maintenance schedule by ID.
// @Tags         Maintenance Schedules
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Maintenance schedule ID"
// @Param        payload body dto.MaintenanceScheduleRequest true "Maintenance schedule payload"
// @Success      200 {object} MaintenanceScheduleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-schedules/{id} [put]
func (h *MaintenanceScheduleHandler) UpdateMaintenanceSchedule(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.MaintenanceScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "maintenance schedule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// DeleteMaintenanceSchedule godoc
// @Summary      Delete maintenance schedule
// @Description  Deletes maintenance schedule by ID.
// @Tags         Maintenance Schedules
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Maintenance schedule ID"
// @Success      200 {object} MaintenanceScheduleDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-schedules/{id} [delete]
func (h *MaintenanceScheduleHandler) DeleteMaintenanceSchedule(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "maintenance schedule not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func parseListQuery(c *gin.Context) (dto.MaintenanceScheduleListQuery, bool) {
	q := dto.MaintenanceScheduleListQuery{}
	if raw := strings.TrimSpace(c.Query("is_draft")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid is_draft"})
			return q, false
		}
		q.IsDraft = &value
	}
	var ok bool
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return q, false
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return q, false
	}
	return q, true
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
