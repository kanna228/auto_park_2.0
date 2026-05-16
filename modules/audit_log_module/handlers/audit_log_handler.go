package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/audit_log_module/dto"
	"auto_park/modules/audit_log_module/service"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	svc *service.Service
}

func NewAuditLogHandler(svc *service.Service) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

// ListAuditLogs godoc
// @Summary      List audit logs
// @Description  Returns audit logs with function, level, date and search filters.
// @Tags         Audit Logs
// @Produce      json
// @Security     BearerAuth
// @Param        function query string false "Function: request, arrival, tripsheet, shift, incident, vehicle, user"
// @Param        level query string false "Level: system, info, success, warning, error"
// @Param        date query string false "Created date, YYYY-MM-DD"
// @Param        search query string false "Substring by actor/from_status/to_status"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} AuditLogListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/audit-logs [get]
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	q, ok := parseAuditLogQuery(c)
	if !ok {
		return
	}
	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ExportAuditLogs godoc
// @Summary      Export audit logs
// @Description  Exports filtered audit logs as UTF-8 CSV with BOM.
// @Tags         Audit Logs
// @Produce      text/csv
// @Security     BearerAuth
// @Param        function query string false "Function: request, arrival, tripsheet, shift, incident, vehicle, user"
// @Param        level query string false "Level: system, info, success, warning, error"
// @Param        date query string false "Created date, YYYY-MM-DD"
// @Param        search query string false "Substring by actor/from_status/to_status"
// @Success      200 {file} file
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/audit-logs/export [get]
func (h *AuditLogHandler) ExportAuditLogs(c *gin.Context) {
	q, ok := parseAuditLogQuery(c)
	if !ok {
		return
	}
	data, err := h.svc.ExportCSV(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="audit-logs.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

func parseAuditLogQuery(c *gin.Context) (dto.AuditLogListQuery, bool) {
	q := dto.AuditLogListQuery{
		Function: c.Query("function"),
		Level:    c.Query("level"),
		Date:     c.Query("date"),
		Search:   c.Query("search"),
	}
	var ok bool
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return dto.AuditLogListQuery{}, false
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return dto.AuditLogListQuery{}, false
	}
	return q, true
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
