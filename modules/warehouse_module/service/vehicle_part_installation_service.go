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

var ErrVehiclePartInstallationInstalledAtLocked = errors.New("installed_at cannot be changed after save")

type VehiclePartInstallationService interface {
	Create(ctx context.Context, userID int64, req dto.VehiclePartInstallationCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.VehiclePartInstallationResponse, error)
	List(ctx context.Context, q dto.VehiclePartInstallationListQuery) (*dto.VehiclePartInstallationListResponse, error)
	ListHistory(ctx context.Context, partID int64, limit int, offset int) (*dto.VehiclePartInstallationHistoryResponse, error)
	UpdateByID(ctx context.Context, id int64, fallbackUserID int64, roleID int64, req dto.VehiclePartInstallationUpdateRequest) (bool, error)
	UpdateActivityByID(ctx context.Context, id int64, req dto.VehiclePartInstallationActivityUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type vehiclePartInstallationService struct {
	repo repository.VehiclePartInstallationRepository
}

func NewVehiclePartInstallationService(repo repository.VehiclePartInstallationRepository) VehiclePartInstallationService {
	return &vehiclePartInstallationService{repo: repo}
}

func (s *vehiclePartInstallationService) Create(ctx context.Context, userID int64, req dto.VehiclePartInstallationCreateRequest) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid installer user id")
	}
	if req.PartID <= 0 {
		return 0, fmt.Errorf("part_id is required")
	}
	if req.VehicleID <= 0 {
		return 0, fmt.Errorf("vehicle_id is required")
	}
	if req.MechanicShiftID <= 0 {
		return 0, fmt.Errorf("mechanic_shift_id is required")
	}
	if req.Quantity <= 0 {
		return 0, fmt.Errorf("quantity must be greater than 0")
	}

	installedAt, err := normalizeDate(req.InstalledAt, "installed_at")
	if err != nil {
		return 0, err
	}

	plannedReplacementAt, err := normalizeDate(req.PlannedReplacementAt, "planned_replacement_at")
	if err != nil {
		return 0, err
	}

	if plannedReplacementAt < installedAt {
		return 0, fmt.Errorf("planned_replacement_at cannot be earlier than installed_at")
	}

	return s.repo.Create(ctx, repository.CreateVehiclePartInstallationParams{
		PartID:               req.PartID,
		VehicleID:            req.VehicleID,
		MechanicShiftID:      req.MechanicShiftID,
		InstalledAt:          installedAt,
		PlannedReplacementAt: plannedReplacementAt,
		Quantity:             req.Quantity,
		InstalledByUserID:    userID,
	})
}

func (s *vehiclePartInstallationService) GetByID(ctx context.Context, id int64) (*dto.VehiclePartInstallationResponse, error) {
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

	resp := mapVehiclePartInstallationToDTO(*item)
	return &resp, nil
}

func (s *vehiclePartInstallationService) List(ctx context.Context, q dto.VehiclePartInstallationListQuery) (*dto.VehiclePartInstallationListResponse, error) {
	dateFrom, err := normalizeOptionalDate(q.DateFrom, "date_from")
	if err != nil {
		return nil, err
	}

	dateTo, err := normalizeOptionalDate(q.DateTo, "date_to")
	if err != nil {
		return nil, err
	}

	replacementFrom, err := normalizeOptionalDate(q.ReplacementFrom, "replacement_from")
	if err != nil {
		return nil, err
	}

	replacementTo, err := normalizeOptionalDate(q.ReplacementTo, "replacement_to")
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.List(ctx, repository.ListVehiclePartInstallationsParams{
		PartID:            q.PartID,
		VehicleID:         q.VehicleID,
		MechanicShiftID:   q.MechanicShiftID,
		InstalledByUserID: q.InstalledByUserID,
		IsActive:          q.IsActive,
		DateFrom:          dateFrom,
		DateTo:            dateTo,
		ReplacementFrom:   replacementFrom,
		ReplacementTo:     replacementTo,
		Limit:             q.Limit,
		Offset:            q.Offset,
		SortBy:            q.SortBy,
		Order:             q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.VehiclePartInstallationResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapVehiclePartInstallationToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.VehiclePartInstallationListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *vehiclePartInstallationService) ListHistory(ctx context.Context, partID int64, limit int, offset int) (*dto.VehiclePartInstallationHistoryResponse, error) {
	if partID <= 0 {
		return nil, fmt.Errorf("part_id is required")
	}
	items, total, err := s.repo.ListHistory(ctx, repository.ListVehiclePartInstallationHistoryParams{PartID: partID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	out := make([]dto.VehiclePartInstallationHistoryItem, 0, len(items))
	for _, item := range items {
		out = append(out, dto.VehiclePartInstallationHistoryItem{
			EventType:     item.EventType,
			Vehicle:       item.Vehicle,
			PartRequestID: item.PartRequestID,
			CreatedAt:     item.CreatedAt,
			Actor:         item.Actor,
		})
	}
	return &dto.VehiclePartInstallationHistoryResponse{Items: out, Total: total, Limit: normalizeLimit(limit), Offset: normalizeOffset(offset)}, nil
}

func (s *vehiclePartInstallationService) UpdateByID(ctx context.Context, id int64, fallbackUserID int64, roleID int64, req dto.VehiclePartInstallationUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if req.PartID <= 0 {
		return false, fmt.Errorf("part_id is required")
	}
	if req.VehicleID <= 0 {
		return false, fmt.Errorf("vehicle_id is required")
	}
	if req.MechanicShiftID <= 0 {
		return false, fmt.Errorf("mechanic_shift_id is required")
	}
	if req.Quantity <= 0 {
		return false, fmt.Errorf("quantity must be greater than 0")
	}

	installedAt, err := normalizeDate(req.InstalledAt, "installed_at")
	if err != nil {
		return false, err
	}

	plannedReplacementAt, err := normalizeDate(req.PlannedReplacementAt, "planned_replacement_at")
	if err != nil {
		return false, err
	}

	if plannedReplacementAt < installedAt {
		return false, fmt.Errorf("planned_replacement_at cannot be earlier than installed_at")
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if current == nil {
		return false, nil
	}
	if roleID != 1 && current.InstalledAt.Format("2006-01-02") != installedAt {
		return false, ErrVehiclePartInstallationInstalledAtLocked
	}

	installedByUserID := current.InstalledByUserID
	if installedByUserID <= 0 {
		installedByUserID = fallbackUserID
	}

	if installedByUserID <= 0 {
		return false, fmt.Errorf("installed_by_user_id is required")
	}

	isActive := current.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdateVehiclePartInstallationParams{
		PartID:               req.PartID,
		VehicleID:            req.VehicleID,
		MechanicShiftID:      req.MechanicShiftID,
		InstalledAt:          installedAt,
		PlannedReplacementAt: plannedReplacementAt,
		Quantity:             req.Quantity,
		InstalledByUserID:    installedByUserID,
		IsActive:             isActive,
	})
}

func (s *vehiclePartInstallationService) UpdateActivityByID(ctx context.Context, id int64, req dto.VehiclePartInstallationActivityUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	return s.repo.UpdateActivityByID(ctx, id, req.IsActive)
}

func (s *vehiclePartInstallationService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	return s.repo.DeleteByID(ctx, id)
}

func normalizeDate(value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}

	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}

	return trimmed, nil
}

func normalizeOptionalDate(value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}

	return normalizeDate(trimmed, field)
}

func mapVehiclePartInstallationToDTO(item models.VehiclePartInstallation) dto.VehiclePartInstallationResponse {
	resp := dto.VehiclePartInstallationResponse{
		ID:     item.ID,
		PartID: item.PartID,
		Part: dto.VehiclePartInstallationPartBriefResponse{
			ID:            item.PartID,
			CatalogPartID: item.PartCatalogCode,
			Name:          item.PartName,
			Category:      item.PartCategory,
			IsConsumable:  item.PartIsConsumable,
		},
		VehicleID: item.VehicleID,
		Vehicle: dto.VehiclePartInstallationVehicleBriefResponse{
			ID:          item.VehicleID,
			StateNumber: item.VehicleStateNumber,
			BrandModel:  item.VehicleBrandModel,
		},
		MechanicShiftID:      item.MechanicShiftID,
		InstalledAt:          item.InstalledAt.Format("2006-01-02"),
		PlannedReplacementAt: item.PlannedReplacementAt.Format("2006-01-02"),
		Quantity:             item.Quantity,
		UnitPrice:            item.UnitPrice,
		TotalPrice:           item.TotalPrice,
		InstalledByUserID:    item.InstalledByUserID,
		InstallerEmail:       item.InstallerEmail,
		InstallerFullName:    item.InstallerFullName,
		IsActive:             item.IsActive,
		CreatedAt:            item.CreatedAt,
		UpdatedAt:            item.UpdatedAt,
	}

	if item.MechanicShiftID != nil && item.MechanicShiftUserID != nil && item.MechanicShiftDate != nil && item.MechanicShiftTimeFrom != nil {
		resp.MechanicShift = &dto.VehiclePartInstallationMechanicShiftBriefResponse{
			ID:               *item.MechanicShiftID,
			UserID:           *item.MechanicShiftUserID,
			ShiftDate:        item.MechanicShiftDate.Format("2006-01-02"),
			TimeFrom:         item.MechanicShiftTimeFrom.Format("15:04:05"),
			TimeTo:           formatOptionalTime(item.MechanicShiftTimeTo),
			MechanicEmail:    item.MechanicShiftUserEmail,
			MechanicFullName: item.MechanicShiftUserFullName,
		}
	}

	return resp
}

func formatOptionalTime(v *time.Time) *string {
	if v == nil {
		return nil
	}
	formatted := v.Format("15:04:05")
	return &formatted
}
