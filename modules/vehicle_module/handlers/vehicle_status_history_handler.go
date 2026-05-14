package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/vehicle_module/dto"

	"github.com/gin-gonic/gin"
)

// GetVehicleStatusHistoryByID godoc
// @Summary      Get vehicle status history record by ID
// @Description  Returns one vehicle status history record. If end_date is null, end_date_display is returned as "По настоящее время".
// @Tags         Vehicle Status History
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle status history ID"
// @Success      200 {object} VehicleStatusHistoryResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/status-history/{id} [get]
func (h *VehicleHandler) GetVehicleStatusHistoryByID(c *gin.Context) {
	id, ok := parseVehicleStatusHistoryID(c, "id")
	if !ok {
		return
	}

	item, err := h.svc.GetVehicleStatusHistoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle status history not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListVehicleStatusHistory godoc
// @Summary      List vehicle status history
// @Description  Returns vehicle status history with filters and pagination. Current status records have end_date=null and end_date_display="По настоящее время".
// @Tags         Vehicle Status History
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id query int false "Filter by vehicle ID"
// @Param        status_id query int false "Filter by status ID"
// @Param        start_from query string false "Start date from, YYYY-MM-DD"
// @Param        start_to query string false "Start date to, YYYY-MM-DD"
// @Param        end_from query string false "End date from, YYYY-MM-DD"
// @Param        end_to query string false "End date to, YYYY-MM-DD"
// @Param        is_current query bool false "Filter current records where end_date is null"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by: start_date, end_date, vehicle_id, status_id, status_name, created_at, updated_at"
// @Param        order query string false "Order: asc or desc"
// @Success      200 {object} VehicleStatusHistoryListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/status-history [get]
func (h *VehicleHandler) ListVehicleStatusHistory(c *gin.Context) {
	q := dto.VehicleStatusHistoryListQuery{
		StartFrom: c.Query("start_from"),
		StartTo:   c.Query("start_to"),
		EndFrom:   c.Query("end_from"),
		EndTo:     c.Query("end_to"),
		SortBy:    c.Query("sort_by"),
		Order:     c.Query("order"),
	}

	var ok bool
	if q.VehicleID, ok = parseVehicleStatusHistoryInt64Query(c, "vehicle_id", false); !ok {
		return
	}
	if q.StatusID, ok = parseVehicleStatusHistoryInt64Query(c, "status_id", false); !ok {
		return
	}
	if q.Limit, ok = parseVehicleStatusHistoryIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseVehicleStatusHistoryIntQuery(c, "offset", true); !ok {
		return
	}

	if active, provided, ok := parseVehicleStatusHistoryBoolQuery(c, "is_current"); !ok {
		return
	} else if provided {
		q.IsCurrent = &active
	}

	resp, err := h.svc.ListVehicleStatusHistory(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func parseVehicleStatusHistoryID(c *gin.Context, key string) (int64, bool) {
	value := strings.TrimSpace(c.Param(key))
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}
	return id, true
}

func parseVehicleStatusHistoryInt64Query(c *gin.Context, key string, allowZero bool) (int64, bool) {
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

func parseVehicleStatusHistoryIntQuery(c *gin.Context, key string, allowZero bool) (int, bool) {
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

func parseVehicleStatusHistoryBoolQuery(c *gin.Context, key string) (bool, bool, bool) {
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
