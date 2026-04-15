package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/repository"
)

func (s *tireService) UpdateByID(ctx context.Context, id int64, req dto.TireUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	req.Tire = strings.TrimSpace(req.Tire)

	if req.PlaceID <= 0 {
		return false, fmt.Errorf("place_id is required")
	}
	if req.Tire == "" {
		return false, fmt.Errorf("tire is required")
	}
	if req.Mileage < 0 {
		return false, fmt.Errorf("mileage must be >= 0")
	}
	if req.MaxUsage < 0 {
		return false, fmt.Errorf("max_usage must be >= 0")
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdateTireParams{
		PlaceID:   req.PlaceID,
		VehicleID: req.VehicleID,
		Tire:      req.Tire,
		Mileage:   req.Mileage,
		MaxUsage:  req.MaxUsage,
	})
}
