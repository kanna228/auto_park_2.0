package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/repository"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type DriverShiftHandler struct {
	svc      *service.DriverShiftService
	auditSvc *auditlogservice.Service
}

func NewDriverShiftHandler(svc *service.DriverShiftService, auditSvc *auditlogservice.Service) *DriverShiftHandler {
	return &DriverShiftHandler{svc: svc, auditSvc: auditSvc}
}

type DriverShiftCreateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type DriverShiftResponseWrap struct {
	Success bool                    `json:"success" example:"true"`
	Data    dto.DriverShiftResponse `json:"data"`
}

type DriverShiftListResponseWrap struct {
	Success bool                        `json:"success" example:"true"`
	Data    dto.DriverShiftListResponse `json:"data"`
}

type DriverShiftUpdateResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

type DriverShiftDeleteResponseWrap struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"1"`
	} `json:"data"`
}

// CreateDriverShift godoc
// @Summary      Create driver shift
// @Description  Creates a driver shift. Driver shifts are connected with tripsheets through driver_shift_id.
// @Tags         Driver Shifts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.CreateDriverShiftRequest true "Driver shift create payload"
// @Success      201 {object} DriverShiftCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-shifts [post]
func (h *DriverShiftHandler) Create(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	var req dto.CreateDriverShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), userID, roleID, req)
	if err != nil {
		writeDriverShiftError(c, err)
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"info",
		"shift",
		"",
		"driver_shift_created",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id, "driver_id", req.DriverID, "date", req.ShiftDate),
	)

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// ListDriverShifts godoc
// @Summary      List driver shifts
// @Description  Returns driver shifts with filters by driver, date range and activity. Supports pagination.
// @Tags         Driver Shifts
// @Produce      json
// @Security     BearerAuth
// @Param        driver_id query int false "Filter by driver ID"
// @Param        date_from query string false "Shift date from, YYYY-MM-DD"
// @Param        date_to query string false "Shift date to, YYYY-MM-DD"
// @Param        is_active query bool false "Filter by active flag"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Param        sort_by query string false "Sort by field: shift_date, driver_id, time_from, time_to, is_active, tripsheets_count, created_at, updated_at"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} DriverShiftListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-shifts [get]
func (h *DriverShiftHandler) List(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	q := dto.DriverShiftListQuery{
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}

	var okQuery bool
	if q.DriverID, okQuery = parseInt64Query(c, "driver_id", false); !okQuery {
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
		writeDriverShiftError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// GetDriverShiftByID godoc
// @Summary      Get driver shift by ID
// @Description  Returns driver shift by ID including tripsheets connected to this shift.
// @Tags         Driver Shifts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver shift ID"
// @Success      200 {object} DriverShiftResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-shifts/{id} [get]
func (h *DriverShiftHandler) GetByID(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseDriverShiftID(c)
	if !ok {
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), userID, roleID, id)
	if err != nil {
		writeDriverShiftError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// UpdateDriverShift godoc
// @Summary      Update driver shift
// @Description  Updates driver shift fields.
// @Tags         Driver Shifts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver shift ID"
// @Param        payload body dto.UpdateDriverShiftRequest true "Driver shift update payload"
// @Success      200 {object} DriverShiftUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-shifts/{id} [put]
func (h *DriverShiftHandler) Update(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseDriverShiftID(c)
	if !ok {
		return
	}

	var req dto.UpdateDriverShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateByID(c.Request.Context(), userID, roleID, id, req)
	if err != nil {
		writeDriverShiftError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver shift not found"})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"info",
		"shift",
		"",
		"driver_shift_updated",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id, "driver_id", req.DriverID, "date", req.ShiftDate),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// UpdateDriverShiftActivity godoc
// @Summary      Update driver shift activity
// @Description  Updates only is_active flag for a driver shift.
// @Tags         Driver Shifts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver shift ID"
// @Param        payload body dto.UpdateDriverShiftActivityRequest true "Driver shift activity payload"
// @Success      200 {object} DriverShiftUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-shifts/{id}/activity [patch]
func (h *DriverShiftHandler) UpdateActivity(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseDriverShiftID(c)
	if !ok {
		return
	}

	var req dto.UpdateDriverShiftActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateActivityByID(c.Request.Context(), userID, roleID, id, req)
	if err != nil {
		writeDriverShiftError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver shift not found"})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"info",
		"shift",
		"",
		"driver_shift_activity_updated",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id, "is_active", req.IsActive),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeleteDriverShift godoc
// @Summary      Soft delete driver shift
// @Description  Does not physically delete shift. It marks shift as deleted and inactive, keeping data for audit.
// @Tags         Driver Shifts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver shift ID"
// @Success      200 {object} DriverShiftDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-shifts/{id} [delete]
func (h *DriverShiftHandler) Delete(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	id, ok := parseDriverShiftID(c)
	if !ok {
		return
	}

	deleted, err := h.svc.DeleteByID(c.Request.Context(), userID, roleID, id)
	if err != nil {
		writeDriverShiftError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver shift not found"})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"warning",
		"shift",
		"active",
		"driver_shift_deleted",
		auditlog.Actor(middleware.CurrentEmail(c), userID),
		auditlog.Message("shift_id", id),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func parseDriverShiftID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return 0, false
	}
	return id, true
}

func writeDriverShiftError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrDriverShiftNotFound):
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver shift not found"})
	case errors.Is(err, repository.ErrDriverShiftDriverNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "driver not found"})
	case errors.Is(err, service.ErrDriverShiftAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "access denied"})
	case errors.Is(err, service.ErrDriverShiftInvalidTime):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "time_to must be greater than time_from"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
