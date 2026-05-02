package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/models"
	"auto_park/modules/warehouse_module/repository"
)

const defaultPartRequestStatusID int64 = 1
const editablePartRequestStatusCode = "new"

type PartRequestService interface {
	Create(ctx context.Context, authorUserID int64, req dto.PartRequestCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.PartRequestResponse, error)
	List(ctx context.Context, q dto.PartRequestListQuery) (*dto.PartRequestListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.PartRequestUpdateRequest) (bool, error)
	UpdateStatusByID(ctx context.Context, id int64, req dto.PartRequestStatusUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	ListStatuses(ctx context.Context) ([]dto.PartRequestStatusResponse, error)
}

type partRequestService struct {
	repo repository.PartRequestRepository
}

func NewPartRequestService(repo repository.PartRequestRepository) PartRequestService {
	return &partRequestService{repo: repo}
}

func (s *partRequestService) Create(ctx context.Context, authorUserID int64, req dto.PartRequestCreateRequest) (int64, error) {
	if authorUserID <= 0 {
		return 0, fmt.Errorf("invalid author user id")
	}
	if err := s.validatePartID(ctx, req.PartID); err != nil {
		return 0, err
	}
	if req.Quantity <= 0 {
		return 0, fmt.Errorf("quantity must be greater than 0")
	}
	comment, err := normalizeMechanicComment(req.MechanicComment)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, repository.CreatePartRequestParams{
		PartID:          req.PartID,
		Quantity:        req.Quantity,
		MechanicComment: comment,
		StatusID:        defaultPartRequestStatusID,
		AuthorUserID:    authorUserID,
	})
}

func (s *partRequestService) GetByID(ctx context.Context, id int64) (*dto.PartRequestResponse, error) {
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
	resp := mapPartRequestToDTO(*item)
	return &resp, nil
}

func (s *partRequestService) List(ctx context.Context, q dto.PartRequestListQuery) (*dto.PartRequestListResponse, error) {
	items, total, err := s.repo.List(ctx, repository.ListPartRequestsParams{
		PartID:       q.PartID,
		StatusID:     q.StatusID,
		StatusCode:   strings.TrimSpace(q.StatusCode),
		AuthorUserID: q.AuthorUserID,
		Limit:        q.Limit,
		Offset:       q.Offset,
		SortBy:       q.SortBy,
		Order:        q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.PartRequestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartRequestToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.PartRequestListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *partRequestService) UpdateByID(ctx context.Context, id int64, req dto.PartRequestUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if current == nil {
		return false, nil
	}
	if current.StatusCode != editablePartRequestStatusCode {
		return false, repository.ErrPartRequestLocked
	}

	if err := s.validatePartID(ctx, req.PartID); err != nil {
		return false, err
	}
	if req.Quantity <= 0 {
		return false, fmt.Errorf("quantity must be greater than 0")
	}
	comment, err := normalizeMechanicComment(req.MechanicComment)
	if err != nil {
		return false, err
	}
	if err := s.validateStatusID(ctx, req.StatusID); err != nil {
		return false, err
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdatePartRequestParams{
		PartID:          req.PartID,
		Quantity:        req.Quantity,
		MechanicComment: comment,
		StatusID:        req.StatusID,
	})
}

func (s *partRequestService) UpdateStatusByID(ctx context.Context, id int64, req dto.PartRequestStatusUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if err := s.validateStatusID(ctx, req.StatusID); err != nil {
		return false, err
	}
	return s.repo.UpdateStatusByID(ctx, id, req.StatusID)
}

func (s *partRequestService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if current == nil {
		return false, nil
	}
	if current.StatusCode != editablePartRequestStatusCode {
		return false, repository.ErrPartRequestLocked
	}

	return s.repo.DeleteByID(ctx, id)
}

func (s *partRequestService) ListStatuses(ctx context.Context) ([]dto.PartRequestStatusResponse, error) {
	items, err := s.repo.ListStatuses(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PartRequestStatusResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartRequestStatusToDTO(item))
	}
	return out, nil
}

func (s *partRequestService) validatePartID(ctx context.Context, partID int64) error {
	if partID <= 0 {
		return fmt.Errorf("part_id is required")
	}
	exists, err := s.repo.PartExists(ctx, partID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrPartRequestPartNotFound
	}
	return nil
}

func (s *partRequestService) validateStatusID(ctx context.Context, statusID int64) error {
	if statusID <= 0 {
		return fmt.Errorf("status_id is required")
	}
	exists, err := s.repo.StatusExists(ctx, statusID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrPartRequestStatusNotFound
	}
	return nil
}

func normalizeMechanicComment(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("mechanic_comment is required")
	}
	if len([]rune(trimmed)) > 2000 {
		return "", fmt.Errorf("mechanic_comment must be less than or equal to 2000 characters")
	}
	return trimmed, nil
}

func mapPartRequestToDTO(item models.PartRequest) dto.PartRequestResponse {
	return dto.PartRequestResponse{
		ID:     item.ID,
		PartID: item.PartID,
		Part: dto.PartRequestPartBriefResponse{
			ID:            item.PartID,
			CatalogPartID: item.PartCatalogCode,
			Name:          item.PartName,
			Category:      item.PartCategory,
		},
		Quantity:        item.Quantity,
		MechanicComment: item.MechanicComment,
		StatusID:        item.StatusID,
		Status: dto.PartRequestStatusResponse{
			ID:   item.StatusID,
			Code: item.StatusCode,
			Name: item.StatusName,
		},
		AuthorUserID:   item.AuthorUserID,
		AuthorEmail:    item.AuthorEmail,
		AuthorFullName: item.AuthorFullName,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func mapPartRequestStatusToDTO(item models.PartRequestStatus) dto.PartRequestStatusResponse {
	return dto.PartRequestStatusResponse{
		ID:   item.ID,
		Code: item.Code,
		Name: item.Name,
	}
}
