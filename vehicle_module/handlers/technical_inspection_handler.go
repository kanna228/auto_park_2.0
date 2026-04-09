package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/vehicle_module/dto"
	"auto_park/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

type TechnicalInspectionHandler struct {
	svc service.TechnicalInspectionService
}

func NewTechnicalInspectionHandler(svc service.TechnicalInspectionService) *TechnicalInspectionHandler {
	return &TechnicalInspectionHandler{svc: svc}
}

// CreateTechnicalInspection godoc
// @Summary      Create technical inspection
// @Description  Creates a new technical inspection record.
// @Tags         Vehicle Technical Inspection
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.TechnicalInspectionCreateRequest true "Technical inspection create payload"
// @Success      201 {object} TechnicalInspectionCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections [post]
func (h *TechnicalInspectionHandler) CreateTechnicalInspection(c *gin.Context) {
	var req dto.TechnicalInspectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": dto.TechnicalInspectionCreateResponse{
			ID: id,
		},
	})
}

// GetTechnicalInspectionByID godoc
// @Summary      Get technical inspection by ID
// @Description  Returns technical inspection by id.
// @Tags         Vehicle Technical Inspection
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Technical inspection ID"
// @Success      200 {object} TechnicalInspectionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections/{id} [get]
func (h *TechnicalInspectionHandler) GetTechnicalInspectionByID(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "technical inspection not found"})
		return
	}

	item.FileURL = vehiclePhotoURL(c, item.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// ListTechnicalInspections godoc
// @Summary      List technical inspections
// @Description  Returns technical inspection list with filters, pagination and sorting.
// @Tags         Vehicle Technical Inspection
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id query int false "Filter by vehicle ID"
// @Param        is_active  query bool false "Filter by active flag"
// @Param        name       query string false "Filter by name"
// @Param        limit      query int false "Limit" default(50)
// @Param        offset     query int false "Offset" default(0)
// @Param        sort_by    query string false "Sort by: id, vehicle_id, name, start_date, end_date, created_at"
// @Param        order      query string false "Order: asc or desc"
// @Success      200 {object} TechnicalInspectionListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections [get]
func (h *TechnicalInspectionHandler) ListTechnicalInspections(c *gin.Context) {
	q := dto.TechnicalInspectionListQuery{
		Name:   c.Query("name"),
		SortBy: c.Query("sort_by"),
		Order:  c.Query("order"),
	}

	if s := strings.TrimSpace(c.Query("vehicle_id")); s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid vehicle_id"})
			return
		}
		q.VehicleID = &n
	}

	if s := strings.TrimSpace(c.Query("is_active")); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid is_active"})
			return
		}
		q.IsActive = &b
	}

	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid limit"})
			return
		}
		q.Limit = n
	} else {
		q.Limit = 50
	}

	if s := strings.TrimSpace(c.Query("offset")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid offset"})
			return
		}
		q.Offset = n
	} else {
		q.Offset = 0
	}

	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	for i := range resp.Items {
		resp.Items[i].FileURL = vehiclePhotoURL(c, resp.Items[i].FilePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// UpdateTechnicalInspection godoc
// @Summary      Update technical inspection
// @Description  Updates technical inspection by id.
// @Tags         Vehicle Technical Inspection
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Technical inspection ID"
// @Param        payload body dto.TechnicalInspectionUpdateRequest true "Technical inspection update payload"
// @Success      200 {object} TechnicalInspectionUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections/{id} [put]
func (h *TechnicalInspectionHandler) UpdateTechnicalInspection(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.TechnicalInspectionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	ok, err := h.svc.UpdateByID(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "technical inspection not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}

// DeleteTechnicalInspection godoc
// @Summary      Delete technical inspection
// @Description  Deletes technical inspection by id. Linked file is also removed if it exists.
// @Tags         Vehicle Technical Inspection
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Technical inspection ID"
// @Success      200 {object} TechnicalInspectionDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections/{id} [delete]
func (h *TechnicalInspectionHandler) DeleteTechnicalInspection(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	ok, err := h.svc.DeleteByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "technical inspection not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}

// UploadTechnicalInspectionFile godoc
// @Summary      Upload or replace technical inspection file
// @Description  Uploads one file for technical inspection and replaces the previous one if it exists.
// @Tags         Vehicle Technical Inspection
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Technical inspection ID"
// @Param        file formData file true "Technical inspection file"
// @Success      200 {object} TechnicalInspectionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections/{id}/file [post]
func (h *TechnicalInspectionHandler) UploadTechnicalInspectionFile(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "file is required"})
		return
	}

	item, err := h.svc.UploadFile(c.Request.Context(), id, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "technical inspection not found"})
		return
	}

	item.FileURL = vehiclePhotoURL(c, item.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// DeleteTechnicalInspectionFile godoc
// @Summary      Delete technical inspection file
// @Description  Deletes technical inspection file and clears file_path.
// @Tags         Vehicle Technical Inspection
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Technical inspection ID"
// @Success      200 {object} TechnicalInspectionResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/technical-inspections/{id}/file [delete]
func (h *TechnicalInspectionHandler) DeleteTechnicalInspectionFile(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	item, err := h.svc.DeleteFile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "technical inspection not found"})
		return
	}

	item.FileURL = vehiclePhotoURL(c, item.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}
