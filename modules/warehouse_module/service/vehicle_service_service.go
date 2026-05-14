package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/models"
	"auto_park/modules/warehouse_module/repository"
)

type VehicleServiceService interface {
	CreatePartsCollection(ctx context.Context, req dto.PartsCollectionCreateRequest) (int64, error)
	GetPartsCollectionByID(ctx context.Context, id int64) (*dto.PartsCollectionResponse, error)
	ListPartsCollection(ctx context.Context, q dto.PartsCollectionListQuery) (*dto.PartsCollectionListResponse, error)
	UpdatePartsCollectionByID(ctx context.Context, id int64, req dto.PartsCollectionUpdateRequest) (bool, error)
	DeletePartsCollectionByID(ctx context.Context, id int64) (bool, error)

	CreateServiceType(ctx context.Context, req dto.ServiceTypeCreateRequest) (int64, error)
	GetServiceTypeByID(ctx context.Context, id int64) (*dto.ServiceTypeResponse, error)
	ListServiceTypes(ctx context.Context, q dto.ServiceTypeListQuery) (*dto.ServiceTypeListResponse, error)
	UpdateServiceTypeByID(ctx context.Context, id int64, req dto.ServiceTypeUpdateRequest) (bool, error)
	DeleteServiceTypeByID(ctx context.Context, id int64) (bool, error)

	CreateVehicleService(ctx context.Context, req dto.VehicleServiceCreateRequest) (int64, error)
	GetVehicleServiceByID(ctx context.Context, id int64) (*dto.VehicleServiceResponse, error)
	ListVehicleServices(ctx context.Context, q dto.VehicleServiceListQuery) (*dto.VehicleServiceListResponse, error)
	UpdateVehicleServiceByID(ctx context.Context, id int64, req dto.VehicleServiceUpdateRequest) (bool, error)
	DeleteVehicleServiceByID(ctx context.Context, id int64) (bool, error)
}

type vehicleServiceService struct {
	repo repository.VehicleServiceRepository
}

func NewVehicleServiceService(repo repository.VehicleServiceRepository) VehicleServiceService {
	return &vehicleServiceService{repo: repo}
}

func (s *vehicleServiceService) CreatePartsCollection(ctx context.Context, req dto.PartsCollectionCreateRequest) (int64, error) {
	name, err := normalizeVehicleServiceRequiredText("name", req.Name)
	if err != nil {
		return 0, err
	}
	description := normalizeVehicleServiceOptionalText(req.Description)
	return s.repo.CreatePartsCollection(ctx, repository.PartsCollectionParams{Name: name, Description: description})
}

func (s *vehicleServiceService) GetPartsCollectionByID(ctx context.Context, id int64) (*dto.PartsCollectionResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	item, err := s.repo.GetPartsCollectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	resp := mapPartsCollectionToDTO(*item)
	return &resp, nil
}

func (s *vehicleServiceService) ListPartsCollection(ctx context.Context, q dto.PartsCollectionListQuery) (*dto.PartsCollectionListResponse, error) {
	items, total, err := s.repo.ListPartsCollection(ctx, repository.ListPartsCollectionParams{
		Name:   strings.TrimSpace(q.Name),
		Limit:  q.Limit,
		Offset: q.Offset,
		SortBy: q.SortBy,
		Order:  q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.PartsCollectionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartsCollectionToDTO(item))
	}
	limit, offset := normalizeVehicleServiceLimitOffset(q.Limit, q.Offset)
	return &dto.PartsCollectionListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *vehicleServiceService) UpdatePartsCollectionByID(ctx context.Context, id int64, req dto.PartsCollectionUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	name, err := normalizeVehicleServiceRequiredText("name", req.Name)
	if err != nil {
		return false, err
	}
	description := normalizeVehicleServiceOptionalText(req.Description)
	return s.repo.UpdatePartsCollectionByID(ctx, id, repository.PartsCollectionParams{Name: name, Description: description})
}

func (s *vehicleServiceService) DeletePartsCollectionByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeletePartsCollectionByID(ctx, id)
}

func (s *vehicleServiceService) CreateServiceType(ctx context.Context, req dto.ServiceTypeCreateRequest) (int64, error) {
	name, err := normalizeVehicleServiceRequiredText("name", req.Name)
	if err != nil {
		return 0, err
	}
	description := normalizeVehicleServiceOptionalText(req.Description)
	return s.repo.CreateServiceType(ctx, repository.ServiceTypeParams{Name: name, Description: description})
}

func (s *vehicleServiceService) GetServiceTypeByID(ctx context.Context, id int64) (*dto.ServiceTypeResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	item, err := s.repo.GetServiceTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	resp := mapServiceTypeToDTO(*item)
	return &resp, nil
}

func (s *vehicleServiceService) ListServiceTypes(ctx context.Context, q dto.ServiceTypeListQuery) (*dto.ServiceTypeListResponse, error) {
	items, total, err := s.repo.ListServiceTypes(ctx, repository.ListServiceTypesParams{
		Name:   strings.TrimSpace(q.Name),
		Limit:  q.Limit,
		Offset: q.Offset,
		SortBy: q.SortBy,
		Order:  q.Order,
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.ServiceTypeResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapServiceTypeToDTO(item))
	}
	limit, offset := normalizeVehicleServiceLimitOffset(q.Limit, q.Offset)
	return &dto.ServiceTypeListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *vehicleServiceService) UpdateServiceTypeByID(ctx context.Context, id int64, req dto.ServiceTypeUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	name, err := normalizeVehicleServiceRequiredText("name", req.Name)
	if err != nil {
		return false, err
	}
	description := normalizeVehicleServiceOptionalText(req.Description)
	return s.repo.UpdateServiceTypeByID(ctx, id, repository.ServiceTypeParams{Name: name, Description: description})
}

func (s *vehicleServiceService) DeleteServiceTypeByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeleteServiceTypeByID(ctx, id)
}

func (s *vehicleServiceService) CreateVehicleService(ctx context.Context, req dto.VehicleServiceCreateRequest) (int64, error) {
	params, err := s.normalizeVehicleServiceParams(ctx, req.TypeID, req.PartID, req.VehicleID, req.Date)
	if err != nil {
		return 0, err
	}
	return s.repo.CreateVehicleService(ctx, repository.CreateVehicleServiceParams(params))
}

func (s *vehicleServiceService) GetVehicleServiceByID(ctx context.Context, id int64) (*dto.VehicleServiceResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	item, err := s.repo.GetVehicleServiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	resp := mapVehicleServiceToDTO(*item)
	return &resp, nil
}

func (s *vehicleServiceService) ListVehicleServices(ctx context.Context, q dto.VehicleServiceListQuery) (*dto.VehicleServiceListResponse, error) {
	if q.TypeID < 0 {
		return nil, fmt.Errorf("type_id cannot be negative")
	}
	if q.PartID < 0 {
		return nil, fmt.Errorf("part_id cannot be negative")
	}
	if q.VehicleID < 0 {
		return nil, fmt.Errorf("vehicle_id cannot be negative")
	}
	dateFrom, err := normalizeVehicleServiceOptionalDate(q.DateFrom, "date_from")
	if err != nil {
		return nil, err
	}
	dateTo, err := normalizeVehicleServiceOptionalDate(q.DateTo, "date_to")
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListVehicleServices(ctx, repository.ListVehicleServicesParams{
		TypeID:    q.TypeID,
		PartID:    q.PartID,
		VehicleID: q.VehicleID,
		TypeName:  strings.TrimSpace(q.TypeName),
		PartName:  strings.TrimSpace(q.PartName),
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Limit:     q.Limit,
		Offset:    q.Offset,
		SortBy:    q.SortBy,
		Order:     q.Order,
	})
	if err != nil {
		return nil, err
	}
	out := make([]dto.VehicleServiceResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapVehicleServiceToDTO(item))
	}
	limit, offset := normalizeVehicleServiceLimitOffset(q.Limit, q.Offset)
	return &dto.VehicleServiceListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *vehicleServiceService) UpdateVehicleServiceByID(ctx context.Context, id int64, req dto.VehicleServiceUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	params, err := s.normalizeVehicleServiceParams(ctx, req.TypeID, req.PartID, req.VehicleID, req.Date)
	if err != nil {
		return false, err
	}
	return s.repo.UpdateVehicleServiceByID(ctx, id, repository.UpdateVehicleServiceParams(params))
}

func (s *vehicleServiceService) DeleteVehicleServiceByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeleteVehicleServiceByID(ctx, id)
}

type normalizedVehicleServiceParams struct {
	TypeID      int64
	PartID      int64
	VehicleID   int64
	ServiceDate string
}

func (s *vehicleServiceService) normalizeVehicleServiceParams(ctx context.Context, typeID int64, partID int64, vehicleID int64, date string) (normalizedVehicleServiceParams, error) {
	if typeID <= 0 {
		return normalizedVehicleServiceParams{}, fmt.Errorf("type_id is required")
	}
	if partID <= 0 {
		return normalizedVehicleServiceParams{}, fmt.Errorf("part_id is required")
	}
	if vehicleID <= 0 {
		return normalizedVehicleServiceParams{}, fmt.Errorf("vehicle_id is required")
	}
	serviceDate, err := normalizeVehicleServiceDate(date, "date")
	if err != nil {
		return normalizedVehicleServiceParams{}, err
	}

	exists, err := s.repo.ServiceTypeExists(ctx, typeID)
	if err != nil {
		return normalizedVehicleServiceParams{}, err
	}
	if !exists {
		return normalizedVehicleServiceParams{}, repository.ErrVehicleServiceTypeNotFound
	}

	exists, err = s.repo.PartsCollectionExists(ctx, partID)
	if err != nil {
		return normalizedVehicleServiceParams{}, err
	}
	if !exists {
		return normalizedVehicleServiceParams{}, repository.ErrVehicleServicePartNotFound
	}

	exists, err = s.repo.VehicleExists(ctx, vehicleID)
	if err != nil {
		return normalizedVehicleServiceParams{}, err
	}
	if !exists {
		return normalizedVehicleServiceParams{}, repository.ErrVehicleServiceVehicleNotFound
	}

	return normalizedVehicleServiceParams{TypeID: typeID, PartID: partID, VehicleID: vehicleID, ServiceDate: serviceDate}, nil
}

func normalizeVehicleServiceRequiredText(field string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if len([]rune(trimmed)) > 255 {
		return "", fmt.Errorf("%s must be less than or equal to 255 characters", field)
	}
	return trimmed, nil
}

func normalizeVehicleServiceOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeVehicleServiceDate(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return trimmed, nil
}

func normalizeVehicleServiceOptionalDate(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	return normalizeVehicleServiceDate(trimmed, field)
}

func normalizeVehicleServiceLimitOffset(limit int, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func mapPartsCollectionToDTO(item models.PartsCollection) dto.PartsCollectionResponse {
	return dto.PartsCollectionResponse{ID: item.ID, Name: item.Name, Description: item.Description, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func mapServiceTypeToDTO(item models.ServiceType) dto.ServiceTypeResponse {
	return dto.ServiceTypeResponse{ID: item.ID, Name: item.Name, Description: item.Description, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func mapVehicleServiceToDTO(item models.VehicleService) dto.VehicleServiceResponse {
	return dto.VehicleServiceResponse{
		ID:     item.ID,
		TypeID: item.TypeID,
		Type: dto.VehicleServiceTypeBriefResponse{
			ID:          item.TypeID,
			Name:        item.TypeName,
			Description: item.TypeDescription,
		},
		PartID: item.PartID,
		Part: dto.VehicleServicePartBriefResponse{
			ID:          item.PartID,
			Name:        item.PartName,
			Description: item.PartDescription,
		},
		VehicleID: item.VehicleID,
		Vehicle: dto.VehicleServiceVehicleBriefResponse{
			ID:          item.VehicleID,
			StateNumber: item.VehicleStateNumber,
			BrandModel:  item.VehicleBrandModel,
		},
		Date:      item.ServiceDate.Format("2006-01-02"),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
