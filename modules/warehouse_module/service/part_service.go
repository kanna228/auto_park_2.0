package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/models"
	"auto_park/modules/warehouse_module/repository"
)

var allowedTextRe = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё0-9\s\-_/().,xX№]+$`)

type PartService interface {
	Create(ctx context.Context, req dto.PartCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.PartResponse, error)
	List(ctx context.Context, q dto.PartListQuery) (*dto.PartListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.PartUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type partService struct {
	repo repository.PartRepository
}

func NewPartService(repo repository.PartRepository) PartService {
	return &partService{repo: repo}
}

func (s *partService) Create(ctx context.Context, req dto.PartCreateRequest) (int64, error) {
	partID, err := normalizeRequiredField("part_id", req.PartID)
	if err != nil {
		return 0, err
	}
	name, err := normalizeRequiredField("name", req.Name)
	if err != nil {
		return 0, err
	}
	category, err := normalizeRequiredField("category", req.Category)
	if err != nil {
		return 0, err
	}
	if req.StartQuantity < 0 {
		return 0, fmt.Errorf("start_quantity must be greater than or equal to 0")
	}
	dimensions, err := normalizeOptionalField("dimensions", req.Dimensions)
	if err != nil {
		return 0, err
	}
	manufacturer, err := normalizeOptionalField("manufacturer", req.Manufacturer)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, repository.CreatePartParams{
		PartID:       partID,
		Name:         name,
		Quantity:     req.StartQuantity,
		Category:     category,
		Dimensions:   dimensions,
		Manufacturer: manufacturer,
		IsConsumable: req.IsConsumable,
	})
}

func (s *partService) GetByID(ctx context.Context, id int64) (*dto.PartResponse, error) {
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
	resp := mapPartToDTO(*item)
	return &resp, nil
}

func (s *partService) List(ctx context.Context, q dto.PartListQuery) (*dto.PartListResponse, error) {
	items, total, err := s.repo.List(ctx, repository.ListPartsParams{
		PartID:   strings.TrimSpace(q.PartID),
		Name:     strings.TrimSpace(q.Name),
		Category: strings.TrimSpace(q.Category),
		Limit:    q.Limit,
		Offset:   q.Offset,
		SortBy:   q.SortBy,
		Order:    q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.PartResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.PartListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *partService) UpdateByID(ctx context.Context, id int64, req dto.PartUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	name, err := normalizeRequiredField("name", req.Name)
	if err != nil {
		return false, err
	}
	category, err := normalizeRequiredField("category", req.Category)
	if err != nil {
		return false, err
	}
	if req.Quantity < 0 {
		return false, fmt.Errorf("quantity must be greater than or equal to 0")
	}
	dimensions, err := normalizeOptionalField("dimensions", req.Dimensions)
	if err != nil {
		return false, err
	}
	manufacturer, err := normalizeOptionalField("manufacturer", req.Manufacturer)
	if err != nil {
		return false, err
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdatePartParams{
		Name:         name,
		Quantity:     req.Quantity,
		Category:     category,
		Dimensions:   dimensions,
		Manufacturer: manufacturer,
		IsConsumable: req.IsConsumable,
	})
}

func (s *partService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeleteByID(ctx, id)
}

func normalizeRequiredField(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if !allowedTextRe.MatchString(trimmed) {
		return "", fmt.Errorf("%s contains invalid characters", field)
	}
	return trimmed, nil
}

func normalizeOptionalField(field string, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if !allowedTextRe.MatchString(trimmed) {
		return nil, fmt.Errorf("%s contains invalid characters", field)
	}
	return &trimmed, nil
}

func mapPartToDTO(item models.Part) dto.PartResponse {
	return dto.PartResponse{
		ID:           item.ID,
		PartID:       item.PartID,
		Name:         item.Name,
		Quantity:     item.Quantity,
		Category:     item.Category,
		Dimensions:   item.Dimensions,
		Manufacturer: item.Manufacturer,
		IsConsumable: item.IsConsumable,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}
