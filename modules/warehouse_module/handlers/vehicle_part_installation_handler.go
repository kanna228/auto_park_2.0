package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/repository"
	"auto_park/modules/warehouse_module/service"

	"github.com/gin-gonic/gin"
)

type VehiclePartInstallationHandler struct {
	svc service.VehiclePartInstallationService
}

func NewVehiclePartInstallationHandler(svc service.VehiclePartInstallationService) *VehiclePartInstallationHandler {
	return &VehiclePartInstallationHandler{svc: svc}
}

// CreateVehiclePartInstallation godoc
// @Summary      Install part on vehicle
// @Description  Creates a vehicle part installation record, validates warehouse stock, decreases part quantity and stores installation/replacement dates for vehicle passport history. Available to duty mechanic and warehouse manager.
// @Tags         Warehouse Vehicle Part Installations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.VehiclePartInstallationCreateRequest true "Vehicle part installation create payload"
// @Success      201 {object} VehiclePartInstallationCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-part-installations [post]
func (h *VehiclePartInstallationHandler) CreateVehiclePartInstallation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	var req dto.VehiclePartInstallationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		writeVehiclePartInstallationError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}

// GetVehiclePartInstallationByID godoc
// @Summary      Get vehicle part installation by ID
// @Description  Returns installed vehicle part record by ID.
// @Tags         Warehouse Vehicle Part Installations
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle part installation ID"
// @Success      200 {object} VehiclePartInstallationResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-part-installations/{id} [get]
func (h *VehiclePartInstallationHandler) GetVehiclePartInstallationByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle part installation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListVehiclePartInstallations godoc
// @Summary      List vehicle part installations
// @Description  Returns vehicle part installation history with filters, pagination and sorting.
// @Tags         Warehouse Vehicle Part Installations
// @Produce      json
// @Security     BearerAuth
// @Param        part_id query int false "Filter by warehouse part internal ID"
// @Param        vehicle_id query int false "Filter by vehicle ID"
// @Param        installed_by_user_id query int false "Filter by installer user ID"
// @Param        is_active query bool false "Filter by active flag"
// @Param        date_from query string false "Installed date from, YYYY-MM-DD"
// @Param        date_to query string false "Installed date to, YYYY-MM-DD"
// @Param        replacement_from query string false "Planned replacement date from, YYYY-MM-DD"
// @Param        replacement_to query string false "Planned replacement date to, YYYY-MM-DD"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by field: installed_at, planned_replacement_at, created_at, updated_at, part_id, vehicle_id, quantity, installed_by_user_id, is_active"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} VehiclePartInstallationListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-part-installations [get]
func (h *VehiclePartInstallationHandler) ListVehiclePartInstallations(c *gin.Context) {
	q := dto.VehiclePartInstallationListQuery{
		DateFrom:        c.Query("date_from"),
		DateTo:          c.Query("date_to"),
		ReplacementFrom: c.Query("replacement_from"),
		ReplacementTo:   c.Query("replacement_to"),
		SortBy:          c.Query("sort_by"),
		Order:           c.Query("order"),
	}

	var ok bool

	if q.PartID, ok = parseInt64Query(c, "part_id", false); !ok {
		return
	}
	if q.VehicleID, ok = parseInt64Query(c, "vehicle_id", false); !ok {
		return
	}
	if q.InstalledByUserID, ok = parseInt64Query(c, "installed_by_user_id", false); !ok {
		return
	}
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return
	}

	if active, provided, ok := parseBoolQuery(c, "is_active"); !ok {
		return
	} else if provided {
		q.IsActive = &active
	}

	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// UpdateVehiclePartInstallation godoc
// @Summary      Update vehicle part installation
// @Description  Fully updates vehicle part installation. If part or quantity is changed, warehouse stock is adjusted safely. The installed_at date is locked after save for all users except administrator.
// @Tags         Warehouse Vehicle Part Installations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle part installation ID"
// @Param        payload body dto.VehiclePartInstallationUpdateRequest true "Vehicle part installation update payload"
// @Success      200 {object} VehiclePartInstallationUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-part-installations/{id} [put]
func (h *VehiclePartInstallationHandler) UpdateVehiclePartInstallation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	roleID, ok := getRoleIDFromContext(c)
	if !ok {
		return
	}

	var req dto.VehiclePartInstallationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateByID(c.Request.Context(), id, userID, roleID, req)
	if err != nil {
		writeVehiclePartInstallationError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle part installation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}

// UpdateVehiclePartInstallationActivity godoc
// @Summary      Change vehicle part installation activity
// @Description  Changes only active flag. Use this to close old exploitation record before repeating non-consumable part installation.
// @Tags         Warehouse Vehicle Part Installations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle part installation ID"
// @Param        payload body dto.VehiclePartInstallationActivityUpdateRequest true "Activity update payload"
// @Success      200 {object} VehiclePartInstallationActivityUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-part-installations/{id}/activity [patch]
func (h *VehiclePartInstallationHandler) UpdateVehiclePartInstallationActivity(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.VehiclePartInstallationActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateActivityByID(c.Request.Context(), id, req)
	if err != nil {
		writeVehiclePartInstallationError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle part installation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}

// DeleteVehiclePartInstallation godoc
// @Summary      Delete vehicle part installation
// @Description  Deletes vehicle part installation and returns its quantity back to warehouse stock. Available to duty mechanic and warehouse manager.
// @Tags         Warehouse Vehicle Part Installations
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle part installation ID"
// @Success      200 {object} VehiclePartInstallationDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-part-installations/{id} [delete]
func (h *VehiclePartInstallationHandler) DeleteVehiclePartInstallation(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	deleted, err := h.svc.DeleteByID(c.Request.Context(), id)
	if err != nil {
		writeVehiclePartInstallationError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle part installation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id": id,
		},
	})
}

func getRoleIDFromContext(c *gin.Context) (int64, bool) {
	v, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "role id not found in context"})
		return 0, false
	}

	roleID, ok := v.(int64)
	if !ok || roleID <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "invalid role id"})
		return 0, false
	}

	return roleID, true
}

func parseBoolQuery(c *gin.Context, key string) (bool, bool, bool) {
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

func writeVehiclePartInstallationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrVehiclePartInstallationPartNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part not found"})

	case errors.Is(err, repository.ErrVehiclePartInstallationVehicleNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "vehicle not found"})

	case errors.Is(err, repository.ErrVehiclePartInstallationUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "installer user not found"})

	case errors.Is(err, repository.ErrVehiclePartInstallationInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "not enough part quantity in warehouse"})

	case errors.Is(err, repository.ErrVehiclePartInstallationActiveDuplicate):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "non-consumable part is already active on this vehicle"})

	case errors.Is(err, service.ErrVehiclePartInstallationInstalledAtLocked):
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "installed_at cannot be changed after save except by administrator"})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
