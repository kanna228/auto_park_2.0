package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/models"
	"auto_park/modules/vehicle_module/repository"
)

type TirePlaceService interface {
	Create(ctx context.Context, req dto.TirePlaceCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.TirePlaceResponse, error)
	List(ctx context.Context, q dto.TirePlaceListQuery) (*dto.TirePlaceListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.TirePlaceUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type tirePlaceService struct {
	repo repository.TirePlaceRepository
}

func NewTirePlaceService(repo repository.TirePlaceRepository) TirePlaceService {
	return &tirePlaceService{repo: repo}
}

func (s *tirePlaceService) Create(ctx context.Context, req dto.TirePlaceCreateRequest) (int64, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return 0, fmt.Errorf("name is required")
	}

	return s.repo.Create(ctx, req.Name)
}

func (s *tirePlaceService) GetByID(ctx context.Context, id int64) (*dto.TirePlaceResponse, error) {
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

	resp := mapTirePlaceToDTO(*item)
	return &resp, nil
}

func (s *tirePlaceService) List(ctx context.Context, q dto.TirePlaceListQuery) (*dto.TirePlaceListResponse, error) {
	items, total, err := s.repo.List(ctx, repository.ListTirePlacesParams{
		Name:   q.Name,
		Limit:  q.Limit,
		Offset: q.Offset,
		SortBy: q.SortBy,
		Order:  q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.TirePlaceResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapTirePlaceToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.TirePlaceListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *tirePlaceService) UpdateByID(ctx context.Context, id int64, req dto.TirePlaceUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return false, fmt.Errorf("name is required")
	}

	return s.repo.UpdateByID(ctx, id, req.Name)
}

func (s *tirePlaceService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	return s.repo.DeleteByID(ctx, id)
}

func mapTirePlaceToDTO(item models.TirePlace) dto.TirePlaceResponse {
	return dto.TirePlaceResponse{
		ID:   item.ID,
		Name: item.Name,
	}
}
