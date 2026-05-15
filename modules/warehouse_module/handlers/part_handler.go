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

type PartHandler struct {
	svc service.PartService
}

func NewPartHandler(svc service.PartService) *PartHandler {
	return &PartHandler{svc: svc}
}

// CreatePart godoc
// @Summary      Create part
// @Description  Creates a new part in the warehouse catalog. Only warehouse manager can create parts.
// @Tags         Warehouse Parts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.PartCreateRequest true "Part create payload"
// @Success      201 {object} PartCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/parts [post]
func (h *PartHandler) CreatePart(c *gin.Context) {
	var req dto.PartCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPartIDExists):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part_id already exists"})
		case errors.Is(err, repository.ErrPartNameExists):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part name already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// GetPartByID godoc
// @Summary      Get part by ID
// @Description  Returns part information by ID. Accessible to admin, duty mechanic and warehouse manager.
// @Tags         Warehouse Parts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part ID"
// @Success      200 {object} PartResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/parts/{id} [get]
func (h *PartHandler) GetPartByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (h *PartHandler) GetPartsSummary(c *gin.Context) {
	resp, err := h.svc.Summary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *PartHandler) ListPartMovements(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	limit, ok := parseIntQuery(c, "limit", false)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}
	resp, err := h.svc.ListMovements(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ListParts godoc
// @Summary      List parts
// @Description  Returns warehouse parts with filters, pagination and sorting. Accessible to admin, duty mechanic and warehouse manager.
// @Tags         Warehouse Parts
// @Produce      json
// @Security     BearerAuth
// @Param        part_id query string false "Filter by part ID"
// @Param        name query string false "Filter by part name"
// @Param        category query string false "Filter by category"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by field: created_at, updated_at, part_id, name, quantity, category"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} PartListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/parts [get]
func (h *PartHandler) ListParts(c *gin.Context) {
	q := dto.PartListQuery{
		PartID:   c.Query("part_id"),
		Name:     c.Query("name"),
		Category: c.Query("category"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid limit"})
			return
		}
		q.Limit = n
	}
	if s := strings.TrimSpace(c.Query("offset")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid offset"})
			return
		}
		q.Offset = n
	}

	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *PartHandler) ExportParts(c *gin.Context) {
	q := dto.PartListQuery{
		PartID:   c.Query("part_id"),
		Name:     c.Query("name"),
		Category: c.Query("category"),
		SortBy:   c.Query("sort_by"),
		Order:    c.Query("order"),
	}
	data, filename, err := h.svc.Export(c.Request.Context(), q, c.DefaultQuery("format", "xlsx"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// UpdatePart godoc
// @Summary      Update part
// @Description  Updates part fields such as quantity, dimensions, manufacturer and category. Only warehouse manager can update parts.
// @Tags         Warehouse Parts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part ID"
// @Param        payload body dto.PartUpdateRequest true "Part update payload"
// @Success      200 {object} PartUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/parts/{id} [put]
func (h *PartHandler) UpdatePart(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.PartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateByID(c.Request.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPartNameExists):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part name already exists"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		}
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeletePart godoc
// @Summary      Delete part
// @Description  Deletes part from the warehouse catalog. Only admin can delete parts.
// @Tags         Warehouse Parts
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part ID"
// @Success      200 {object} PartDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/parts/{id} [delete]
func (h *PartHandler) DeletePart(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	deleted, err := h.svc.DeleteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func (h *PartHandler) CreatePartArrival(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}
	var req dto.PartArrivalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	id, err := h.svc.CreateArrival(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

func (h *PartHandler) ListPartArrivals(c *gin.Context) {
	limit, ok := parseIntQuery(c, "limit", false)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}
	resp, err := h.svc.ListArrivals(c.Request.Context(), c.Query("date_from"), c.Query("date_to"), c.Query("status"), limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *PartHandler) AcceptPartArrival(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}
	accepted, err := h.svc.AcceptArrival(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !accepted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part arrival not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func (h *PartHandler) GetPartArrivalsSummary(c *gin.Context) {
	resp, err := h.svc.ArrivalSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return 0, false
	}
	return id, true
}
