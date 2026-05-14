package dto

// CreateTripsheetRequest represents request body for creating a tripsheet.
type CreateTripsheetRequest struct {
	TripsheetNumber    string  `json:"tripsheet_number" example:"PL-0002"`
	TripsheetDate      string  `json:"tripsheet_date" example:"2026-03-18"`
	VehicleID          *int64  `json:"vehicle_id" example:"12"`
	VehicleBrand       *string `json:"vehicle_brand" example:"Toyota Camry"`
	VehiclePlateNumber string  `json:"vehicle_plate_number" example:"123ABC01"`

	DriverLastName   *string `json:"driver_last_name" example:"Ivanov"`
	DriverFirstName  *string `json:"driver_first_name" example:"Ivan"`
	DriverMiddleName *string `json:"driver_middle_name" example:"Ivanovich"`
	DriverID         *int64  `json:"driver_id" example:"55"`

	StartTime *string `json:"start_time" example:"2026-03-18T08:00:00Z"`
	EndTime   *string `json:"end_time" example:"2026-03-18T18:00:00Z"`

	MileageStart               *int `json:"mileage_start" example:"125000"`
	MileageEnd                 *int `json:"mileage_end" example:"125180"`
	FuelStart                  *int `json:"fuel_start" example:"30"`
	FuelIssued                 *int `json:"fuel_issued" example:"20"`
	FuelConsumptionTheoretical *int `json:"fuel_consumption_theoretical" example:"18"`
	FuelConsumptionActual      *int `json:"fuel_consumption_actual" example:"20"`

	StatusID *int64 `json:"status_id" example:"1"`
}

// UpdateTripsheetRequest represents request body for updating a tripsheet.
type UpdateTripsheetRequest struct {
	TripsheetNumber    string  `json:"tripsheet_number" example:"PL-0002"`
	TripsheetDate      string  `json:"tripsheet_date" example:"2026-03-18"`
	VehicleID          *int64  `json:"vehicle_id" example:"12"`
	VehicleBrand       *string `json:"vehicle_brand" example:"Toyota Camry"`
	VehiclePlateNumber string  `json:"vehicle_plate_number" example:"123ABC01"`

	DriverLastName   *string `json:"driver_last_name" example:"Ivanov"`
	DriverFirstName  *string `json:"driver_first_name" example:"Ivan"`
	DriverMiddleName *string `json:"driver_middle_name" example:"Ivanovich"`
	DriverID         *int64  `json:"driver_id" example:"55"`

	StartTime *string `json:"start_time" example:"2026-03-18T08:00:00Z"`
	EndTime   *string `json:"end_time" example:"2026-03-18T18:00:00Z"`

	MileageStart               *int `json:"mileage_start" example:"125000"`
	MileageEnd                 *int `json:"mileage_end" example:"125180"`
	FuelStart                  *int `json:"fuel_start" example:"30"`
	FuelIssued                 *int `json:"fuel_issued" example:"20"`
	FuelConsumptionTheoretical *int `json:"fuel_consumption_theoretical" example:"18"`
	FuelConsumptionActual      *int `json:"fuel_consumption_actual" example:"20"`

	StatusID *int64 `json:"status_id" example:"1"`
}

// CreateTripsheetResponse represents created or updated tripsheet response.
type CreateTripsheetResponse struct {
	ID                         int64   `json:"id" example:"1"`
	TripsheetNumber            string  `json:"tripsheet_number" example:"PL-0002"`
	TripsheetDate              string  `json:"tripsheet_date" example:"2026-03-18"`
	VehicleID                  *int64  `json:"vehicle_id,omitempty" example:"12"`
	VehicleBrand               *string `json:"vehicle_brand,omitempty" example:"Toyota Camry"`
	VehiclePlateNumber         string  `json:"vehicle_plate_number" example:"123ABC01"`
	DriverLastName             *string `json:"driver_last_name,omitempty" example:"Ivanov"`
	DriverFirstName            *string `json:"driver_first_name,omitempty" example:"Ivan"`
	DriverMiddleName           *string `json:"driver_middle_name,omitempty" example:"Ivanovich"`
	DriverID                   *int64  `json:"driver_id,omitempty" example:"55"`
	StartTime                  *string `json:"start_time,omitempty" example:"2026-03-18T08:00:00Z"`
	EndTime                    *string `json:"end_time,omitempty" example:"2026-03-18T18:00:00Z"`
	MileageStart               int     `json:"mileage_start" example:"125000"`
	MileageEnd                 int     `json:"mileage_end" example:"125180"`
	FuelStart                  int     `json:"fuel_start" example:"30"`
	FuelIssued                 int     `json:"fuel_issued" example:"20"`
	FuelConsumptionTheoretical int     `json:"fuel_consumption_theoretical" example:"18"`
	FuelConsumptionActual      int     `json:"fuel_consumption_actual" example:"20"`
	StatusID                   int64   `json:"status_id" example:"1"`
	CreatedAt                  string  `json:"created_at" example:"2026-03-18T08:00:00Z"`
	UpdatedAt                  string  `json:"updated_at" example:"2026-03-18T08:00:00Z"`
}

// TripsheetFilter represents filters for listing tripsheets.
type TripsheetFilter struct {
	VehicleID          *int64  `form:"vehicle_id" example:"12"`
	DriverID           *int64  `form:"driver_id" example:"55"`
	DateFrom           *string `form:"date_from" example:"2026-03-01"`
	DateTo             *string `form:"date_to" example:"2026-03-31"`
	StatusID           *int64  `form:"status_id" example:"1"`
	TripsheetNumber    *string `form:"tripsheet_number" example:"PL-0002"`
	VehiclePlateNumber *string `form:"vehicle_plate_number" example:"123ABC01"`
	VehicleBrand       *string `form:"vehicle_brand" example:"Toyota"`
	DriverLastName     *string `form:"driver_last_name" example:"Ivanov"`
	DriverFirstName    *string `form:"driver_first_name" example:"Ivan"`
	DriverMiddleName   *string `form:"driver_middle_name" example:"Ivanovich"`
	Limit              int     `form:"limit" example:"50"`
	Offset             int     `form:"offset" example:"0"`
	SortBy             string  `form:"sort_by" example:"tripsheet_date"`
	Order              string  `form:"order" example:"desc"`
}

// TripsheetResponse represents tripsheet response.
type TripsheetResponse struct {
	ID                         int64   `json:"id" example:"1"`
	TripsheetNumber            string  `json:"tripsheet_number" example:"PL-0002"`
	TripsheetDate              string  `json:"tripsheet_date" example:"2026-03-18"`
	VehicleID                  *int64  `json:"vehicle_id,omitempty" example:"12"`
	VehicleBrand               *string `json:"vehicle_brand,omitempty" example:"Toyota Camry"`
	VehiclePlateNumber         string  `json:"vehicle_plate_number" example:"123ABC01"`
	DriverLastName             *string `json:"driver_last_name,omitempty" example:"Ivanov"`
	DriverFirstName            *string `json:"driver_first_name,omitempty" example:"Ivan"`
	DriverMiddleName           *string `json:"driver_middle_name,omitempty" example:"Ivanovich"`
	DriverID                   *int64  `json:"driver_id,omitempty" example:"55"`
	StartTime                  *string `json:"start_time,omitempty" example:"2026-03-18T08:00:00Z"`
	EndTime                    *string `json:"end_time,omitempty" example:"2026-03-18T18:00:00Z"`
	MileageStart               int     `json:"mileage_start" example:"125000"`
	MileageEnd                 int     `json:"mileage_end" example:"125180"`
	FuelStart                  int     `json:"fuel_start" example:"30"`
	FuelIssued                 int     `json:"fuel_issued" example:"20"`
	FuelConsumptionTheoretical int     `json:"fuel_consumption_theoretical" example:"18"`
	FuelConsumptionActual      int     `json:"fuel_consumption_actual" example:"20"`
	StatusID                   int64   `json:"status_id" example:"1"`
	CreatedAt                  string  `json:"created_at" example:"2026-03-18T08:00:00Z"`
	UpdatedAt                  string  `json:"updated_at" example:"2026-03-18T08:00:00Z"`
}
