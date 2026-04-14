package service

import (
	"auto_park/vehicle_module/dto"
	"context"
)

func (s *vehicleService) ListVehicleStatuses(ctx context.Context) (*dto.VehicleStatusListResponse, error) {
	items, err := s.repo.ListVehicleStatuses(ctx)
	if err != nil {
		return nil, err
	}

	resp := &dto.VehicleStatusListResponse{
		Items: make([]dto.VehicleStatusResponse, 0, len(items)),
		Total: int64(len(items)),
	}

	for _, item := range items {
		resp.Items = append(resp.Items, dto.VehicleStatusResponse{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	return resp, nil
}
