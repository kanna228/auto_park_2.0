package handlers

import (
	"auto_park/modules/incident_module/dto"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAll godoc
// @Summary      Get all incidents
// @Description  Returns all incidents with optional filters.
// @Tags         Incidents
// @Produce      json
// @Security     BearerAuth
// @Param        incident_type_id query int false "Incident type ID"
// @Param        vehicle_id query int false "Vehicle ID"
// @Param        driver_id query int false "Driver ID"
// @Param        mechanic_id query int false "Mechanic ID"
// @Param        tripsheet_id query int false "Tripsheet ID"
// @Param        date_from query string false "Date from (YYYY-MM-DD)"
// @Param        date_to query string false "Date to (YYYY-MM-DD)"
// @Success      200 {object} IncidentListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/incidents [get]
func (h *IncidentHandler) GetAll(c *gin.Context) {
	var filter dto.IncidentListQuery
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	items, total, err := h.svc.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "items": items, "total": total})
}

// GetByID godoc
// @Summary      Get incident by ID
// @Description  Returns a single incident by its ID.
// @Tags         Incidents
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Incident ID"
// @Success      200 {object} IncidentResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/incidents/{id} [get]
func (h *IncidentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListIncidentTypes godoc
// @Summary      List incident types
// @Description  Returns all available incident types.
// @Tags         Incidents
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} IncidentTypeListResponseWrap
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/incidents/types [get]
func (h *IncidentHandler) ListIncidentTypes(c *gin.Context) {
	items, total, err := h.svc.ListIncidentTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "items": items, "total": total})
}
