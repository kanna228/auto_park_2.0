package handlers

import (
	"net/http"

	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/service"

	"github.com/gin-gonic/gin"
)

type TripsheetHandler struct {
	svc service.TripsheetService
}

func NewTripsheetHandler(svc service.TripsheetService) *TripsheetHandler {
	return &TripsheetHandler{
		svc: svc,
	}
}

// Create godoc
// @Summary      Create tripsheet
// @Description  Creates a new tripsheet.
// @Tags         Tripsheets
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateTripsheetRequest true "Tripsheet create request"
// @Success      201 {object} TripsheetCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Router       /api/tripsheet [post]
func (h *TripsheetHandler) Create(c *gin.Context) {
	var req dto.CreateTripsheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    resp,
	})
}
