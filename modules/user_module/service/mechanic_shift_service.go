package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/models"
	"auto_park/modules/user_module/repository"
)

const (
	mechanicShiftRoleAdmin        int64 = 1
	mechanicShiftRoleManager      int64 = 2
	mechanicShiftRoleDispatcher   int64 = 3
	mechanicShiftRoleDutyMechanic int64 = 4
)

var (
	ErrMechanicShiftAccessDenied = errors.New("mechanic shift access denied")
	ErrMechanicShiftInvalidTime  = errors.New("time_to must be greater than time_from")
)

type MechanicShiftService struct {
	repo repository.MechanicShiftRepository
}

func NewMechanicShiftService(repo repository.MechanicShiftRepository) *MechanicShiftService {
	return &MechanicShiftService{repo: repo}
}

func (s *MechanicShiftService) Create(ctx context.Context, currentUserID int64, roleID int64, req dto.CreateMechanicShiftRequest) (int64, error) {
	if currentUserID <= 0 {
		return 0, fmt.Errorf("invalid current user id")
	}

	userID := currentUserID
	if req.UserID != nil && *req.UserID > 0 {
		userID = *req.UserID
	}

	if roleID == mechanicShiftRoleDutyMechanic && userID != currentUserID {
		return 0, ErrMechanicShiftAccessDenied
	}
	if roleID != mechanicShiftRoleAdmin && roleID != mechanicShiftRoleDutyMechanic {
		return 0, ErrMechanicShiftAccessDenied
	}

	if err := s.ensureDutyMechanicUser(ctx, userID); err != nil {
		return 0, err
	}

	shiftDate, err := normalizeShiftDate(req.ShiftDate)
	if err != nil {
		return 0, err
	}

	timeFrom, err := normalizeShiftTime(req.TimeFrom, "time_from")
	if err != nil {
		return 0, err
	}

	var timeTo *string
	if req.TimeTo != nil {
		v, err := normalizeOptionalShiftTime(*req.TimeTo, "time_to")
		if err != nil {
			return 0, err
		}
		timeTo = v
	}

	if err := validateShiftTimeRange(timeFrom, timeTo); err != nil {
		return 0, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	comment := normalizeOptionalTextPtr(req.Comment)

	return s.repo.Create(ctx, repository.CreateMechanicShiftParams{
		UserID:    userID,
		ShiftDate: shiftDate,
		TimeFrom:  timeFrom,
		TimeTo:    timeTo,
		Comment:   comment,
		IsActive:  isActive,
	})
}

func (s *MechanicShiftService) GetByID(ctx context.Context, currentUserID int64, roleID int64, id int64) (*dto.MechanicShiftResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return nil, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !s.canViewShift(currentUserID, roleID, item.UserID) {
		return nil, ErrMechanicShiftAccessDenied
	}

	resp := mapMechanicShiftToDTO(*item)
	return &resp, nil
}

func (s *MechanicShiftService) List(ctx context.Context, currentUserID int64, roleID int64, q dto.MechanicShiftListQuery) (*dto.MechanicShiftListResponse, error) {
	if currentUserID <= 0 {
		return nil, fmt.Errorf("invalid current user id")
	}

	if q.UserID < 0 {
		return nil, fmt.Errorf("user_id cannot be negative")
	}

	dateFrom, err := normalizeOptionalShiftDate(q.DateFrom, "date_from")
	if err != nil {
		return nil, err
	}

	dateTo, err := normalizeOptionalShiftDate(q.DateTo, "date_to")
	if err != nil {
		return nil, err
	}

	userID := q.UserID
	switch roleID {
	case mechanicShiftRoleAdmin, mechanicShiftRoleManager, mechanicShiftRoleDispatcher:
	case mechanicShiftRoleDutyMechanic:
		userID = currentUserID
	default:
		return nil, ErrMechanicShiftAccessDenied
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return nil, err
	}

	items, total, err := s.repo.List(ctx, repository.ListMechanicShiftsParams{
		UserID:   userID,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		IsActive: q.IsActive,
		Limit:    q.Limit,
		Offset:   q.Offset,
		SortBy:   q.SortBy,
		Order:    q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.MechanicShiftResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapMechanicShiftToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.MechanicShiftListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *MechanicShiftService) UpdateByID(ctx context.Context, currentUserID int64, roleID int64, id int64, req dto.UpdateMechanicShiftRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return false, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	if !s.canManageShift(currentUserID, roleID, current.UserID) {
		return false, ErrMechanicShiftAccessDenied
	}

	params := repository.UpdateMechanicShiftParams{}

	if req.UserID != nil {
		if roleID != mechanicShiftRoleAdmin {
			return false, ErrMechanicShiftAccessDenied
		}
		if *req.UserID <= 0 {
			return false, fmt.Errorf("user_id must be greater than 0")
		}
		if err := s.ensureDutyMechanicUser(ctx, *req.UserID); err != nil {
			return false, err
		}
		params.UserID = req.UserID
	}

	if req.ShiftDate != nil {
		v, err := normalizeShiftDate(*req.ShiftDate)
		if err != nil {
			return false, err
		}
		params.ShiftDate = &v
	}

	if req.TimeFrom != nil {
		v, err := normalizeShiftTime(*req.TimeFrom, "time_from")
		if err != nil {
			return false, err
		}
		params.TimeFrom = &v
	}

	if req.TimeTo != nil {
		v, err := normalizeOptionalShiftTime(*req.TimeTo, "time_to")
		if err != nil {
			return false, err
		}
		params.TimeTo = v
	}

	if req.Comment != nil {
		params.Comment = normalizeOptionalTextPtr(req.Comment)
	}

	if req.IsActive != nil {
		params.IsActive = req.IsActive
	}

	timeFromForValidation := current.TimeFrom
	if params.TimeFrom != nil {
		timeFromForValidation = *params.TimeFrom
	}

	timeToForValidation := current.TimeTo
	if req.TimeTo != nil {
		timeToForValidation = params.TimeTo
	}

	if err := validateShiftTimeRange(timeFromForValidation, timeToForValidation); err != nil {
		return false, err
	}

	return s.repo.UpdateByID(ctx, id, params)
}

func (s *MechanicShiftService) UpdateActivityByID(ctx context.Context, currentUserID int64, roleID int64, id int64, req dto.UpdateMechanicShiftActivityRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return false, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	if !s.canManageShift(currentUserID, roleID, current.UserID) {
		return false, ErrMechanicShiftAccessDenied
	}

	return s.repo.UpdateActivityByID(ctx, id, req.IsActive)
}

func (s *MechanicShiftService) DeleteByID(ctx context.Context, currentUserID int64, roleID int64, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return false, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	if !s.canManageShift(currentUserID, roleID, current.UserID) {
		return false, ErrMechanicShiftAccessDenied
	}

	return s.repo.SoftDeleteByID(ctx, id, currentUserID)
}

func (s *MechanicShiftService) ensureDutyMechanicUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("user_id is required")
	}

	exists, err := s.repo.UserIsDutyMechanic(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrMechanicShiftUserNotMechanic
	}

	return nil
}

func (s *MechanicShiftService) canViewShift(currentUserID int64, roleID int64, shiftUserID int64) bool {
	switch roleID {
	case mechanicShiftRoleAdmin, mechanicShiftRoleManager, mechanicShiftRoleDispatcher:
		return true
	case mechanicShiftRoleDutyMechanic:
		return currentUserID == shiftUserID
	default:
		return false
	}
}

func (s *MechanicShiftService) canManageShift(currentUserID int64, roleID int64, shiftUserID int64) bool {
	switch roleID {
	case mechanicShiftRoleAdmin:
		return true
	case mechanicShiftRoleDutyMechanic:
		return currentUserID == shiftUserID
	default:
		return false
	}
}

func normalizeShiftDate(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("shift_date is required")
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("shift_date must be in YYYY-MM-DD format")
	}
	return trimmed, nil
}

func normalizeOptionalShiftDate(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return trimmed, nil
}

func normalizeShiftTime(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if _, err := parseShiftTime(trimmed); err != nil {
		return "", fmt.Errorf("%s must be in HH:MM or HH:MM:SS format", field)
	}
	return trimmed, nil
}

func normalizeOptionalShiftTime(value string, field string) (*string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	if _, err := parseShiftTime(trimmed); err != nil {
		return nil, fmt.Errorf("%s must be in HH:MM or HH:MM:SS format", field)
	}
	return &trimmed, nil
}

func parseShiftTime(value string) (time.Time, error) {
	if t, err := time.Parse("15:04", value); err == nil {
		return t, nil
	}
	return time.Parse("15:04:05", value)
}

func validateShiftTimeRange(timeFrom string, timeTo *string) error {
	if timeTo == nil || strings.TrimSpace(*timeTo) == "" {
		return nil
	}

	from, err := parseShiftTime(timeFrom)
	if err != nil {
		return err
	}
	to, err := parseShiftTime(*timeTo)
	if err != nil {
		return err
	}
	if !to.After(from) {
		return ErrMechanicShiftInvalidTime
	}
	return nil
}

func normalizeOptionalTextPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func mapMechanicShiftToDTO(item models.MechanicShift) dto.MechanicShiftResponse {
	return dto.MechanicShiftResponse{
		ID:     item.ID,
		UserID: item.UserID,
		User: dto.MechanicShiftUserBriefResponse{
			ID:         item.UserID,
			Email:      item.UserEmail,
			FirstName:  item.UserFirstName,
			LastName:   item.UserLastName,
			MiddleName: item.UserMiddleName,
			RoleID:     item.UserRoleID,
			RoleName:   item.UserRoleName,
		},
		ShiftDate: item.ShiftDate.Format("2006-01-02"),
		TimeFrom:  trimSeconds(item.TimeFrom),
		TimeTo:    trimTimePtr(item.TimeTo),
		Comment:   item.Comment,
		IsActive:  item.IsActive,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func trimTimePtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := trimSeconds(*value)
	return &trimmed
}

func trimSeconds(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 5 {
		return trimmed[:5]
	}
	return trimmed
}
