package handlers

import (
	"errors"
	"net/http"
	"strings"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/repository"
	"auto_park/modules/warehouse_module/service"

	"github.com/gin-gonic/gin"
)

type VehicleServiceHandler struct {
	svc service.VehicleServiceService
}

func NewVehicleServiceHandler(svc service.VehicleServiceService) *VehicleServiceHandler {
	return &VehicleServiceHandler{svc: svc}
}

// CreatePartsCollection godoc
// @Summary      Create vehicle service part
// @Description  Creates a car body/area item for cosmetic vehicle services, for example bumper, door, hood or roof.
// @Tags         Warehouse Vehicle Service Parts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.PartsCollectionCreateRequest true "Parts collection create payload"
// @Success      201 {object} PartsCollectionCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-parts [post]
func (h *VehicleServiceHandler) CreatePartsCollection(c *gin.Context) {
	var req dto.PartsCollectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	id, err := h.svc.CreatePartsCollection(c.Request.Context(), req)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// GetPartsCollectionByID godoc
// @Summary      Get vehicle service part by ID
// @Description  Returns one car body/area item by ID.
// @Tags         Warehouse Vehicle Service Parts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Parts collection ID"
// @Success      200 {object} PartsCollectionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-parts/{id} [get]
func (h *VehicleServiceHandler) GetPartsCollectionByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	item, err := h.svc.GetPartsCollectionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "service part not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListPartsCollection godoc
// @Summary      List vehicle service parts
// @Description  Returns car body/area items with filters, pagination and sorting.
// @Tags         Warehouse Vehicle Service Parts
// @Produce      json
// @Security     BearerAuth
// @Param        name query string false "Filter by name"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by: created_at, updated_at, name"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} PartsCollectionListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-parts [get]
func (h *VehicleServiceHandler) ListPartsCollection(c *gin.Context) {
	q := dto.PartsCollectionListQuery{Name: c.Query("name"), SortBy: c.Query("sort_by"), Order: c.Query("order")}
	var ok bool
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return
	}

	resp, err := h.svc.ListPartsCollection(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// UpdatePartsCollection godoc
// @Summary      Update vehicle service part
// @Description  Updates car body/area item. Description is optional.
// @Tags         Warehouse Vehicle Service Parts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Parts collection ID"
// @Param        payload body dto.PartsCollectionUpdateRequest true "Parts collection update payload"
// @Success      200 {object} PartsCollectionUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-parts/{id} [put]
func (h *VehicleServiceHandler) UpdatePartsCollection(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.PartsCollectionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}
	updated, err := h.svc.UpdatePartsCollectionByID(c.Request.Context(), id, req)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "service part not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeletePartsCollection godoc
// @Summary      Delete vehicle service part
// @Description  Deletes car body/area item if it is not used by service records.
// @Tags         Warehouse Vehicle Service Parts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Parts collection ID"
// @Success      200 {object} PartsCollectionDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-parts/{id} [delete]
func (h *VehicleServiceHandler) DeletePartsCollection(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deleted, err := h.svc.DeletePartsCollectionByID(c.Request.Context(), id)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "service part not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// CreateServiceType godoc
// @Summary      Create service type
// @Description  Creates a service type, for example polishing, tinting or protective film.
// @Tags         Warehouse Service Types
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.ServiceTypeCreateRequest true "Service type create payload"
// @Success      201 {object} ServiceTypeCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-types [post]
func (h *VehicleServiceHandler) CreateServiceType(c *gin.Context) {
	var req dto.ServiceTypeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}
	id, err := h.svc.CreateServiceType(c.Request.Context(), req)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// GetServiceTypeByID godoc
// @Summary      Get service type by ID
// @Description  Returns one service type by ID.
// @Tags         Warehouse Service Types
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Service type ID"
// @Success      200 {object} ServiceTypeResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-types/{id} [get]
func (h *VehicleServiceHandler) GetServiceTypeByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.svc.GetServiceTypeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "service type not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListServiceTypes godoc
// @Summary      List service types
// @Description  Returns service types with filters, pagination and sorting.
// @Tags         Warehouse Service Types
// @Produce      json
// @Security     BearerAuth
// @Param        name query string false "Filter by name"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by: created_at, updated_at, name"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} ServiceTypeListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-types [get]
func (h *VehicleServiceHandler) ListServiceTypes(c *gin.Context) {
	q := dto.ServiceTypeListQuery{Name: c.Query("name"), SortBy: c.Query("sort_by"), Order: c.Query("order")}
	var ok bool
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return
	}
	resp, err := h.svc.ListServiceTypes(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// UpdateServiceType godoc
// @Summary      Update service type
// @Description  Updates service type. Description is optional.
// @Tags         Warehouse Service Types
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Service type ID"
// @Param        payload body dto.ServiceTypeUpdateRequest true "Service type update payload"
// @Success      200 {object} ServiceTypeUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-types/{id} [put]
func (h *VehicleServiceHandler) UpdateServiceType(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ServiceTypeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}
	updated, err := h.svc.UpdateServiceTypeByID(c.Request.Context(), id, req)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "service type not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeleteServiceType godoc
// @Summary      Delete service type
// @Description  Deletes service type if it is not used by service records.
// @Tags         Warehouse Service Types
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Service type ID"
// @Success      200 {object} ServiceTypeDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/service-types/{id} [delete]
func (h *VehicleServiceHandler) DeleteServiceType(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deleted, err := h.svc.DeleteServiceTypeByID(c.Request.Context(), id)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "service type not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// CreateVehicleService godoc
// @Summary      Create vehicle service record
// @Description  Creates a cosmetic/service maintenance record for a vehicle without spending warehouse stock or installing a new part.
// @Tags         Warehouse Vehicle Services
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.VehicleServiceCreateRequest true "Vehicle service create payload"
// @Success      201 {object} VehicleServiceCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-services [post]
func (h *VehicleServiceHandler) CreateVehicleService(c *gin.Context) {
	var req dto.VehicleServiceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}
	id, err := h.svc.CreateVehicleService(c.Request.Context(), req)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// GetVehicleServiceByID godoc
// @Summary      Get vehicle service record by ID
// @Description  Returns vehicle service record by ID with vehicle, service type and car part information.
// @Tags         Warehouse Vehicle Services
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle service ID"
// @Success      200 {object} VehicleServiceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-services/{id} [get]
func (h *VehicleServiceHandler) GetVehicleServiceByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.svc.GetVehicleServiceByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle service not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListVehicleServices godoc
// @Summary      List vehicle service records
// @Description  Returns vehicle service records with filters by vehicle, service type, car part and date.
// @Tags         Warehouse Vehicle Services
// @Produce      json
// @Security     BearerAuth
// @Param        type_id query int false "Filter by service type ID"
// @Param        part_id query int false "Filter by service part ID"
// @Param        vehicle_id query int false "Filter by vehicle ID"
// @Param        type_name query string false "Filter by service type name"
// @Param        part_name query string false "Filter by service part name"
// @Param        date_from query string false "Service date from, YYYY-MM-DD"
// @Param        date_to query string false "Service date to, YYYY-MM-DD"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by: date, created_at, updated_at, type_id, part_id, vehicle_id"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} VehicleServiceListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-services [get]
func (h *VehicleServiceHandler) ListVehicleServices(c *gin.Context) {
	q := dto.VehicleServiceListQuery{
		TypeName: c.Query("type_name"),
		PartName: c.Query("part_name"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}
	var ok bool
	if q.TypeID, ok = parseInt64Query(c, "type_id", false); !ok {
		return
	}
	if q.PartID, ok = parseInt64Query(c, "part_id", false); !ok {
		return
	}
	if q.VehicleID, ok = parseInt64Query(c, "vehicle_id", false); !ok {
		return
	}
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return
	}
	resp, err := h.svc.ListVehicleServices(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// UpdateVehicleService godoc
// @Summary      Update vehicle service record
// @Description  Updates vehicle service record.
// @Tags         Warehouse Vehicle Services
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle service ID"
// @Param        payload body dto.VehicleServiceUpdateRequest true "Vehicle service update payload"
// @Success      200 {object} VehicleServiceUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-services/{id} [put]
func (h *VehicleServiceHandler) UpdateVehicleService(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.VehicleServiceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}
	updated, err := h.svc.UpdateVehicleServiceByID(c.Request.Context(), id, req)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle service not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeleteVehicleService godoc
// @Summary      Delete vehicle service record
// @Description  Deletes vehicle service record.
// @Tags         Warehouse Vehicle Services
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle service ID"
// @Success      200 {object} VehicleServiceDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/vehicle-services/{id} [delete]
func (h *VehicleServiceHandler) DeleteVehicleService(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deleted, err := h.svc.DeleteVehicleServiceByID(c.Request.Context(), id)
	if err != nil {
		writeVehicleServiceError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle service not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func writeVehicleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrPartsCollectionNameExists):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "service part name already exists"})
	case errors.Is(err, repository.ErrServiceTypeNameExists):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "service type name already exists"})
	case errors.Is(err, repository.ErrVehicleServiceTypeNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "service type not found"})
	case errors.Is(err, repository.ErrVehicleServicePartNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "service part not found"})
	case errors.Is(err, repository.ErrVehicleServiceVehicleNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "vehicle not found"})
	default:
		if strings.Contains(strings.ToLower(err.Error()), "violates foreign key constraint") {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "record is used by another entity"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
