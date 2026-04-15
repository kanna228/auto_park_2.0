package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

type InsuranceHandler struct {
	svc service.InsuranceService
}

func NewInsuranceHandler(svc service.InsuranceService) *InsuranceHandler {
	return &InsuranceHandler{svc: svc}
}

// CreateInsurance godoc
// @Summary      Create insurance
// @Description  Creates a new insurance record.
// @Tags         Vehicle Insurance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.InsuranceCreateRequest true "Insurance create payload"
// @Success      201 {object} InsuranceCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance [post]
func (h *InsuranceHandler) CreateInsurance(c *gin.Context) {
	var req dto.InsuranceCreateRequest
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
		"data": dto.InsuranceCreateResponse{
			ID: id,
		},
	})
}

// GetInsuranceByID godoc
// @Summary      Get insurance by ID
// @Description  Returns insurance by id.
// @Tags         Vehicle Insurance
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Insurance ID"
// @Success      200 {object} InsuranceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance/{id} [get]
func (h *InsuranceHandler) GetInsuranceByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "insurance not found"})
		return
	}

	item.FileURL = vehiclePhotoURL(c, item.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// ListInsurance godoc
// @Summary      List insurance
// @Description  Returns insurance list with filters, pagination and sorting.
// @Tags         Vehicle Insurance
// @Produce      json
// @Security     BearerAuth
// @Param        vehicle_id query int false "Filter by vehicle ID"
// @Param        is_active  query bool false "Filter by active flag"
// @Param        name       query string false "Filter by name"
// @Param        limit      query int false "Limit" default(50)
// @Param        offset     query int false "Offset" default(0)
// @Param        sort_by    query string false "Sort by: id, vehicle_id, name, start_date, end_date, created_at"
// @Param        order      query string false "Order: asc or desc"
// @Success      200 {object} InsuranceListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance [get]
func (h *InsuranceHandler) ListInsurance(c *gin.Context) {
	q := dto.InsuranceListQuery{
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

// UpdateInsurance godoc
// @Summary      Update insurance
// @Description  Updates insurance by id.
// @Tags         Vehicle Insurance
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Insurance ID"
// @Param        payload body dto.InsuranceUpdateRequest true "Insurance update payload"
// @Success      200 {object} InsuranceUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance/{id} [put]
func (h *InsuranceHandler) UpdateInsurance(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.InsuranceUpdateRequest
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "insurance not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}

// DeleteInsurance godoc
// @Summary      Delete insurance
// @Description  Deletes insurance by id. Linked file is also removed if it exists.
// @Tags         Vehicle Insurance
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Insurance ID"
// @Success      200 {object} InsuranceDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance/{id} [delete]
func (h *InsuranceHandler) DeleteInsurance(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "insurance not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": id},
	})
}

// UploadInsuranceFile godoc
// @Summary      Upload or replace insurance file
// @Description  Uploads one file for insurance and replaces the previous one if it exists.
// @Tags         Vehicle Insurance
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Insurance ID"
// @Param        file formData file true "Insurance file"
// @Success      200 {object} InsuranceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance/{id}/file [post]
func (h *InsuranceHandler) UploadInsuranceFile(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "insurance not found"})
		return
	}

	item.FileURL = vehiclePhotoURL(c, item.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}

// DeleteInsuranceFile godoc
// @Summary      Delete insurance file
// @Description  Deletes insurance file and clears file_path.
// @Tags         Vehicle Insurance
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Insurance ID"
// @Success      200 {object} InsuranceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/insurance/{id}/file [delete]
func (h *InsuranceHandler) DeleteInsuranceFile(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "insurance not found"})
		return
	}

	item.FileURL = vehiclePhotoURL(c, item.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    item,
	})
}
