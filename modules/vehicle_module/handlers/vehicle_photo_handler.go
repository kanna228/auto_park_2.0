package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// UploadVehiclePhoto godoc
// @Summary      Upload or replace vehicle photo
// @Description  Uploads one photo for a vehicle and replaces the previous one if it exists.
// @Tags         Vehicles
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle ID"
// @Param        photo formData file true "Vehicle photo"
// @Success      200 {object} VehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/{id}/photo [post]
func (h *VehicleHandler) UploadVehiclePhoto(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "photo file is required"})
		return
	}

	v, err := h.svc.UploadPhoto(c.Request.Context(), id, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle not found"})
		return
	}

	v.PhotoURL = vehiclePhotoURL(c, v.PhotoPath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    v,
	})
}

// UpdateVehiclePhoto godoc
// @Summary      Update vehicle photo
// @Description  Replaces vehicle photo.
// @Tags         Vehicles
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Vehicle ID"
// @Param        photo formData file true "Vehicle photo"
// @Success      200 {object} VehicleResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/{id}/photo [put]
func (h *VehicleHandler) UpdateVehiclePhoto(c *gin.Context) {
	h.UploadVehiclePhoto(c)
}

// DeleteVehiclePhoto godoc
// @Summary      Delete vehicle photo
// @Description  Deletes vehicle photo and clears photo_path.
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
// @Router       /api/vehicles/{id}/photo [delete]
func (h *VehicleHandler) DeleteVehiclePhoto(c *gin.Context) {
	idStr := strings.TrimSpace(c.Param("id"))
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	v, err := h.svc.DeletePhoto(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if v == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle not found"})
		return
	}

	v.PhotoURL = vehiclePhotoURL(c, v.PhotoPath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    v,
	})
}
