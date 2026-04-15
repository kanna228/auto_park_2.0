package handlers

import (
	"auto_park/incident_module/dto"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Update godoc
// @Summary      Update incident
// @Description  Updates an existing incident by ID.
// @Tags         Incidents
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Incident ID"
// @Param        request body dto.IncidentUpdateRequest true "Incident update request"
// @Success      200 {object} IncidentResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/incidents/{id} [put]
func (h *IncidentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.IncidentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "incident not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}
