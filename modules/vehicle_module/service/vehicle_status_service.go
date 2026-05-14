package service

import (
	"context"

	"auto_park/modules/vehicle_module/dto"
)

func (s *vehicleService) ListVehicleStatuses(ctx context.Context, limit int, offset int) (*dto.VehicleStatusListResponse, error) {
	items, err := s.repo.ListVehicleStatuses(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	total := int64(len(items))
	start := offset
	if start > len(items) {
		start = len(items)
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	resp := &dto.VehicleStatusListResponse{
		Items:  make([]dto.VehicleStatusResponse, 0, end-start),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	for _, item := range items[start:end] {
		resp.Items = append(resp.Items, dto.VehicleStatusResponse{
			ID:   item.ID,
			Name: item.Name,
		})
	}

	return resp, nil
}
