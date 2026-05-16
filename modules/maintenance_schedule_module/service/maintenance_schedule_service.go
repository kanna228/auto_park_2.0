package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/maintenance_schedule_module/dto"
	"auto_park/modules/maintenance_schedule_module/repository"
)

type MaintenanceScheduleService interface {
	Create(ctx context.Context, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.MaintenanceScheduleResponse, error)
	List(ctx context.Context, q dto.MaintenanceScheduleListQuery) (*dto.MaintenanceScheduleListResponse, error)
	Update(ctx context.Context, id int64, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type maintenanceScheduleService struct {
	repo repository.MaintenanceScheduleRepository
}

func NewMaintenanceScheduleService(repo repository.MaintenanceScheduleRepository) MaintenanceScheduleService {
	return &maintenanceScheduleService{repo: repo}
}

func (s *maintenanceScheduleService) Create(ctx context.Context, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error) {
	if err := validateSchedule(req); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, req)
}

func (s *maintenanceScheduleService) GetByID(ctx context.Context, id int64) (*dto.MaintenanceScheduleResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *maintenanceScheduleService) List(ctx context.Context, q dto.MaintenanceScheduleListQuery) (*dto.MaintenanceScheduleListResponse, error) {
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}
	return &dto.MaintenanceScheduleListResponse{
		Items:  items,
		Total:  total,
		Limit:  normalizeLimit(q.Limit),
		Offset: normalizeOffset(q.Offset),
	}, nil
}

func (s *maintenanceScheduleService) Update(ctx context.Context, id int64, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if err := validateSchedule(req); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, req)
}

func (s *maintenanceScheduleService) Delete(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}

func validateSchedule(req dto.MaintenanceScheduleRequest) error {
	dateFrom, err := parseRequiredDate(req.DateFrom, "date_from")
	if err != nil {
		return err
	}
	dateTo, err := parseRequiredDate(req.DateTo, "date_to")
	if err != nil {
		return err
	}
	if dateTo.Before(dateFrom) {
		return fmt.Errorf("date_to must be greater than or equal to date_from")
	}
	if req.ConsecutiveCount <= 0 {
		return fmt.Errorf("consecutive_count must be greater than 0")
	}
	if strings.TrimSpace(req.ConsecutiveUnit) == "" {
		return fmt.Errorf("consecutive_unit is required")
	}
	if req.EveryCount <= 0 {
		return fmt.Errorf("every_count must be greater than 0")
	}
	if strings.TrimSpace(req.EveryUnit) == "" {
		return fmt.Errorf("every_unit is required")
	}
	if _, err := time.Parse("15:04", strings.TrimSpace(req.TimeFrom)); err != nil {
		return fmt.Errorf("time_from must be in HH:MM format")
	}
	if _, err := time.Parse("15:04", strings.TrimSpace(req.TimeTo)); err != nil {
		return fmt.Errorf("time_to must be in HH:MM format")
	}
	if req.DurationValue <= 0 {
		return fmt.Errorf("duration_value must be greater than 0")
	}
	if strings.TrimSpace(req.DurationUnit) == "" {
		return fmt.Errorf("duration_unit is required")
	}
	return nil
}

func parseRequiredDate(value string, field string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("%s is required", field)
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return parsed, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
