package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/repository"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type MechanicShiftHandler struct {
	svc      *service.MechanicShiftService
	auditSvc *auditlogservice.Service
}

func NewMechanicShiftHandler(svc *service.MechanicShiftService, auditSvc *auditlogservice.Service) *MechanicShiftHandler {
	return &MechanicShiftHandler{svc: svc, auditSvc: auditSvc}
}

type MechanicShiftCreateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type MechanicShiftResponseWrap struct {
	Success bool                      `json:"success" example:"true"`
	Data    dto.MechanicShiftResponse `json:"data"`
}

type MechanicShiftListResponseWrap struct {
	Success bool                          `json:"success" example:"true"`
	Data    dto.MechanicShiftListResponse `json:"data"`
}

type MechanicShiftUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type MechanicShiftDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

// CreateMechanicShift godoc
// @Summary      Create mechanic shift
// @Description  Creates a mechanic shift. Admin can create shifts for any duty mechanic. Duty mechanic can create only own shift. Data is kept for audit.
// @Tags         Mechanic Shifts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.CreateMechanicShiftRequest true "Mechanic shift create payload"
// @Success      201 {object} MechanicShiftCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/mechanic-shifts [post]
func (h *MechanicShiftHandler) Create(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	var req dto.CreateMechanicShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), userID, roleID, req)
	if err != nil {
		writeMechanicShiftError(c, err)
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"info",
		"shift",
		"",
		"mechanic_shift_created",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id, "mechanic_user_id", req.UserID, "date", req.ShiftDate),
	)

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// ListMechanicShifts godoc
// @Summary      List mechanic shifts
// @Description  Returns mechanic shifts with filters by mechanic, date range and activity. Admin, manager and garage dispatcher see all shifts; duty mechanic sees only own shifts.
// @Tags         Mechanic Shifts
// @Produce      json
// @Security     BearerAuth
// @Param        user_id query int false "Filter by mechanic user ID. Ignored for duty mechanic because they can see only own shifts"
// @Param        date_from query string false "Shift date from, YYYY-MM-DD"
// @Param        date_to query string false "Shift date to, YYYY-MM-DD"
// @Param        is_active query bool false "Filter by active flag"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by field: shift_date, user_id, time_from, time_to, is_active, created_at, updated_at"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} MechanicShiftListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/mechanic-shifts [get]
func (h *MechanicShiftHandler) List(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	q := dto.MechanicShiftListQuery{
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}

	var okQuery bool
	if q.UserID, okQuery = parseInt64Query(c, "user_id", false); !okQuery {
		return
	}
	if q.Limit, okQuery = parseIntQuery(c, "limit", false); !okQuery {
		return
	}
	if q.Offset, okQuery = parseIntQuery(c, "offset", true); !okQuery {
		return
	}
	if active, provided, ok := parseBoolQuery(c, "is_active"); !ok {
		return
	} else if provided {
		q.IsActive = &active
	}

	resp, err := h.svc.List(c.Request.Context(), userID, roleID, q)
	if err != nil {
		writeMechanicShiftError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// GetMechanicShiftByID godoc
// @Summary      Get mechanic shift by ID
// @Description  Returns mechanic shift by ID. Duty mechanic can see only own shift.
// @Tags         Mechanic Shifts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Mechanic shift ID"
// @Success      200 {object} MechanicShiftResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/mechanic-shifts/{id} [get]
func (h *MechanicShiftHandler) GetByID(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseMechanicShiftID(c)
	if !ok {
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), userID, roleID, id)
	if err != nil {
		writeMechanicShiftError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// UpdateMechanicShift godoc
// @Summary      Update mechanic shift
// @Description  Updates mechanic shift. Admin can update any shift. Duty mechanic can update only own shift. Shift user can be changed only by admin.
// @Tags         Mechanic Shifts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Mechanic shift ID"
// @Param        payload body dto.UpdateMechanicShiftRequest true "Mechanic shift update payload"
// @Success      200 {object} MechanicShiftUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/mechanic-shifts/{id} [put]
func (h *MechanicShiftHandler) Update(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseMechanicShiftID(c)
	if !ok {
		return
	}

	var req dto.UpdateMechanicShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateByID(c.Request.Context(), userID, roleID, id, req)
	if err != nil {
		writeMechanicShiftError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "mechanic shift not found"})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"info",
		"shift",
		"",
		"mechanic_shift_updated",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id, "mechanic_user_id", req.UserID, "date", req.ShiftDate),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// UpdateMechanicShiftActivity godoc
// @Summary      Update mechanic shift activity
// @Description  Updates only is_active flag for a mechanic shift.
// @Tags         Mechanic Shifts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Mechanic shift ID"
// @Param        payload body dto.UpdateMechanicShiftActivityRequest true "Mechanic shift activity payload"
// @Success      200 {object} MechanicShiftUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/mechanic-shifts/{id}/activity [patch]
func (h *MechanicShiftHandler) UpdateActivity(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseMechanicShiftID(c)
	if !ok {
		return
	}

	var req dto.UpdateMechanicShiftActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateActivityByID(c.Request.Context(), userID, roleID, id, req)
	if err != nil {
		writeMechanicShiftError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "mechanic shift not found"})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"info",
		"shift",
		"",
		"mechanic_shift_activity_updated",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id, "is_active", req.IsActive),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeleteMechanicShift godoc
// @Summary      Soft delete mechanic shift
// @Description  Does not physically delete shift. It marks shift as deleted and inactive, keeping data for audit.
// @Tags         Mechanic Shifts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Mechanic shift ID"
// @Success      200 {object} MechanicShiftDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/mechanic-shifts/{id} [delete]
func (h *MechanicShiftHandler) Delete(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseMechanicShiftID(c)
	if !ok {
		return
	}

	deleted, err := h.svc.DeleteByID(c.Request.Context(), userID, roleID, id)
	if err != nil {
		writeMechanicShiftError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "mechanic shift not found"})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"warning",
		"shift",
		"active",
		"mechanic_shift_deleted",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func parseMechanicShiftID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return 0, false
	}
	return id, true
}

func parseInt64Query(c *gin.Context, key string, allowZero bool) (int64, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, true
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}

	return n, true
}

func parseIntQuery(c *gin.Context, key string, allowZero bool) (int, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, true
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}

	return n, true
}

func parseBoolQuery(c *gin.Context, key string) (bool, bool, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return false, false, true
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return false, true, false
	}

	return parsed, true, true
}

func writeMechanicShiftError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrMechanicShiftNotFound):
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "mechanic shift not found"})
	case errors.Is(err, repository.ErrMechanicShiftUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "user not found"})
	case errors.Is(err, repository.ErrMechanicShiftUserNotMechanic):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "user is not duty mechanic"})
	case errors.Is(err, service.ErrMechanicShiftAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "access denied"})
	case errors.Is(err, service.ErrMechanicShiftInvalidTime):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "time_to must be greater than time_from"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
