package service

import (
	"context"
	"fmt"

	"auto_park/modules/fuel_norm_module/dto"
	"auto_park/modules/fuel_norm_module/repository"
)

type FuelNormService interface {
	Create(ctx context.Context, req dto.FuelNormRequest) (*dto.FuelNormResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.FuelNormResponse, error)
	List(ctx context.Context, limit int, offset int) (*dto.FuelNormListResponse, error)
	Update(ctx context.Context, id int64, req dto.FuelNormRequest) (*dto.FuelNormResponse, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type fuelNormService struct {
	repo repository.FuelNormRepository
}

func NewFuelNormService(repo repository.FuelNormRepository) FuelNormService {
	return &fuelNormService{repo: repo}
}

func (s *fuelNormService) Create(ctx context.Context, req dto.FuelNormRequest) (*dto.FuelNormResponse, error) {
	if err := validateFuelNorm(req); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, req)
}

func (s *fuelNormService) GetByID(ctx context.Context, id int64) (*dto.FuelNormResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *fuelNormService) List(ctx context.Context, limit int, offset int) (*dto.FuelNormListResponse, error) {
	items, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return &dto.FuelNormListResponse{
		Items:  items,
		Total:  total,
		Limit:  normalizeLimit(limit),
		Offset: normalizeOffset(offset),
	}, nil
}

func (s *fuelNormService) Update(ctx context.Context, id int64, req dto.FuelNormRequest) (*dto.FuelNormResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if err := validateFuelNorm(req); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, req)
}

func (s *fuelNormService) Delete(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}

func validateFuelNorm(req dto.FuelNormRequest) error {
	if req.VehicleID <= 0 {
		return fmt.Errorf("vehicle_id is required")
	}
	if req.NormPer100KM <= 0 {
		return fmt.Errorf("norm_per_100km must be greater than 0")
	}
	if req.SummerNorm <= 0 {
		return fmt.Errorf("summer_norm must be greater than 0")
	}
	if req.WinterNorm <= 0 {
		return fmt.Errorf("winter_norm must be greater than 0")
	}
	if req.ColdAirNorm < 0 {
		return fmt.Errorf("cold_air_norm must be greater than or equal to 0")
	}
	if req.WarmAirNorm < 0 {
		return fmt.Errorf("warm_air_norm must be greater than or equal to 0")
	}
	return nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
