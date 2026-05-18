package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

type VehicleDocumentHandler struct {
	svc service.VehicleDocumentService
}

func NewVehicleDocumentHandler(svc service.VehicleDocumentService) *VehicleDocumentHandler {
	return &VehicleDocumentHandler{svc: svc}
}

func (h *VehicleDocumentHandler) List(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}

	items, err := h.svc.List(c.Request.Context(), vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *VehicleDocumentHandler) Create(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}

	var req dto.VehicleDocumentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	id, err := h.svc.Create(c.Request.Context(), vehicleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

func (h *VehicleDocumentHandler) Delete(c *gin.Context) {
	vehicleID, ok := parseVehiclePathID(c)
	if !ok {
		return
	}
	documentID, ok := parseDocumentPathID(c)
	if !ok {
		return
	}

	deleted, err := h.svc.Delete(c.Request.Context(), vehicleID, documentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "document not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": documentID}})
}

func parseDocumentPathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("doc_id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid document id"})
		return 0, false
	}
	return id, true
}
