package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/models"
	"auto_park/modules/warehouse_module/repository"
)

type PurchaseRequestService interface {
	Create(ctx context.Context, userID int64, req dto.PurchaseRequestCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.PurchaseRequestResponse, error)
	List(ctx context.Context, q dto.PurchaseRequestListQuery) (*dto.PurchaseRequestListResponse, error)
	Confirm(ctx context.Context, id int64, userID int64) (bool, error)
	Cancel(ctx context.Context, id int64, userID int64, comment *string) (bool, error)
}

type purchaseRequestService struct {
	repo repository.PurchaseRequestRepository
}

func NewPurchaseRequestService(repo repository.PurchaseRequestRepository) PurchaseRequestService {
	return &purchaseRequestService{repo: repo}
}

func (s *purchaseRequestService) Create(ctx context.Context, userID int64, req dto.PurchaseRequestCreateRequest) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user id")
	}
	if req.PartID <= 0 {
		return 0, fmt.Errorf("part_id is required")
	}
	if req.Quantity <= 0 {
		return 0, fmt.Errorf("quantity must be greater than 0")
	}
	if req.SourcePartRequestID != nil && *req.SourcePartRequestID <= 0 {
		return 0, fmt.Errorf("source_part_request_id must be greater than 0")
	}
	comment, err := normalizeOptionalField("comment", req.Comment)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, repository.CreatePurchaseRequestParams{
		PartID:              req.PartID,
		Quantity:            req.Quantity,
		SourcePartRequestID: req.SourcePartRequestID,
		Comment:             comment,
		CreatedByUserID:     userID,
	})
}

func (s *purchaseRequestService) GetByID(ctx context.Context, id int64) (*dto.PurchaseRequestResponse, error) {
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
	resp := mapPurchaseRequestToDTO(*item)
	return &resp, nil
}

func (s *purchaseRequestService) List(ctx context.Context, q dto.PurchaseRequestListQuery) (*dto.PurchaseRequestListResponse, error) {
	if q.PartID < 0 {
		return nil, fmt.Errorf("part_id cannot be negative")
	}
	if q.SourcePartRequestID < 0 {
		return nil, fmt.Errorf("source_part_request_id cannot be negative")
	}
	status := strings.ToLower(strings.TrimSpace(q.Status))
	if status != "" && status != "new" && status != "confirmed" && status != "cancelled" {
		return nil, fmt.Errorf("invalid status")
	}

	items, total, err := s.repo.List(ctx, repository.ListPurchaseRequestsParams{
		PartID:              q.PartID,
		SourcePartRequestID: q.SourcePartRequestID,
		Status:              status,
		Limit:               q.Limit,
		Offset:              q.Offset,
		SortBy:              q.SortBy,
		Order:               q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.PurchaseRequestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPurchaseRequestToDTO(item))
	}

	return &dto.PurchaseRequestListResponse{
		Items:  out,
		Total:  total,
		Limit:  normalizeLimit(q.Limit),
		Offset: normalizeOffset(q.Offset),
	}, nil
}

func (s *purchaseRequestService) Confirm(ctx context.Context, id int64, userID int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if userID <= 0 {
		return false, fmt.Errorf("invalid user id")
	}
	return s.repo.Confirm(ctx, id, userID)
}

func (s *purchaseRequestService) Cancel(ctx context.Context, id int64, userID int64, comment *string) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if userID <= 0 {
		return false, fmt.Errorf("invalid user id")
	}
	normalizedComment, err := normalizeOptionalField("comment", comment)
	if err != nil {
		return false, err
	}
	return s.repo.Cancel(ctx, id, userID, normalizedComment)
}

func mapPurchaseRequestToDTO(item models.PurchaseRequest) dto.PurchaseRequestResponse {
	return dto.PurchaseRequestResponse{
		ID:     item.ID,
		PartID: item.PartID,
		Part: dto.PurchaseRequestPartBriefResponse{
			ID:            item.PartID,
			CatalogPartID: item.PartCatalogCode,
			Name:          item.PartName,
			Category:      item.PartCategory,
		},
		Quantity:            item.Quantity,
		Status:              item.Status,
		SourcePartRequestID: item.SourcePartRequestID,
		Comment:             item.Comment,
		CreatedByUserID:     item.CreatedByUserID,
		CreatedByEmail:      item.CreatedByEmail,
		CreatedByFullName:   item.CreatedByFullName,
		ConfirmedByUserID:   item.ConfirmedByUserID,
		ConfirmedByEmail:    item.ConfirmedByEmail,
		ConfirmedByFullName: item.ConfirmedByFullName,
		ConfirmedAt:         item.ConfirmedAt,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}
