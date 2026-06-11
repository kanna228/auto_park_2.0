package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	"auto_park/modules/incident_module/dto"

	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary      Delete incident
// @Description  Deletes an incident by ID.
// @Tags         Incidents
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Incident ID"
// @Success      200 {object} SuccessMessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/incidents/{id} [delete]
func (h *IncidentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	incident, _ := h.svc.GetByID(c.Request.Context(), id)

	err = h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "incident not found"})
			return
		}
		if err.Error() == "invalid id" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"warning",
		"incident",
		"registered",
		"deleted",
		auditlog.Actor(middleware.CurrentEmail(c), 0),
		auditlog.Message(
			"id", id,
			"type", incidentAuditType(incident),
			"plate", incidentAuditPlate(incident),
			"driver", incidentAuditDriver(incident),
			"mechanic", incidentAuditMechanic(incident),
			"location", incidentAuditLocation(incident),
		),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "incident deleted successfully"})
}

func incidentAuditType(i *dto.IncidentResponse) string {
	if i == nil {
		return ""
	}
	return i.IncidentTypeName
}

func incidentAuditPlate(i *dto.IncidentResponse) string {
	if i == nil {
		return ""
	}
	return i.VehicleStateNumber
}

func incidentAuditDriver(i *dto.IncidentResponse) string {
	if i == nil {
		return ""
	}
	return i.DriverFullName
}

func incidentAuditMechanic(i *dto.IncidentResponse) string {
	if i == nil {
		return ""
	}
	return i.MechanicFullName
}

func incidentAuditLocation(i *dto.IncidentResponse) string {
	if i == nil {
		return ""
	}
	return i.Location
}
