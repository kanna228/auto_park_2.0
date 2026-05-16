package dto

import "time"

type FuelNormRequest struct {
	VehicleID    int64   `json:"vehicle_id" example:"12"`
	NormPer100KM float64 `json:"norm_per_100km" example:"18.36"`
	SummerNorm   float64 `json:"summer_norm" example:"16.56"`
	WinterNorm   float64 `json:"winter_norm" example:"18.36"`
	ColdAirNorm  float64 `json:"cold_air_norm" example:"1.20"`
	WarmAirNorm  float64 `json:"warm_air_norm" example:"1.20"`
	Deviation    float64 `json:"deviation" example:"5.0"`
}

type FuelNormResponse struct {
	ID           int64     `json:"id" example:"1"`
	VehicleID    int64     `json:"vehicle_id" example:"12"`
	NormPer100KM float64   `json:"norm_per_100km" example:"18.36"`
	SummerNorm   float64   `json:"summer_norm" example:"16.56"`
	WinterNorm   float64   `json:"winter_norm" example:"18.36"`
	ColdAirNorm  float64   `json:"cold_air_norm" example:"1.20"`
	WarmAirNorm  float64   `json:"warm_air_norm" example:"1.20"`
	Deviation    float64   `json:"deviation" example:"5.0"`
	CreatedAt    time.Time `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

type FuelNormListResponse struct {
	Items  []FuelNormResponse `json:"items"`
	Total  int64              `json:"total" example:"1"`
	Limit  int                `json:"limit" example:"50"`
	Offset int                `json:"offset" example:"0"`
}
