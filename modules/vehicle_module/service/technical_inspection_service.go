package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/models"
	"auto_park/modules/vehicle_module/repository"
)

type TechnicalInspectionService interface {
	Create(ctx context.Context, req dto.TechnicalInspectionCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.TechnicalInspectionResponse, error)
	List(ctx context.Context, q dto.TechnicalInspectionListQuery) (*dto.TechnicalInspectionListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.TechnicalInspectionUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	UploadFile(ctx context.Context, id int64, fh *multipart.FileHeader) (*dto.TechnicalInspectionResponse, error)
	DeleteFile(ctx context.Context, id int64) (*dto.TechnicalInspectionResponse, error)
}

type technicalInspectionService struct {
	repo    repository.TechnicalInspectionRepository
	storage *VehicleDocumentStorage
}

func NewTechnicalInspectionService(repo repository.TechnicalInspectionRepository, storage *VehicleDocumentStorage) TechnicalInspectionService {
	return &technicalInspectionService{
		repo:    repo,
		storage: storage,
	}
}

func (s *technicalInspectionService) Create(ctx context.Context, req dto.TechnicalInspectionCreateRequest) (int64, error) {
	req.Name = strings.TrimSpace(req.Name)

	if req.VehicleID <= 0 {
		return 0, fmt.Errorf("vehicle_id is required")
	}
	if req.Name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if req.IsActive == nil {
		return 0, fmt.Errorf("is_active is required")
	}

	startDate, err := parseFlexibleDate(req.StartDate)
	if err != nil {
		return 0, fmt.Errorf("invalid start_date")
	}

	endDate, err := parseFlexibleDate(req.EndDate)
	if err != nil {
		return 0, fmt.Errorf("invalid end_date")
	}

	if endDate.Before(startDate) {
		return 0, fmt.Errorf("end_date must be greater than or equal to start_date")
	}

	return s.repo.Create(ctx, repository.CreateTechnicalInspectionParams{
		VehicleID: req.VehicleID,
		Name:      req.Name,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		IsActive:  *req.IsActive,
	})
}

func (s *technicalInspectionService) GetByID(ctx context.Context, id int64) (*dto.TechnicalInspectionResponse, error) {
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

	resp := mapTechnicalInspectionToDTO(*item)
	return &resp, nil
}

func (s *technicalInspectionService) List(ctx context.Context, q dto.TechnicalInspectionListQuery) (*dto.TechnicalInspectionListResponse, error) {
	items, total, err := s.repo.List(ctx, repository.ListTechnicalInspectionParams{
		VehicleID: q.VehicleID,
		IsActive:  q.IsActive,
		Name:      q.Name,
		Limit:     q.Limit,
		Offset:    q.Offset,
		SortBy:    q.SortBy,
		Order:     q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.TechnicalInspectionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapTechnicalInspectionToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.TechnicalInspectionListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *technicalInspectionService) UpdateByID(ctx context.Context, id int64, req dto.TechnicalInspectionUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	req.Name = strings.TrimSpace(req.Name)

	if req.VehicleID <= 0 {
		return false, fmt.Errorf("vehicle_id is required")
	}
	if req.Name == "" {
		return false, fmt.Errorf("name is required")
	}
	if req.IsActive == nil {
		return false, fmt.Errorf("is_active is required")
	}

	startDate, err := parseFlexibleDate(req.StartDate)
	if err != nil {
		return false, fmt.Errorf("invalid start_date")
	}

	endDate, err := parseFlexibleDate(req.EndDate)
	if err != nil {
		return false, fmt.Errorf("invalid end_date")
	}

	if endDate.Before(startDate) {
		return false, fmt.Errorf("end_date must be greater than or equal to start_date")
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdateTechnicalInspectionParams{
		VehicleID: req.VehicleID,
		Name:      req.Name,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		IsActive:  *req.IsActive,
	})
}

func (s *technicalInspectionService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}

	if strings.TrimSpace(item.FilePath) != "" {
		if err := s.storage.Delete(item.FilePath); err != nil {
			return false, err
		}
	}

	return s.repo.DeleteByID(ctx, id)
}

func (s *technicalInspectionService) UploadFile(ctx context.Context, id int64, fh *multipart.FileHeader) (*dto.TechnicalInspectionResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if fh == nil {
		return nil, fmt.Errorf("file is required")
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	relativePath, err := s.storage.Save("technical-inspection", id, fh, item.FilePath)
	if err != nil {
		return nil, err
	}

	ok, err := s.repo.UpdateFilePath(ctx, id, relativePath)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}

	resp := mapTechnicalInspectionToDTO(*updated)
	return &resp, nil
}

func (s *technicalInspectionService) DeleteFile(ctx context.Context, id int64) (*dto.TechnicalInspectionResponse, error) {
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

	if strings.TrimSpace(item.FilePath) != "" {
		if err := s.storage.Delete(item.FilePath); err != nil {
			return nil, err
		}
	}

	ok, err := s.repo.UpdateFilePath(ctx, id, "")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}

	resp := mapTechnicalInspectionToDTO(*updated)
	return &resp, nil
}

func mapTechnicalInspectionToDTO(item models.TechnicalInspection) dto.TechnicalInspectionResponse {
	return dto.TechnicalInspectionResponse{
		ID:        item.ID,
		VehicleID: item.VehicleID,
		Name:      item.Name,
		StartDate: item.StartDate,
		EndDate:   item.EndDate,
		FilePath:  item.FilePath,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
