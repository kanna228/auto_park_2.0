package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/service"

	"github.com/gin-gonic/gin"
)

type WaybillRoutePointHandler struct {
	svc service.WaybillRoutePointService
}

func NewWaybillRoutePointHandler(svc service.WaybillRoutePointService) *WaybillRoutePointHandler {
	return &WaybillRoutePointHandler{svc: svc}
}

func (h *WaybillRoutePointHandler) List(c *gin.Context) {
	waybillID, ok := parseWaybillPathID(c)
	if !ok {
		return
	}
	items, err := h.svc.List(c.Request.Context(), waybillID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *WaybillRoutePointHandler) Create(c *gin.Context) {
	waybillID, ok := parseWaybillPathID(c)
	if !ok {
		return
	}
	var req dto.WaybillRoutePointCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	id, err := h.svc.Create(c.Request.Context(), waybillID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id}})
}

func (h *WaybillRoutePointHandler) Update(c *gin.Context) {
	waybillID, ok := parseWaybillPathID(c)
	if !ok {
		return
	}
	routeID, ok := parseRoutePathID(c)
	if !ok {
		return
	}
	var req dto.WaybillRoutePointUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	updated, err := h.svc.Update(c.Request.Context(), waybillID, routeID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "route point not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": routeID}})
}

func (h *WaybillRoutePointHandler) Delete(c *gin.Context) {
	waybillID, ok := parseWaybillPathID(c)
	if !ok {
		return
	}
	routeID, ok := parseRoutePathID(c)
	if !ok {
		return
	}
	deleted, err := h.svc.Delete(c.Request.Context(), waybillID, routeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "route point not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": routeID}})
}

func parseWaybillPathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid waybill id"})
		return 0, false
	}
	return id, true
}

func parseRoutePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("rid")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid route id"})
		return 0, false
	}
	return id, true
}
