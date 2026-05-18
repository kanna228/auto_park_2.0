package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/vehicle_module/dto"

	"github.com/gin-gonic/gin"
)

func (h *TireHandler) ListVehicleTires(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}

	items, err := h.svc.ListVehicleTires(c.Request.Context(), vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *TireHandler) CreateVehicleTire(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}

	var req dto.VehicleTireCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.CreateVehicleTire(c.Request.Context(), vehicleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

func (h *TireHandler) UpdateVehicleTire(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}
	tireID, ok := parseTirePathID(c)
	if !ok {
		return
	}

	var req dto.VehicleTireUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	updated, err := h.svc.UpdateVehicleTire(c.Request.Context(), vehicleID, tireID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": tireID}})
}

func (h *TireHandler) DeleteVehicleTire(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}
	tireID, ok := parseTirePathID(c)
	if !ok {
		return
	}

	updated, err := h.svc.DetachVehicleTire(c.Request.Context(), vehicleID, tireID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "tire not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": tireID}})
}

func parseVehiclePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid vehicle id"})
		return 0, false
	}
	return id, true
}

func parseTirePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("tid")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid tire id"})
		return 0, false
	}
	return id, true
}
