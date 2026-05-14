package dto

type IncidentCreateRequest struct {
	IncidentTypeID  int64  `json:"incident_type_id" binding:"required"`
	VehicleID       int64  `json:"vehicle_id" binding:"required"`
	DriverID        int64  `json:"driver_id" binding:"required"`
	MechanicID      int64  `json:"mechanic_id" binding:"required"`
	MechanicShiftID int64  `json:"mechanic_shift_id" binding:"required"`
	TripsheetID     *int64 `json:"tripsheet_id,omitempty"`
	Date            string `json:"date" binding:"required"`
	Time            string `json:"time" binding:"required"`
	Location        string `json:"location" binding:"required"`
	Description     string `json:"description"`
}

type IncidentUpdateRequest struct {
	IncidentTypeID  int64  `json:"incident_type_id" binding:"required"`
	VehicleID       int64  `json:"vehicle_id" binding:"required"`
	DriverID        int64  `json:"driver_id" binding:"required"`
	MechanicID      int64  `json:"mechanic_id" binding:"required"`
	MechanicShiftID int64  `json:"mechanic_shift_id" binding:"required"`
	TripsheetID     *int64 `json:"tripsheet_id,omitempty"`
	Date            string `json:"date" binding:"required"`
	Time            string `json:"time" binding:"required"`
	Location        string `json:"location" binding:"required"`
	Description     string `json:"description"`
}

type IncidentCreateResponse struct {
	ID int64 `json:"id"`
}

type IncidentTripsheetResponse struct {
	ID                         int64   `json:"id"`
	TripsheetNumber            string  `json:"tripsheet_number"`
	TripsheetDate              string  `json:"tripsheet_date"`
	VehicleBrand               *string `json:"vehicle_brand,omitempty"`
	VehiclePlateNumber         string  `json:"vehicle_plate_number"`
	DriverID                   *int64  `json:"driver_id,omitempty"`
	DriverLastName             *string `json:"driver_last_name,omitempty"`
	DriverFirstName            *string `json:"driver_first_name,omitempty"`
	DriverMiddleName           *string `json:"driver_middle_name,omitempty"`
	StatusID                   int64   `json:"status_id"`
	StatusName                 string  `json:"status_name"`
	StartTime                  *string `json:"start_time,omitempty"`
	EndTime                    *string `json:"end_time,omitempty"`
	MileageStart               int     `json:"mileage_start"`
	MileageEnd                 int     `json:"mileage_end"`
	FuelStart                  int     `json:"fuel_start"`
	FuelIssued                 int     `json:"fuel_issued"`
	FuelConsumptionTheoretical int     `json:"fuel_consumption_theoretical"`
	FuelConsumptionActual      int     `json:"fuel_consumption_actual"`
}

type IncidentMechanicShiftResponse struct {
	ID               int64   `json:"id"`
	UserID           int64   `json:"user_id"`
	ShiftDate        string  `json:"shift_date"`
	TimeFrom         string  `json:"time_from"`
	TimeTo           *string `json:"time_to,omitempty"`
	Comment          *string `json:"comment,omitempty"`
	IsActive         bool    `json:"is_active"`
	MechanicEmail    *string `json:"mechanic_email,omitempty"`
	MechanicFullName *string `json:"mechanic_full_name,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type IncidentResponse struct {
	ID                 int64                          `json:"id"`
	IncidentTypeID     int64                          `json:"incident_type_id"`
	IncidentTypeName   string                         `json:"incident_type_name"`
	VehicleID          int64                          `json:"vehicle_id"`
	VehicleStateNumber string                         `json:"vehicle_state_number"`
	DriverID           int64                          `json:"driver_id"`
	DriverFullName     string                         `json:"driver_full_name"`
	MechanicID         int64                          `json:"mechanic_id"`
	MechanicFullName   string                         `json:"mechanic_full_name"`
	MechanicShiftID    *int64                         `json:"mechanic_shift_id,omitempty"`
	MechanicShift      *IncidentMechanicShiftResponse `json:"mechanic_shift,omitempty"`
	TripsheetID        *int64                         `json:"tripsheet_id,omitempty"`
	Tripsheet          *IncidentTripsheetResponse     `json:"tripsheet,omitempty"`
	Date               string                         `json:"date"`
	Time               string                         `json:"time"`
	Location           string                         `json:"location"`
	Description        string                         `json:"description,omitempty"`
	CreatedAt          string                         `json:"created_at"`
	UpdatedAt          string                         `json:"updated_at"`
}

type IncidentListQuery struct {
	IncidentTypeID  *int64 `form:"incident_type_id"`
	VehicleID       *int64 `form:"vehicle_id"`
	DriverID        *int64 `form:"driver_id"`
	MechanicID      *int64 `form:"mechanic_id"`
	MechanicShiftID *int64 `form:"mechanic_shift_id"`
	TripsheetID     *int64 `form:"tripsheet_id"`
	DateFrom        string `form:"date_from"`
	DateTo          string `form:"date_to"`
	Limit           int    `form:"limit"`
	Offset          int    `form:"offset"`
	SortBy          string `form:"sort_by"`
	Order           string `form:"order"`
}

type IncidentTypeResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
