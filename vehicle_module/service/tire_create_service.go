package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/vehicle_module/dto"
	"auto_park/vehicle_module/repository"
)

type TireService interface {
	Create(ctx context.Context, req dto.TireCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.TireResponse, error)
	List(ctx context.Context, q dto.TireListQuery) (*dto.TireListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.TireUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	GetByVehicleID(ctx context.Context, vehicleID int64) (*dto.TireListResponse, error)
}

type tireService struct {
	repo repository.TireRepository
}

func NewTireService(repo repository.TireRepository) TireService {
	return &tireService{repo: repo}
}

func (s *tireService) Create(ctx context.Context, req dto.TireCreateRequest) (int64, error) {
	req.Tire = strings.TrimSpace(req.Tire)

	if req.PlaceID <= 0 {
		return 0, fmt.Errorf("place_id is required")
	}
	if req.Tire == "" {
		return 0, fmt.Errorf("tire is required")
	}
	if req.Mileage < 0 {
		return 0, fmt.Errorf("mileage must be >= 0")
	}
	if req.MaxUsage < 0 {
		return 0, fmt.Errorf("max_usage must be >= 0")
	}

	return s.repo.Create(ctx, repository.CreateTireParams{
		PlaceID:   req.PlaceID,
		VehicleID: req.VehicleID,
		Tire:      req.Tire,
		Mileage:   req.Mileage,
		MaxUsage:  req.MaxUsage,
	})
}
