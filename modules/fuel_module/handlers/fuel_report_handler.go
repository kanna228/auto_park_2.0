package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"auto_park/modules/fuel_module/dto"
	"auto_park/modules/fuel_module/service"

	"github.com/gin-gonic/gin"
)

type FuelReportHandler struct {
	svc service.FuelReportService
}

func NewFuelReportHandler(svc service.FuelReportService) *FuelReportHandler {
	return &FuelReportHandler{svc: svc}
}

// DownloadDriverReport godoc
// @Summary      Download fuel report by driver
// @Description  Downloads fuel report for a specific driver in PDF or Excel format.
// @Tags         Fuel Reports
// @Produce      application/pdf,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        driver_id path int true "Driver ID"
// @Param        format query string false "Report format: pdf or xlsx" Enums(pdf,xlsx) default(pdf)
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to query string false "End date (YYYY-MM-DD)"
// @Success      200 {file} binary
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/fuel/reports/driver/{driver_id} [get]
func (h *FuelReportHandler) DownloadDriverReport(c *gin.Context) {
	driverID, err := strconv.ParseInt(c.Param("driver_id"), 10, 64)
	if err != nil || driverID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid driver_id"})
		return
	}

	var filter dto.FuelReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	file, err := h.svc.DownloadDriverReport(c.Request.Context(), driverID, filter)
	if err != nil {
		h.handleReportError(c, err, "driver not found")
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
	c.Data(http.StatusOK, file.ContentType, file.Data)
}

// DownloadVehicleReport godoc
// @Summary      Download fuel report by vehicle
// @Description  Downloads fuel report for a specific vehicle in PDF or Excel format.
// @Tags         Fuel Reports
// @Produce      application/pdf,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        vehicle_id path int true "Vehicle ID"
// @Param        format query string false "Report format: pdf or xlsx" Enums(pdf,xlsx) default(pdf)
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to query string false "End date (YYYY-MM-DD)"
// @Success      200 {file} binary
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/fuel/reports/vehicle/{vehicle_id} [get]
func (h *FuelReportHandler) DownloadVehicleReport(c *gin.Context) {
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil || vehicleID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid vehicle_id"})
		return
	}

	var filter dto.FuelReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	file, err := h.svc.DownloadVehicleReport(c.Request.Context(), vehicleID, filter)
	if err != nil {
		h.handleReportError(c, err, "vehicle not found")
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
	c.Data(http.StatusOK, file.ContentType, file.Data)
}

// DownloadTripsheetReport godoc
// @Summary      Download fuel report by tripsheet
// @Description  Downloads fuel report for a specific tripsheet in PDF or Excel format.
// @Tags         Fuel Reports
// @Produce      application/pdf,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        tripsheet_id path int true "Tripsheet ID"
// @Param        format query string false "Report format: pdf or xlsx" Enums(pdf,xlsx) default(pdf)
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to query string false "End date (YYYY-MM-DD)"
// @Success      200 {file} binary
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/fuel/reports/tripsheet/{tripsheet_id} [get]
func (h *FuelReportHandler) DownloadTripsheetReport(c *gin.Context) {
	tripsheetID, err := strconv.ParseInt(c.Param("tripsheet_id"), 10, 64)
	if err != nil || tripsheetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid tripsheet_id"})
		return
	}

	var filter dto.FuelReportFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	file, err := h.svc.DownloadTripsheetReport(c.Request.Context(), tripsheetID, filter)
	if err != nil {
		h.handleReportError(c, err, "tripsheet not found")
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
	c.Data(http.StatusOK, file.ContentType, file.Data)
}

func (h *FuelReportHandler) handleReportError(c *gin.Context, err error, notFoundMessage string) {
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": notFoundMessage})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
}
