package handlers

import (
	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TireHandler struct {
	svc service.TireService
}

func NewTireHandler(svc service.TireService) *TireHandler {
	return &TireHandler{svc: svc}
}

// CreateTire godoc
// @Summary      Create tire
// @Description  Creates a new tire record.
// @Tags         Tires
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.TireCreateRequest true "Tire create payload"
// @Success      201 {object} TireCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles/tires [post]
func (h *TireHandler) CreateTire(c *gin.Context) {
	var req dto.TireCreateRequest
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
		"data": dto.TireCreateResponse{
			ID: id,
		},
	})
}
