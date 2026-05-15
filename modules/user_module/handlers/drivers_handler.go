package handlers

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/models"
	"auto_park/modules/user_module/repository"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

// =======================
// SWAGGER RESPONSE MODELS
// =======================

type DriverStatusDTO struct {
	ID          int64  `json:"id" example:"1"`
	Code        string `json:"code" example:"available"`
	Name        string `json:"name" example:"Доступен"`
	Description string `json:"description,omitempty" example:"Водитель доступен и может быть назначен на рейс"`
	CreatedAt   string `json:"created_at" example:"2026-05-14T06:00:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2026-05-14T06:00:00Z"`
}

type DriverDTO struct {
	ID               int64                      `json:"id" example:"1"`
	IIN              string                     `json:"iin" example:"001122334455"`
	Name             string                     `json:"name" example:"Dias"`
	Surname          string                     `json:"surname" example:"Arnold"`
	Middlename       string                     `json:"middlename,omitempty" example:"A."`
	Phone            string                     `json:"phone,omitempty" example:"+77001234567"`
	Mail             string                     `json:"mail,omitempty" example:"dias@example.com"`
	PhotoPath        string                     `json:"photo_path,omitempty" example:"drivers/driver_1_1710000000.jpg"`
	PhotoURL         string                     `json:"photo_url,omitempty" example:"http://localhost:8080/static/drivers/driver_1_1710000000.jpg"`
	BirthDate        *string                    `json:"birth_date,omitempty" example:"1990-11-11"`
	LicenseNumber    string                     `json:"license_number,omitempty" example:"DL-123456"`
	LicenseCategory  string                     `json:"license_category,omitempty" example:"B, C"`
	ExperienceYears  *int                       `json:"experience_years,omitempty" example:"5"`
	StatusID         int64                      `json:"status_id" example:"1"`
	Status           DriverStatusDTO            `json:"status"`
	AssignedVehicles []DriverAssignedVehicleDTO `json:"assigned_vehicles"`
	CreatedAt        string                     `json:"created_at" example:"2026-02-18T12:34:56Z"`
	UpdatedAt        string                     `json:"updated_at" example:"2026-02-18T12:34:56Z"`
}

type DriverAssignedVehicleDTO struct {
	ID          int64  `json:"id" example:"1"`
	BoardNumber string `json:"board_number" example:"55"`
	StateNumber string `json:"state_number" example:"777ABC01"`
	BrandModel  string `json:"brand_model" example:"Toyota Camry"`
	StatusID    int64  `json:"status_id" example:"1"`
	StatusName  string `json:"status_name" example:"В использовании"`
}

type DriverPassportTripsheetDTO struct {
	ID                 int64   `json:"id" example:"1"`
	TripsheetNumber    string  `json:"tripsheet_number" example:"TS-001"`
	TripsheetDate      string  `json:"tripsheet_date" example:"2026-05-14"`
	VehicleID          *int64  `json:"vehicle_id,omitempty" example:"1"`
	VehicleBrand       *string `json:"vehicle_brand,omitempty" example:"Toyota Camry"`
	VehiclePlateNumber string  `json:"vehicle_plate_number" example:"777ABC01"`
	StartTime          *string `json:"start_time,omitempty" example:"2026-05-14T08:00:00Z"`
	EndTime            *string `json:"end_time,omitempty" example:"2026-05-14T18:00:00Z"`
	WorkedHours        float64 `json:"worked_hours" example:"10"`
	TripsCount         int64   `json:"trips_count" example:"5"`
	StatusID           int64   `json:"status_id" example:"1"`
	StatusName         *string `json:"status_name,omitempty" example:"Окончен"`
	CreatedAt          string  `json:"created_at" example:"2026-05-14T08:00:00Z"`
	UpdatedAt          string  `json:"updated_at" example:"2026-05-14T18:00:00Z"`
}

type DriverPassportIncidentDTO struct {
	ID                 int64  `json:"id" example:"1"`
	IncidentTypeID     int64  `json:"incident_type_id" example:"1"`
	IncidentTypeName   string `json:"incident_type_name" example:"ДТП"`
	VehicleID          int64  `json:"vehicle_id" example:"1"`
	VehicleStateNumber string `json:"vehicle_state_number" example:"777ABC01"`
	TripsheetID        *int64 `json:"tripsheet_id,omitempty" example:"1"`
	IncidentDate       string `json:"incident_date" example:"2026-05-14"`
	IncidentTime       string `json:"incident_time" example:"14:30:00"`
	Location           string `json:"location" example:"Астана"`
	Description        string `json:"description,omitempty" example:"Minor accident"`
	CreatedAt          string `json:"created_at" example:"2026-05-14T14:30:00Z"`
	UpdatedAt          string `json:"updated_at" example:"2026-05-14T14:30:00Z"`
}

type DriverPassportDTO struct {
	Driver           DriverDTO                    `json:"driver"`
	Status           string                       `json:"status" example:"Доступен"`
	AssignedVehicles []DriverAssignedVehicleDTO   `json:"assigned_vehicles"`
	TotalWorkedHours float64                      `json:"total_worked_hours" example:"40"`
	IncidentsCount   int64                        `json:"incidents_count" example:"5"`
	Tripsheets       []DriverPassportTripsheetDTO `json:"tripsheets"`
	Incidents        []DriverPassportIncidentDTO  `json:"incidents"`
}

type DriverResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    DriverDTO `json:"data"`
}

type DriverPassportResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    DriverPassportDTO `json:"data"`
}

type DriversListResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    []DriverDTO `json:"data"`
	Total   int64       `json:"total" example:"1"`
	Limit   int         `json:"limit" example:"50"`
	Offset  int         `json:"offset" example:"0"`
}

type DriverStatusListResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    []DriverStatusDTO `json:"data"`
	Total   int64             `json:"total" example:"5"`
	Limit   int               `json:"limit" example:"50"`
	Offset  int               `json:"offset" example:"0"`
}

type DriverStatusUpdateResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    DriverDTO `json:"data"`
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

// ListStatuses godoc
// @Summary      List driver statuses
// @Description  Returns paginated list of driver statuses: available, on_trip, unavailable, vacation, sick_leave. JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Success      200 {object} DriverStatusListResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/driver-statuses [get]
func (h *DriversHandler) ListStatuses(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.svc.ListStatuses(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	out := make([]DriverStatusDTO, 0, len(items))
	for _, item := range items {
		out = append(out, driverStatusToDTO(item))
	}

	c.JSON(http.StatusOK, DriverStatusListResponse{Success: true, Data: out, Total: total, Limit: limit, Offset: offset})
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, DriverResponse{Success: true, Data: driverToDTO(c, *drv)})
}

// List godoc
// @Summary      List drivers
// @Description  Returns paginated list of drivers (any authorized role). JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Limit"  default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        search query string false "Search by full name, IIN, phone or email"
// @Param        status query string false "Filter by status code or name"
// @Param        board_number query string false "Filter by assigned vehicle board number"
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Success      200 {object} DriversListResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers [get]
func (h *DriversHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	q := dto.DriverListQuery{
		Search:      c.Query("search"),
		Status:      c.Query("status"),
		BoardNumber: c.Query("board_number"),
		Limit:       limit,
		Offset:      offset,
	}

	list, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	out := make([]DriverDTO, 0, len(list))
	for i := range list {
		out = append(out, driverToDTO(c, list[i]))
	}

	c.JSON(http.StatusOK, DriversListResponse{Success: true, Data: out, Total: total, Limit: limit, Offset: offset})
}

func (h *DriversHandler) AssignVehicle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}
	var req dto.AssignDriverVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	if err := h.svc.AssignVehicle(c.Request.Context(), id, req.VehicleID); err != nil {
		writeDriverVehicleAssignmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"driver_id": id, "vehicle_id": req.VehicleID}})
}

func (h *DriversHandler) UnassignVehicle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}
	vehicleID, err := strconv.ParseInt(c.Param("vehicle_id"), 10, 64)
	if err != nil || vehicleID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid vehicle_id"})
		return
	}
	if err := h.svc.UnassignVehicle(c.Request.Context(), id, vehicleID); err != nil {
		writeDriverVehicleAssignmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"driver_id": id, "vehicle_id": vehicleID}})
}

// GetByID godoc
// @Summary      Get driver passport by ID
// @Description  Returns full driver passport by id: base data, photo, license data, assigned vehicles, total worked hours, incidents count, latest tripsheets and latest incidents. JWT required.
// @Tags         Drivers
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver ID"
// @Success      200 {object} DriverPassportResponse
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

	passport, err := h.svc.GetPassport(c.Request.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverPassportResponse{Success: true, Data: driverPassportToDTO(c, *passport)})
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverResponse{Success: true, Data: driverToDTO(c, *drv)})
}

// UpdateStatus godoc
// @Summary      Update driver status
// @Description  Updates only driver status by id. Example statuses: 1=Доступен, 2=На выезде. Roles: 1,2,3 only. JWT required.
// @Tags         Drivers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Driver ID"
// @Param        payload body dto.UpdateDriverStatusRequest true "Driver status update payload"
// @Success      200 {object} DriverStatusUpdateResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/drivers/{id}/status [patch]
func (h *DriversHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.UpdateDriverStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	drv, err := h.svc.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DriverStatusUpdateResponse{Success: true, Data: driverToDTO(c, *drv)})
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

	c.JSON(http.StatusOK, DriverResponse{Success: true, Data: driverToDTO(c, *drv)})
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

	c.JSON(http.StatusOK, DriverResponse{Success: true, Data: driverToDTO(c, *drv)})
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

func driverStatusToDTO(status models.DriverStatus) DriverStatusDTO {
	return DriverStatusDTO{
		ID:          status.ID,
		Code:        status.Code,
		Name:        status.Name,
		Description: status.Description,
		CreatedAt:   status.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   status.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func driverToDTO(c *gin.Context, drv models.Driver) DriverDTO {
	var birthDate *string
	if drv.BirthDate != nil {
		v := drv.BirthDate.Format("2006-01-02")
		birthDate = &v
	}

	assignedVehicles := make([]DriverAssignedVehicleDTO, 0, len(drv.AssignedVehicles))
	for _, vehicle := range drv.AssignedVehicles {
		assignedVehicles = append(assignedVehicles, DriverAssignedVehicleDTO{
			ID:          vehicle.ID,
			BoardNumber: vehicle.BoardNumber,
			StateNumber: vehicle.StateNumber,
			BrandModel:  vehicle.BrandModel,
			StatusID:    vehicle.StatusID,
			StatusName:  vehicle.StatusName,
		})
	}

	return DriverDTO{
		ID:               drv.ID,
		IIN:              drv.IIN,
		Name:             drv.Name,
		Surname:          drv.Surname,
		Middlename:       drv.Middlename,
		Phone:            drv.Phone,
		Mail:             drv.Mail,
		PhotoPath:        drv.PhotoPath,
		PhotoURL:         driverPhotoURL(c, drv.PhotoPath),
		BirthDate:        birthDate,
		LicenseNumber:    drv.LicenseNumber,
		LicenseCategory:  drv.LicenseCategory,
		ExperienceYears:  drv.ExperienceYears,
		StatusID:         drv.StatusID,
		Status:           driverStatusToDTO(drv.Status),
		AssignedVehicles: assignedVehicles,
		CreatedAt:        drv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        drv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func driverPassportToDTO(c *gin.Context, passport models.DriverPassport) DriverPassportDTO {
	assignedVehicles := make([]DriverAssignedVehicleDTO, 0, len(passport.AssignedVehicles))
	for _, vehicle := range passport.AssignedVehicles {
		assignedVehicles = append(assignedVehicles, DriverAssignedVehicleDTO{
			ID:          vehicle.ID,
			BoardNumber: vehicle.BoardNumber,
			StateNumber: vehicle.StateNumber,
			BrandModel:  vehicle.BrandModel,
			StatusID:    vehicle.StatusID,
			StatusName:  vehicle.StatusName,
		})
	}

	tripsheets := make([]DriverPassportTripsheetDTO, 0, len(passport.Tripsheets))
	for _, item := range passport.Tripsheets {
		var startTime *string
		if item.StartTime != nil {
			v := item.StartTime.Format("15:04")
			startTime = &v
		}
		var endTime *string
		if item.EndTime != nil {
			v := item.EndTime.Format("15:04")
			endTime = &v
		}

		tripsheets = append(tripsheets, DriverPassportTripsheetDTO{
			ID:                 item.ID,
			TripsheetNumber:    item.TripsheetNumber,
			TripsheetDate:      item.TripsheetDate,
			VehicleID:          item.VehicleID,
			VehicleBrand:       item.VehicleBrand,
			VehiclePlateNumber: item.VehiclePlateNumber,
			StartTime:          startTime,
			EndTime:            endTime,
			WorkedHours:        item.WorkedHours,
			TripsCount:         item.TripsCount,
			StatusID:           item.StatusID,
			StatusName:         item.StatusName,
			CreatedAt:          item.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:          item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	incidents := make([]DriverPassportIncidentDTO, 0, len(passport.Incidents))
	for _, item := range passport.Incidents {
		incidents = append(incidents, DriverPassportIncidentDTO{
			ID:                 item.ID,
			IncidentTypeID:     item.IncidentTypeID,
			IncidentTypeName:   item.IncidentTypeName,
			VehicleID:          item.VehicleID,
			VehicleStateNumber: item.VehicleStateNumber,
			TripsheetID:        item.TripsheetID,
			IncidentDate:       item.IncidentDate,
			IncidentTime:       item.IncidentTime,
			Location:           item.Location,
			Description:        item.Description,
			CreatedAt:          item.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:          item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return DriverPassportDTO{
		Driver:           driverToDTO(c, passport.Driver),
		Status:           passport.Status,
		AssignedVehicles: assignedVehicles,
		TotalWorkedHours: passport.TotalWorkedHours,
		IncidentsCount:   passport.IncidentsCount,
		Tripsheets:       tripsheets,
		Incidents:        incidents,
	}
}

func writeDriverVehicleAssignmentError(c *gin.Context, err error) {
	switch {
	case err == repository.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "driver not found"})
	case err == repository.ErrVehicleNotFound:
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "vehicle not found"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
