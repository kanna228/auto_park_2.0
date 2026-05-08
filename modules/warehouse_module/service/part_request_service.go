package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/models"
	"auto_park/modules/warehouse_module/repository"
)

const defaultPartRequestStatusID int64 = 1
const editablePartRequestStatusCode = "new"
const rejectedPartRequestStatusCode = "rejected"
const warehouseManagerRoleID int64 = 5
const AdminRoleID int64 = 1
const dutyMechanicRoleID int64 = 4

var ErrPartRequestRejectionCommentRequired = errors.New("rejection comment is required when request status is rejected")
var ErrPartRequestRejectForbidden = errors.New("only warehouse manager can reject part request")
var ErrPartRequestViewForbidden = errors.New("part request list is not available for this role")

type PartRequestService interface {
	Create(ctx context.Context, authorUserID int64, req dto.PartRequestCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.PartRequestResponse, error)
	List(ctx context.Context, currentUserID int64, roleID int64, q dto.PartRequestListQuery) (*dto.PartRequestListResponse, error)
	UpdateByID(ctx context.Context, id int64, changedByUserID int64, roleID int64, req dto.PartRequestUpdateRequest) (bool, error)
	UpdateStatusByID(ctx context.Context, id int64, changedByUserID int64, roleID int64, req dto.PartRequestStatusUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64, changedByUserID int64) (bool, error)
	ListStatuses(ctx context.Context) ([]dto.PartRequestStatusResponse, error)
	ListHistoryByRequestID(ctx context.Context, id int64) (*dto.PartRequestHistoryListResponse, error)
	ListHistory(ctx context.Context, q dto.PartRequestHistoryListQuery) (*dto.PartRequestHistoryListResponse, error)
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
		HistoryComment:  "Заявка создана",
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

	history, err := s.repo.ListHistoryByRequestID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := mapPartRequestToDTO(*item)
	resp.History = mapPartRequestHistoryListToDTO(history)
	return &resp, nil
}

func (s *partRequestService) List(ctx context.Context, currentUserID int64, roleID int64, q dto.PartRequestListQuery) (*dto.PartRequestListResponse, error) {
	if currentUserID <= 0 {
		return nil, fmt.Errorf("invalid current user id")
	}

	if q.PartID < 0 {
		return nil, fmt.Errorf("part_id cannot be negative")
	}
	if q.StatusID < 0 {
		return nil, fmt.Errorf("status_id cannot be negative")
	}
	if q.AuthorUserID < 0 {
		return nil, fmt.Errorf("author_user_id cannot be negative")
	}

	dateFrom, err := normalizePartRequestOptionalDate(q.DateFrom, "date_from")
	if err != nil {
		return nil, err
	}

	dateTo, err := normalizePartRequestOptionalDate(q.DateTo, "date_to")
	if err != nil {
		return nil, err
	}

	effectiveAuthorUserID := q.AuthorUserID

	switch roleID {
	case AdminRoleID, warehouseManagerRoleID:
		// Админ и заведующий складом видят все заявки.
		// Если передан author_user_id, он работает как дополнительный фильтр.

	case dutyMechanicRoleID:
		// Механик всегда видит только свои заявки.
		// Даже если он передаст author_user_id другого пользователя,
		// backend всё равно заменит его на currentUserID.
		effectiveAuthorUserID = currentUserID

	default:
		return nil, ErrPartRequestViewForbidden
	}

	items, total, err := s.repo.List(ctx, repository.ListPartRequestsParams{
		PartID:       q.PartID,
		StatusID:     q.StatusID,
		StatusCode:   strings.TrimSpace(q.StatusCode),
		AuthorUserID: effectiveAuthorUserID,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
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

	return &dto.PartRequestListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *partRequestService) UpdateByID(ctx context.Context, id int64, changedByUserID int64, roleID int64, req dto.PartRequestUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if changedByUserID <= 0 {
		return false, fmt.Errorf("invalid changed_by_user_id")
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

	status, err := s.getStatusByID(ctx, req.StatusID)
	if err != nil {
		return false, err
	}

	rejectionComment := ""
	if status.Code == rejectedPartRequestStatusCode {
		if roleID != warehouseManagerRoleID && roleID != AdminRoleID {
			return false, ErrPartRequestRejectForbidden
		}

		rejectionComment, err = normalizeRejectionComment(req.RejectionComment)
		if err != nil {
			return false, err
		}
	}

	historyDefault := "Заявка обновлена"
	if status.Code == rejectedPartRequestStatusCode {
		historyDefault = rejectionComment
	}
	historyComment, err := normalizeHistoryComment(req.HistoryComment, historyDefault)
	if err != nil {
		return false, err
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdatePartRequestParams{
		PartID:           req.PartID,
		Quantity:         req.Quantity,
		MechanicComment:  comment,
		StatusID:         req.StatusID,
		RejectionComment: rejectionComment,
		ChangedByUserID:  changedByUserID,
		HistoryComment:   historyComment,
	})
}

func (s *partRequestService) UpdateStatusByID(ctx context.Context, id int64, changedByUserID int64, roleID int64, req dto.PartRequestStatusUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if changedByUserID <= 0 {
		return false, fmt.Errorf("invalid changed_by_user_id")
	}

	status, err := s.getStatusByID(ctx, req.StatusID)
	if err != nil {
		return false, err
	}

	rejectionComment := ""
	historyDefault := "Статус заявки изменён"

	if status.Code == rejectedPartRequestStatusCode {
		if roleID != warehouseManagerRoleID && roleID != AdminRoleID {
			return false, ErrPartRequestRejectForbidden
		}

		rejectionComment, err = normalizeRejectionComment(firstNonEmpty(req.RejectionComment, req.Comment))
		if err != nil {
			return false, err
		}
		historyDefault = rejectionComment
	}

	historyComment, err := normalizeHistoryComment(req.Comment, historyDefault)
	if err != nil {
		return false, err
	}

	return s.repo.UpdateStatusByID(ctx, id, repository.UpdatePartRequestStatusParams{
		StatusID:         req.StatusID,
		RejectionComment: rejectionComment,
		ChangedByUserID:  changedByUserID,
		HistoryComment:   historyComment,
	})
}

func (s *partRequestService) DeleteByID(ctx context.Context, id int64, changedByUserID int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if changedByUserID <= 0 {
		return false, fmt.Errorf("invalid changed_by_user_id")
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

	return s.repo.DeleteByID(ctx, id, repository.DeletePartRequestParams{
		ChangedByUserID: changedByUserID,
		HistoryComment:  "Заявка удалена",
	})
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

func (s *partRequestService) ListHistoryByRequestID(ctx context.Context, id int64) (*dto.PartRequestHistoryListResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	items, err := s.repo.ListHistoryByRequestID(ctx, id)
	if err != nil {
		return nil, err
	}

	out := mapPartRequestHistoryListToDTO(items)

	return &dto.PartRequestHistoryListResponse{
		Items:  out,
		Total:  int64(len(out)),
		Limit:  len(out),
		Offset: 0,
	}, nil
}

func (s *partRequestService) ListHistory(ctx context.Context, q dto.PartRequestHistoryListQuery) (*dto.PartRequestHistoryListResponse, error) {
	if q.PartRequestID < 0 {
		return nil, fmt.Errorf("part_request_id cannot be negative")
	}
	if q.StatusID < 0 {
		return nil, fmt.Errorf("status_id cannot be negative")
	}
	if q.ChangedByUserID < 0 {
		return nil, fmt.Errorf("changed_by_user_id cannot be negative")
	}

	dateFrom, err := normalizePartRequestOptionalDate(q.DateFrom, "date_from")
	if err != nil {
		return nil, err
	}

	dateTo, err := normalizePartRequestOptionalDate(q.DateTo, "date_to")
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListHistory(ctx, repository.ListPartRequestHistoryParams{
		PartRequestID:   q.PartRequestID,
		StatusID:        q.StatusID,
		StatusCode:      strings.TrimSpace(q.StatusCode),
		ChangedByUserID: q.ChangedByUserID,
		DateFrom:        dateFrom,
		DateTo:          dateTo,
		Limit:           q.Limit,
		Offset:          q.Offset,
		SortBy:          q.SortBy,
		Order:           q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := mapPartRequestHistoryListToDTO(items)

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.PartRequestHistoryListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
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

func (s *partRequestService) getStatusByID(ctx context.Context, statusID int64) (*models.PartRequestStatus, error) {
	if statusID <= 0 {
		return nil, fmt.Errorf("status_id is required")
	}

	status, err := s.repo.GetStatusByID(ctx, statusID)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, repository.ErrPartRequestStatusNotFound
	}

	return status, nil
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

func normalizeHistoryComment(value string, defaultValue string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = defaultValue
	}
	if len([]rune(trimmed)) > 2000 {
		return "", fmt.Errorf("history comment must be less than or equal to 2000 characters")
	}
	return trimmed, nil
}

func normalizeRejectionComment(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", ErrPartRequestRejectionCommentRequired
	}
	if len([]rune(trimmed)) > 2000 {
		return "", fmt.Errorf("rejection comment must be less than or equal to 2000 characters")
	}
	return trimmed, nil
}

func normalizePartRequestOptionalDate(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}

	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}

	return trimmed, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
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
		Quantity:         item.Quantity,
		MechanicComment:  item.MechanicComment,
		RejectionComment: item.RejectionComment,
		StatusID:         item.StatusID,
		Status: dto.PartRequestStatusResponse{
			ID:   item.StatusID,
			Code: item.StatusCode,
			Name: item.StatusName,
		},
		AuthorUserID:   item.AuthorUserID,
		AuthorEmail:    item.AuthorEmail,
		AuthorFullName: item.AuthorFullName,
		History:        []dto.PartRequestHistoryResponse{},
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

func mapPartRequestHistoryToDTO(item models.PartRequestHistory) dto.PartRequestHistoryResponse {
	return dto.PartRequestHistoryResponse{
		ID:            item.ID,
		PartRequestID: item.PartRequestID,
		StatusID:      item.StatusID,
		Status: dto.PartRequestStatusResponse{
			ID:   item.StatusID,
			Code: item.StatusCode,
			Name: item.StatusName,
		},
		ChangedByUserID:   item.ChangedByUserID,
		ChangedByEmail:    item.ChangedByEmail,
		ChangedByFullName: item.ChangedByFullName,
		Comment:           item.Comment,
		ChangedAt:         item.ChangedAt,
	}
}

func mapPartRequestHistoryListToDTO(items []models.PartRequestHistory) []dto.PartRequestHistoryResponse {
	out := make([]dto.PartRequestHistoryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartRequestHistoryToDTO(item))
	}
	return out
}
