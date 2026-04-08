package handlers

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"auto_park/user_module/dto"
	"auto_park/user_module/repository"
	"auto_park/user_module/service"

	"github.com/gin-gonic/gin"
)

// =======================
// SWAGGER RESPONSE MODELS
// =======================

type DriverDTO struct {
	ID         int64  `json:"id" example:"1"`
	IIN        string `json:"iin" example:"001122334455"`
	Name       string `json:"name" example:"Dias"`
	Surname    string `json:"surname" example:"Arnold"`
	Middlename string `json:"middlename,omitempty" example:"A."`
	Phone      string `json:"phone,omitempty" example:"+77001234567"`
	Mail       string `json:"mail,omitempty" example:"dias@example.com"`
	PhotoPath  string `json:"photo_path,omitempty" example:"drivers/driver_1_1710000000.jpg"`
	PhotoURL   string `json:"photo_url,omitempty" example:"http://localhost:8080/static/drivers/driver_1_1710000000.jpg"`
	CreatedAt  string `json:"created_at" example:"2026-02-18T12:34:56Z"`
	UpdatedAt  string `json:"updated_at" example:"2026-02-18T12:34:56Z"`
}

type DriverResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    DriverDTO `json:"data"`
}

type DriversListResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    []DriverDTO `json:"data"`
}

type DeleteDriverResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		ID int64 `json:"id" example:"123"`
	} `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid id"`
}

// =======================

type DriversHandler struct {
	svc *service.DriversService
}

func NewDriversHandler(svc *service.DriversService) *DriversHandler {
	return &DriversHandler{svc: svc}
}

func driverPhotoURL(c *gin.Context, rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return ""
	}
	rel = strings.TrimLeft(rel, "/")
	return "http://" + c.Request.Host + path.Join("/static", rel)
}

// Create godoc
// @Summary      Create driver
// @Description  Creates a new driver record (roles: 1,2,3 only). JWT required.
// @Tags         Drivers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        driver body dto.CreateDriverRequest true "Driver payload"
// @Success      201 {object} DriverResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers [post]
func (h *DriversHandler) Create(c *gin.Context) {
	var req dto.CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	drv, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, DriverResponse{
		Success: true,
		Data: DriverDTO{
			ID:         drv.ID,
			IIN:        drv.IIN,
			Name:       drv.Name,
			Surname:    drv.Surname,
			Middlename: drv.Middlename,
			Phone:      drv.Phone,
			Mail:       drv.Mail,
			PhotoPath:  drv.PhotoPath,
			PhotoURL:   driverPhotoURL(c, drv.PhotoPath),
			CreatedAt:  drv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  drv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// List godoc
// @Summary      List drivers
// @Description  Returns paginated list of drivers (any authorized role). JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Limit"  default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} DriversListResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers [get]
func (h *DriversHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	list, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	out := make([]DriverDTO, 0, len(list))
	for i := range list {
		d := list[i]
		out = append(out, DriverDTO{
			ID:         d.ID,
			IIN:        d.IIN,
			Name:       d.Name,
			Surname:    d.Surname,
			Middlename: d.Middlename,
			Phone:      d.Phone,
			Mail:       d.Mail,
			PhotoPath:  d.PhotoPath,
			PhotoURL:   driverPhotoURL(c, d.PhotoPath),
			CreatedAt:  d.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, DriversListResponse{Success: true, Data: out})
}

// GetByID godoc
// @Summary      Get driver by ID
// @Description  Returns driver by id (any authorized role). JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver ID"
// @Success      200 {object} DriverResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers/{id} [get]
func (h *DriversHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	drv, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverResponse{
		Success: true,
		Data: DriverDTO{
			ID:         drv.ID,
			IIN:        drv.IIN,
			Name:       drv.Name,
			Surname:    drv.Surname,
			Middlename: drv.Middlename,
			Phone:      drv.Phone,
			Mail:       drv.Mail,
			PhotoPath:  drv.PhotoPath,
			PhotoURL:   driverPhotoURL(c, drv.PhotoPath),
			CreatedAt:  drv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  drv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// Update godoc
// @Summary      Update driver
// @Description  Updates driver fields by id (roles: 1,2,3 only). JWT required.
// @Tags         Drivers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path int                     true "Driver ID"
// @Param        driver body dto.UpdateDriverRequest true "Fields to update"
// @Success      200 {object} DriverResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers/{id} [put]
func (h *DriversHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	drv, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverResponse{
		Success: true,
		Data: DriverDTO{
			ID:         drv.ID,
			IIN:        drv.IIN,
			Name:       drv.Name,
			Surname:    drv.Surname,
			Middlename: drv.Middlename,
			Phone:      drv.Phone,
			Mail:       drv.Mail,
			PhotoPath:  drv.PhotoPath,
			PhotoURL:   driverPhotoURL(c, drv.PhotoPath),
			CreatedAt:  drv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  drv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// UploadPhoto godoc
// @Summary      Upload or replace driver photo
// @Description  Uploads one photo for a driver and replaces the previous one if it exists (roles: 1,2,3 only). JWT required.
// @Tags         Drivers
// @Accept       mpfd
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver ID"
// @Param        photo formData file true "Driver photo"
// @Success      200 {object} DriverResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers/{id}/photo [post]
func (h *DriversHandler) UploadPhoto(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "photo file is required"})
		return
	}

	drv, err := h.svc.UploadPhoto(c.Request.Context(), id, file)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverResponse{
		Success: true,
		Data: DriverDTO{
			ID:         drv.ID,
			IIN:        drv.IIN,
			Name:       drv.Name,
			Surname:    drv.Surname,
			Middlename: drv.Middlename,
			Phone:      drv.Phone,
			Mail:       drv.Mail,
			PhotoPath:  drv.PhotoPath,
			PhotoURL:   driverPhotoURL(c, drv.PhotoPath),
			CreatedAt:  drv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  drv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// DeletePhoto godoc
// @Summary      Delete driver photo
// @Description  Deletes driver photo and clears photo_path (roles: 1,2,3 only). JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver ID"
// @Success      200 {object} DriverResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers/{id}/photo [delete]
func (h *DriversHandler) DeletePhoto(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	drv, err := h.svc.DeletePhoto(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverResponse{
		Success: true,
		Data: DriverDTO{
			ID:         drv.ID,
			IIN:        drv.IIN,
			Name:       drv.Name,
			Surname:    drv.Surname,
			Middlename: drv.Middlename,
			Phone:      drv.Phone,
			Mail:       drv.Mail,
			PhotoPath:  drv.PhotoPath,
			PhotoURL:   driverPhotoURL(c, drv.PhotoPath),
			CreatedAt:  drv.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  drv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}

// Delete godoc
// @Summary      Delete driver
// @Description  Deletes driver by id (roles: 1,2,3 only). JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver ID"
// @Success      200 {object} DeleteDriverResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers/{id} [delete]
func (h *DriversHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	resp := DeleteDriverResponse{Success: true}
	resp.Data.ID = id
	c.JSON(http.StatusOK, resp)
}
