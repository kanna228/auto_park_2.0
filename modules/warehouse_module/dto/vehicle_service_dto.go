package dto

import "time"

type PartsCollectionCreateRequest struct {
	Name        string  `json:"name" binding:"required" example:"Передний бампер"`
	Description *string `json:"description,omitempty" example:"Передняя наружная часть кузова автомобиля"`
}

type PartsCollectionUpdateRequest struct {
	Name        string  `json:"name" binding:"required" example:"Передний бампер"`
	Description *string `json:"description,omitempty" example:"Передняя наружная часть кузова автомобиля"`
}

type PartsCollectionResponse struct {
	ID          int64     `json:"id" example:"1"`
	Name        string    `json:"name" example:"Передний бампер"`
	Description *string   `json:"description,omitempty" example:"Передняя наружная часть кузова автомобиля"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PartsCollectionListQuery struct {
	Name   string
	Limit  int
	Offset int
	SortBy string
	Order  string
}

type PartsCollectionListResponse struct {
	Items  []PartsCollectionResponse `json:"items"`
	Total  int64                     `json:"total" example:"20"`
	Limit  int                       `json:"limit" example:"50"`
	Offset int                       `json:"offset" example:"0"`
}

type ServiceTypeCreateRequest struct {
	Name        string  `json:"name" binding:"required" example:"Полировка"`
	Description *string `json:"description,omitempty" example:"Восстановление блеска кузова"`
}

type ServiceTypeUpdateRequest struct {
	Name        string  `json:"name" binding:"required" example:"Полировка"`
	Description *string `json:"description,omitempty" example:"Восстановление блеска кузова"`
}

type ServiceTypeResponse struct {
	ID          int64     `json:"id" example:"1"`
	Name        string    `json:"name" example:"Полировка"`
	Description *string   `json:"description,omitempty" example:"Восстановление блеска кузова"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServiceTypeListQuery struct {
	Name   string
	Limit  int
	Offset int
	SortBy string
	Order  string
}

type ServiceTypeListResponse struct {
	Items  []ServiceTypeResponse `json:"items"`
	Total  int64                 `json:"total" example:"10"`
	Limit  int                   `json:"limit" example:"50"`
	Offset int                   `json:"offset" example:"0"`
}

type VehicleServiceCreateRequest struct {
	TypeID    int64  `json:"type_id" binding:"required" example:"1"`
	PartID    int64  `json:"part_id" binding:"required" example:"1"`
	VehicleID int64  `json:"vehicle_id" binding:"required" example:"12"`
	Date      string `json:"date" binding:"required" example:"2026-05-14"`
}

type VehicleServiceUpdateRequest struct {
	TypeID    int64  `json:"type_id" binding:"required" example:"1"`
	PartID    int64  `json:"part_id" binding:"required" example:"1"`
	VehicleID int64  `json:"vehicle_id" binding:"required" example:"12"`
	Date      string `json:"date" binding:"required" example:"2026-05-14"`
}

type VehicleServicePartBriefResponse struct {
	ID          int64   `json:"id" example:"1"`
	Name        string  `json:"name" example:"Передний бампер"`
	Description *string `json:"description,omitempty" example:"Передняя наружная часть кузова автомобиля"`
}

type VehicleServiceTypeBriefResponse struct {
	ID          int64   `json:"id" example:"1"`
	Name        string  `json:"name" example:"Полировка"`
	Description *string `json:"description,omitempty" example:"Восстановление блеска кузова"`
}

type VehicleServiceVehicleBriefResponse struct {
	ID          int64  `json:"id" example:"12"`
	StateNumber string `json:"state_number" example:"777ABC01"`
	BrandModel  string `json:"brand_model" example:"Toyota Camry"`
}

type VehicleServiceResponse struct {
	ID        int64                              `json:"id" example:"1"`
	TypeID    int64                              `json:"type_id" example:"1"`
	Type      VehicleServiceTypeBriefResponse    `json:"type"`
	PartID    int64                              `json:"part_id" example:"1"`
	Part      VehicleServicePartBriefResponse    `json:"part"`
	VehicleID int64                              `json:"vehicle_id" example:"12"`
	Vehicle   VehicleServiceVehicleBriefResponse `json:"vehicle"`
	Date      string                             `json:"date" example:"2026-05-14"`
	CreatedAt time.Time                          `json:"created_at"`
	UpdatedAt time.Time                          `json:"updated_at"`
}

type VehicleServiceListQuery struct {
	TypeID    int64
	PartID    int64
	VehicleID int64
	TypeName  string
	PartName  string
	DateFrom  string
	DateTo    string
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

type VehicleServiceListResponse struct {
	Items  []VehicleServiceResponse `json:"items"`
	Total  int64                    `json:"total" example:"1"`
	Limit  int                      `json:"limit" example:"50"`
	Offset int                      `json:"offset" example:"0"`
}
