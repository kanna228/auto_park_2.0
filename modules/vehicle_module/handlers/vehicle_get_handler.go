package handlers

import (
	"auto_park/modules/vehicle_module/dto"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func vehiclePhotoURL(c *gin.Context, rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return ""
	}
	rel = strings.TrimLeft(rel, "/")

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	return scheme + "://" + c.Request.Host + path.Join("/static", rel)
}

// GetVehicleByID godoc
// @Summary      Get vehicle by ID
// @Description  Returns vehicle by id.
// @Tags         Vehicles
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle ID"
// @Success      200 {object} VehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/{id} [get]
func (h *VehicleHandler) GetVehicleByID(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid id",
		})
		return
	}

	v, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "vehicle not found",
		})
		return
	}

	v.PhotoURL = vehiclePhotoURL(c, v.PhotoPath)

	for i := range v.Insurances {
		v.Insurances[i].FileURL = vehiclePhotoURL(c, v.Insurances[i].FilePath)
	}

	for i := range v.TechnicalInspections {
		v.TechnicalInspections[i].FileURL = vehiclePhotoURL(c, v.TechnicalInspections[i].FilePath)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    v,
	})
}

// ListVehicles godoc
// @Summary      List vehicles
// @Description  List vehicles with filters, pagination and sorting.
// @Tags         Vehicles
// @Produce      json
// @Security     BearerAuth
// @Param        search       query string false "Search by state number, board number or brand/model"
// @Param        board_number query string false "Filter by board number"
// @Param        state_number query string false "Filter by state number"
// @Param        vin          query string false "Filter by VIN"
// @Param        brand_model  query string false "Filter by brand/model"
// @Param        status_id    query int    false "Filter by vehicle status id"
// @Param        status_name  query string false "Filter by vehicle status name"
// @Param        year_from    query int    false "Manufacture year from"
// @Param        year_to      query int    false "Manufacture year to"
// @Param        driver_id    query int    false "Filter by driver id"
// @Param        limit        query int    false "Limit"  default(50)
// @Param        offset       query int    false "Offset" default(0)
// @Param        sort_by      query string false "Sort by: id, board_number, state_number, manufacture_year, mileage, created_at, status_id, status_name"
// @Param        order        query string false "Order: asc or desc"
// @Success      200 {object} VehicleListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles [get]
func (h *VehicleHandler) ListVehicles(c *gin.Context) {
	q := dto.VehicleListQuery{
		Search:      c.Query("search"),
		BoardNumber: c.Query("board_number"),
		StateNumber: c.Query("state_number"),
		VIN:         c.Query("vin"),
		BrandModel:  c.Query("brand_model"),
		StatusName:  c.Query("status_name"),
		SortBy:      c.Query("sort_by"),
		Order:       c.Query("order"),
	}

	if s := strings.TrimSpace(c.Query("status_id")); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
			q.StatusID = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid status_id"})
			return
		}
	}

	if s := strings.TrimSpace(c.Query("year_from")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			q.ManufactureYearFrom = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid year_from"})
			return
		}
	}
	if s := strings.TrimSpace(c.Query("year_to")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			q.ManufactureYearTo = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid year_to"})
			return
		}
	}

	if s := strings.TrimSpace(c.Query("driver_id")); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
			q.DriverID = &n
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid driver_id"})
			return
		}
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	for i := range resp.Items {
		resp.Items[i].PhotoURL = vehiclePhotoURL(c, resp.Items[i].PhotoPath)

		for j := range resp.Items[i].Insurances {
			resp.Items[i].Insurances[j].FileURL = vehiclePhotoURL(c, resp.Items[i].Insurances[j].FilePath)
		}

		for j := range resp.Items[i].TechnicalInspections {
			resp.Items[i].TechnicalInspections[j].FileURL = vehiclePhotoURL(c, resp.Items[i].TechnicalInspections[j].FilePath)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
