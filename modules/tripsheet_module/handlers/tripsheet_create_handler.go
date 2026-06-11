package handlers

import (
	"errors"
	"net/http"

	"auto_park/internal/apierrors"
	"auto_park/internal/auditlog"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/service"

	"github.com/gin-gonic/gin"
)

type TripsheetHandler struct {
	svc      service.TripsheetService
	auditSvc *auditlogservice.Service
}

func NewTripsheetHandler(svc service.TripsheetService, auditSvc *auditlogservice.Service) *TripsheetHandler {
	return &TripsheetHandler{
		svc:      svc,
		auditSvc: auditSvc,
	}
}

// Create godoc
// @Summary      Create tripsheet
// @Description  Creates a new tripsheet.
// @Tags         Tripsheets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateTripsheetRequest true "Tripsheet create request"
// @Success      201 {object} TripsheetCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/tripsheet [post]
func (h *TripsheetHandler) Create(c *gin.Context) {
	var req dto.CreateTripsheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apierrors.ErrEntityArchived) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "code": apierrors.CodeEntityArchived, "error": "Нельзя использовать архивный объект"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"success",
		"tripsheet",
		"",
		"created",
		auditlog.Actor(middleware.CurrentEmail(c), 0),
		auditlog.Message("id", resp.ID, "number", resp.TripsheetNumber, "vehicle_id", req.VehicleID, "driver_id", req.DriverID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    resp,
	})
}
