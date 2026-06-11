package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	"auto_park/modules/tripsheet_module/dto"

	"github.com/gin-gonic/gin"
)

// Delete godoc
// @Summary      Delete tripsheet
// @Description  Deletes a tripsheet by ID together with all related tripsheet trips.
// @Tags         Tripsheets
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Tripsheet ID"
// @Success      200 {object} SuccessMessageResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Router       /api/tripsheet/{id} [delete]
func (h *TripsheetHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	tripsheet, _ := h.svc.GetByID(c.Request.Context(), id)

	err = h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "tripsheet not found",
			})
			return
		}

		if err.Error() == "invalid id" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"warning",
		"tripsheet",
		"active",
		"deleted",
		auditlog.Actor(middleware.CurrentEmail(c), 0),
		auditlog.Message(
			"id", id,
			"number", tripsheetAuditNumber(tripsheet),
			"plate", tripsheetAuditPlate(tripsheet),
			"brand", tripsheetAuditBrand(tripsheet),
			"driver", tripsheetAuditDriver(tripsheet),
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "tripsheet and related trips deleted successfully",
	})
}

func tripsheetAuditNumber(t *dto.TripsheetResponse) string {
	if t == nil {
		return ""
	}
	return t.TripsheetNumber
}

func tripsheetAuditPlate(t *dto.TripsheetResponse) string {
	if t == nil {
		return ""
	}
	return t.VehiclePlateNumber
}

func tripsheetAuditBrand(t *dto.TripsheetResponse) string {
	if t == nil {
		return ""
	}
	return tripsheetAuditString(t.VehicleBrand)
}

func tripsheetAuditDriver(t *dto.TripsheetResponse) string {
	if t == nil {
		return ""
	}
	return strings.TrimSpace(tripsheetAuditString(t.DriverLastName) + " " + tripsheetAuditString(t.DriverFirstName) + " " + tripsheetAuditString(t.DriverMiddleName))
}

func tripsheetAuditString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
