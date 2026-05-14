package service

import (
	"context"
	"fmt"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/models"
	"auto_park/modules/vehicle_module/repository"
)

func (s *tireService) GetByID(ctx context.Context, id int64) (*dto.TireResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	resp := mapTireToDTO(*item)
	return &resp, nil
}

func (s *tireService) List(ctx context.Context, q dto.TireListQuery) (*dto.TireListResponse, error) {
	items, total, err := s.repo.List(ctx, repository.ListTiresParams{
		VehicleID: q.VehicleID,
		PlaceID:   q.PlaceID,
		Tire:      q.Tire,
		Limit:     q.Limit,
		Offset:    q.Offset,
		SortBy:    q.SortBy,
		Order:     q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.TireResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapTireToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.TireListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func mapTireToDTO(item models.Tire) dto.TireResponse {
	remaining := item.MaxUsage - item.Mileage
	if remaining < 0 {
		remaining = 0
	}

	return dto.TireResponse{
		ID:             item.ID,
		PlaceID:        item.PlaceID,
		PlaceName:      item.PlaceName,
		VehicleID:      item.VehicleID,
		Tire:           item.Tire,
		Mileage:        item.Mileage,
		MaxUsage:       item.MaxUsage,
		RemainingUsage: remaining,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func (s *tireService) GetByVehicleID(ctx context.Context, vehicleID int64, limit int, offset int) (*dto.TireListResponse, error) {
	if vehicleID <= 0 {
		return nil, fmt.Errorf("invalid vehicle_id")
	}

	items, total, err := s.repo.GetByVehicleID(ctx, vehicleID, limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]dto.TireResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapTireToDTO(item))
	}

	normalizedLimit := limit
	if normalizedLimit <= 0 || normalizedLimit > 200 {
		normalizedLimit = 50
	}

	normalizedOffset := offset
	if normalizedOffset < 0 {
		normalizedOffset = 0
	}

	return &dto.TireListResponse{
		Items:  out,
		Total:  total,
		Limit:  normalizedLimit,
		Offset: normalizedOffset,
	}, nil
}
