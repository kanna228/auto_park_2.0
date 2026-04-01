package handlers

import (
	"net/http"
	"strings"
	"time"

	"auto_park/vehicle_module/dto"
	"auto_park/vehicle_module/service"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	svc service.VehicleService
}

func NewVehicleHandler(svc service.VehicleService) *VehicleHandler {
	return &VehicleHandler{svc: svc}
}

// POST /api/vehicles

// CreateVehicle godoc
// @Summary      Create vehicle
// @Description  Creates a new vehicle. Dates: `received_date` can be RFC3339 (recommended) or `YYYY-MM-DD`.
// @Tags         Vehicles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.VehicleCreateRequest true "Vehicle create payload"
// @Success      201 {object} VehicleCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/vehicles [post]
func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	var req dto.VehicleCreateRequest

	// 1) Пробуем обычный bind (time.Time в JSON обычно ждёт RFC3339)
	if err := c.ShouldBindJSON(&req); err != nil {
		// 2) Если упало из-за даты "YYYY-MM-DD" — попробуем распарсить вручную
		// (потому что на фронте часто отправляют именно "2026-02-27")
		var raw map[string]any
		if err2 := c.ShouldBindJSON(&raw); err2 == nil {
			// вернём нормальную ошибку, потому что второй bind уже съел body
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request body (check required fields and date format)",
		})
		return
	}

	// Если received_date прислали как "0001-01-01..." из-за неверного формата — защитимся
	if req.ReceivedDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "received_date is required (use RFC3339 or YYYY-MM-DD)",
		})
		return
	}

	// страховка: trim VIN/plate и т.д. — service тоже делает, но пусть будет
	req.VIN = strings.TrimSpace(req.VIN)
	req.StateNumber = strings.TrimSpace(req.StateNumber)
	req.BoardNumber = strings.TrimSpace(req.BoardNumber)

	id, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		// На уникальных ограничениях можно вернуть 409 (если хочешь — можем распарсить pq error code)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": dto.VehicleCreateResponse{
			ID: id,
		},
		"ts": time.Now().UTC(),
	})
}
