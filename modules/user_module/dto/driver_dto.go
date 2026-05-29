package dto

type CreateDriverRequest struct {
	IIN             string `json:"iin" binding:"required" example:"990101350011"`
	Name            string `json:"name" binding:"required" example:"Ivan"`
	Surname         string `json:"surname" binding:"required" example:"Ivanov"`
	Middlename      string `json:"middlename" example:"Ivanovich"`
	Phone           string `json:"phone" example:"+77001234567"`
	Mail            string `json:"mail" example:"driver@mail.com"`
	BirthDate       string `json:"birth_date,omitempty" example:"1990-11-11"`
	LicenseNumber   string `json:"license_number,omitempty" example:"DL-123456"`
	LicenseCategory string `json:"license_category,omitempty" example:"B, C"`
	ExperienceYears *int   `json:"experience_years,omitempty" example:"5"`
	StatusID        int64  `json:"status_id,omitempty" example:"1"`
	Status          string `json:"status,omitempty" example:"available"`
}

type UpdateDriverRequest struct {
	IIN             *string `json:"iin"`
	Name            *string `json:"name"`
	Surname         *string `json:"surname"`
	Middlename      *string `json:"middlename"`
	Phone           *string `json:"phone"`
	Mail            *string `json:"mail"`
	BirthDate       *string `json:"birth_date"`
	LicenseNumber   *string `json:"license_number"`
	LicenseCategory *string `json:"license_category"`
	ExperienceYears *int    `json:"experience_years"`
	StatusID        *int64  `json:"status_id" example:"1"`
	Status          *string `json:"status" example:"available"`
}

type UploadDriverPhotoRequest struct {
	Photo string `form:"photo" swaggerignore:"true"`
}

type UpdateDriverStatusRequest struct {
	StatusID int64  `json:"status_id,omitempty" example:"2"`
	Status   string `json:"status,omitempty" example:"on_trip"`
}

type DriverListQuery struct {
	Search          string
	Status          string
	BoardNumber     string
	SortBy          string
	Order           string
	IncludeArchived bool
	Limit           int
	Offset          int
}

type DriverPassportQuery struct {
	TripsLimit      int
	TripsOffset     int
	IncidentsLimit  int
	IncidentsOffset int
	IncludeArchived bool
}

type AssignDriverVehicleRequest struct {
	VehicleID int64 `json:"vehicle_id" binding:"required" example:"1"`
}
