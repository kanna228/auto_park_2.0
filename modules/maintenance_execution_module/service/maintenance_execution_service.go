package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/modules/maintenance_execution_module/dto"
	"auto_park/modules/maintenance_execution_module/repository"
)

type MaintenanceExecutionService interface {
	Create(ctx context.Context, req dto.MaintenanceExecutionRequest) (*dto.MaintenanceExecutionResponse, error)
	BulkCreate(ctx context.Context, req dto.MaintenanceExecutionBulkRequest) (*dto.MaintenanceExecutionListResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.MaintenanceExecutionResponse, error)
	ListBySchedule(ctx context.Context, scheduleID int64, limit int, offset int) (*dto.MaintenanceExecutionListResponse, error)
	Update(ctx context.Context, id int64, req dto.MaintenanceExecutionRequest) (*dto.MaintenanceExecutionResponse, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type maintenanceExecutionService struct {
	repo repository.MaintenanceExecutionRepository
}

func NewMaintenanceExecutionService(repo repository.MaintenanceExecutionRepository) MaintenanceExecutionService {
	return &maintenanceExecutionService{repo: repo}
}

func (s *maintenanceExecutionService) Create(ctx context.Context, req dto.MaintenanceExecutionRequest) (*dto.MaintenanceExecutionResponse, error) {
	p, err := mapCreateRequest(req)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, p)
}

func (s *maintenanceExecutionService) BulkCreate(ctx context.Context, req dto.MaintenanceExecutionBulkRequest) (*dto.MaintenanceExecutionListResponse, error) {
	if req.ScheduleID <= 0 {
		return nil, fmt.Errorf("schedule_id is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("items is required")
	}
	responsibleMechanic := optionalString(req.ResponsibleMechanic)
	items := make([]repository.CreateMaintenanceExecutionParams, 0, len(req.Items))
	for _, item := range req.Items {
		board, action, comment, err := validateExecutionFields(item.Board, item.Action, item.Comment)
		if err != nil {
			return nil, err
		}
		items = append(items, repository.CreateMaintenanceExecutionParams{
			ScheduleID:          req.ScheduleID,
			Board:               board,
			Action:              action,
			Comment:             comment,
			ResponsibleMechanic: responsibleMechanic,
		})
	}

	created, err := s.repo.BulkCreate(ctx, items)
	if err != nil {
		return nil, err
	}
	return &dto.MaintenanceExecutionListResponse{Items: created, Total: int64(len(created)), Limit: len(created), Offset: 0}, nil
}

func (s *maintenanceExecutionService) GetByID(ctx context.Context, id int64) (*dto.MaintenanceExecutionResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *maintenanceExecutionService) ListBySchedule(ctx context.Context, scheduleID int64, limit int, offset int) (*dto.MaintenanceExecutionListResponse, error) {
	if scheduleID <= 0 {
		return nil, fmt.Errorf("schedule_id is required")
	}
	items, total, err := s.repo.ListBySchedule(ctx, scheduleID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &dto.MaintenanceExecutionListResponse{
		Items:  items,
		Total:  total,
		Limit:  normalizeLimit(limit),
		Offset: normalizeOffset(offset),
	}, nil
}

func (s *maintenanceExecutionService) Update(ctx context.Context, id int64, req dto.MaintenanceExecutionRequest) (*dto.MaintenanceExecutionResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	p, err := mapCreateRequest(req)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, p)
}

func (s *maintenanceExecutionService) Delete(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}

func mapCreateRequest(req dto.MaintenanceExecutionRequest) (repository.CreateMaintenanceExecutionParams, error) {
	if req.ScheduleID <= 0 {
		return repository.CreateMaintenanceExecutionParams{}, fmt.Errorf("schedule_id is required")
	}
	board, action, comment, err := validateExecutionFields(req.Board, req.Action, req.Comment)
	if err != nil {
		return repository.CreateMaintenanceExecutionParams{}, err
	}
	return repository.CreateMaintenanceExecutionParams{
		ScheduleID:          req.ScheduleID,
		Board:               board,
		Action:              action,
		Comment:             comment,
		ResponsibleMechanic: optionalString(req.ResponsibleMechanic),
	}, nil
}

func validateExecutionFields(board string, action string, comment string) (string, string, *string, error) {
	board = strings.TrimSpace(board)
	if board == "" {
		return "", "", nil, fmt.Errorf("board is required")
	}
	if len([]rune(board)) > 16 {
		return "", "", nil, fmt.Errorf("board must be less than or equal to 16 characters")
	}
	action = strings.TrimSpace(action)
	if !validAction(action) {
		return "", "", nil, fmt.Errorf("invalid action")
	}
	return board, action, optionalString(comment), nil
}

func validAction(action string) bool {
	switch action {
	case "serviced", "serviced_replaced", "defect_parked":
		return true
	default:
		return false
	}
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 10000 {
		return 10000
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
