package handlers

import (
	"net/http"

	"auto_park/modules/dashboard_module/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// GetStats godoc
// @Summary      Get dashboard statistics
// @Description  Returns weekly KPI cards, monthly fuel/repair expenses, current fleet status and warehouse stock summary. Warehouse stock items are limited for dashboard preview.
// @Tags         Dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} DashboardStatsResponseWrap
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/dashboard/stats [get]
func (h *DashboardHandler) GetStats(c *gin.Context) {
	resp, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ExportStatsExcel godoc
// @Summary      Export dashboard statistics to Excel
// @Description  Exports dashboard statistics to XLSX. Unlike the JSON dashboard preview, warehouse stock is exported without item limit.
// @Tags         Dashboard
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Success      200 {file} binary
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/dashboard/stats/export/excel [get]
func (h *DashboardHandler) ExportStatsExcel(c *gin.Context) {
	data, filename, err := h.svc.ExportStatsExcel(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ExportStatsPDF godoc
// @Summary      Export dashboard statistics to PDF
// @Description  Exports dashboard statistics to PDF. Unlike the JSON dashboard preview, warehouse stock is exported without item limit.
// @Tags         Dashboard
// @Produce      application/pdf
// @Security     BearerAuth
// @Success      200 {file} binary
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/dashboard/stats/export/pdf [get]
func (h *DashboardHandler) ExportStatsPDF(c *gin.Context) {
	data, filename, err := h.svc.ExportStatsPDF(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", data)
}
