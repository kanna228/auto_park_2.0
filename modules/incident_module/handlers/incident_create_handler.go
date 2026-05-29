package handlers

import (
	"auto_park/internal/apierrors"
	"auto_park/modules/incident_module/dto"
	"auto_park/modules/incident_module/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	svc service.IncidentService
}

func NewIncidentHandler(svc service.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

// Create godoc
// @Summary      Create incident
// @Description  Creates a new vehicle incident, links it to mechanic shift and automatically sets vehicle status to "На ТО".
// @Tags         Incidents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.IncidentCreateRequest true "Incident create request"
// @Success      201 {object} IncidentCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/incidents [post]
func (h *IncidentHandler) Create(c *gin.Context) {
	var req dto.IncidentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apierrors.ErrEntityArchived) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "code": apierrors.CodeEntityArchived, "error": "Нельзя использовать архивный объект"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": resp})
}
