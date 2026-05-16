package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/maintenance_execution_module/dto"
	"auto_park/modules/maintenance_execution_module/repository"
	"auto_park/modules/maintenance_execution_module/service"

	"github.com/gin-gonic/gin"
)

type MaintenanceExecutionHandler struct {
	svc service.MaintenanceExecutionService
}

func NewMaintenanceExecutionHandler(svc service.MaintenanceExecutionService) *MaintenanceExecutionHandler {
	return &MaintenanceExecutionHandler{svc: svc}
}

// ListMaintenanceExecutions godoc
// @Summary      List maintenance executions
// @Description  Returns paginated maintenance execution records filtered by required schedule_id.
// @Tags         Maintenance Executions
// @Produce      json
// @Security     BearerAuth
// @Param        schedule_id query int true "Maintenance schedule ID"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} MaintenanceExecutionListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions [get]
func (h *MaintenanceExecutionHandler) ListMaintenanceExecutions(c *gin.Context) {
	scheduleID, ok := parseInt64Query(c, "schedule_id", false, true)
	if !ok {
		return
	}
	limit, ok := parseIntQuery(c, "limit", false)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}
	resp, err := h.svc.ListBySchedule(c.Request.Context(), scheduleID, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// GetMaintenanceExecutionByID godoc
// @Summary      Get maintenance execution by ID
// @Description  Returns one maintenance execution record by ID.
// @Tags         Maintenance Executions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Maintenance execution ID"
// @Success      200 {object} MaintenanceExecutionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions/{id} [get]
func (h *MaintenanceExecutionHandler) GetMaintenanceExecutionByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "maintenance execution not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListMaintenanceExecutionsBySchedule godoc
// @Summary      List maintenance executions by schedule
// @Description  Returns all maintenance execution records for a schedule.
// @Tags         Maintenance Executions
// @Produce      json
// @Security     BearerAuth
// @Param        schedule_id path int true "Maintenance schedule ID"
// @Success      200 {object} MaintenanceExecutionListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions/by-schedule/{schedule_id} [get]
func (h *MaintenanceExecutionHandler) ListMaintenanceExecutionsBySchedule(c *gin.Context) {
	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}
	resp, err := h.svc.ListBySchedule(c.Request.Context(), scheduleID, 10000, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// CreateMaintenanceExecution godoc
// @Summary      Create maintenance execution
// @Description  Creates one maintenance execution record.
// @Tags         Maintenance Executions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.MaintenanceExecutionRequest true "Maintenance execution payload"
// @Success      201 {object} MaintenanceExecutionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions [post]
func (h *MaintenanceExecutionHandler) CreateMaintenanceExecution(c *gin.Context) {
	var req dto.MaintenanceExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		writeMaintenanceExecutionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

// BulkCreateMaintenanceExecutions godoc
// @Summary      Bulk create maintenance executions
// @Description  Creates maintenance execution records for all boards in a schedule.
// @Tags         Maintenance Executions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.MaintenanceExecutionBulkRequest true "Maintenance execution bulk payload"
// @Success      201 {object} MaintenanceExecutionListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions/bulk [post]
func (h *MaintenanceExecutionHandler) BulkCreateMaintenanceExecutions(c *gin.Context) {
	var req dto.MaintenanceExecutionBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	resp, err := h.svc.BulkCreate(c.Request.Context(), req)
	if err != nil {
		writeMaintenanceExecutionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": resp})
}

// UpdateMaintenanceExecution godoc
// @Summary      Update maintenance execution
// @Description  Updates maintenance execution record by ID.
// @Tags         Maintenance Executions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Maintenance execution ID"
// @Param        payload body dto.MaintenanceExecutionRequest true "Maintenance execution payload"
// @Success      200 {object} MaintenanceExecutionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions/{id} [put]
func (h *MaintenanceExecutionHandler) UpdateMaintenanceExecution(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.MaintenanceExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		writeMaintenanceExecutionError(c, err)
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "maintenance execution not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// DeleteMaintenanceExecution godoc
// @Summary      Delete maintenance execution
// @Description  Deletes maintenance execution record by ID.
// @Tags         Maintenance Executions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Maintenance execution ID"
// @Success      200 {object} MaintenanceExecutionDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/maintenance-executions/{id} [delete]
func (h *MaintenanceExecutionHandler) DeleteMaintenanceExecution(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "maintenance execution not found"})
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

func parseScheduleID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("schedule_id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid schedule_id"})
		return 0, false
	}
	return id, true
}

func parseInt64Query(c *gin.Context, key string, allowZero bool, required bool) (int64, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		if required {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": key + " is required"})
			return 0, false
		}
		return 0, true
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}
	return n, true
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

func writeMaintenanceExecutionError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrMaintenanceScheduleNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "maintenance schedule not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
}
