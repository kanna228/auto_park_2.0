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
	driverShiftRoleAdmin      int64 = 1
	driverShiftRoleManager    int64 = 2
	driverShiftRoleDispatcher int64 = 3
)

var (
	ErrDriverShiftAccessDenied = errors.New("driver shift access denied")
	ErrDriverShiftInvalidTime  = errors.New("time_to must be greater than time_from")
)

type DriverShiftService struct {
	repo repository.DriverShiftRepository
}

func NewDriverShiftService(repo repository.DriverShiftRepository) *DriverShiftService {
	return &DriverShiftService{repo: repo}
}

func (s *DriverShiftService) Create(ctx context.Context, currentUserID int64, roleID int64, req dto.CreateDriverShiftRequest) (int64, error) {
	if currentUserID <= 0 {
		return 0, fmt.Errorf("invalid current user id")
	}
	if !s.canManage(roleID) {
		return 0, ErrDriverShiftAccessDenied
	}

	if err := s.ensureDriverExists(ctx, req.DriverID); err != nil {
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

	if err := validateDriverShiftTimeRange(timeFrom, timeTo); err != nil {
		return 0, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	comment := normalizeOptionalTextPtr(req.Comment)

	return s.repo.Create(ctx, repository.CreateDriverShiftParams{
		DriverID:  req.DriverID,
		ShiftDate: shiftDate,
		TimeFrom:  timeFrom,
		TimeTo:    timeTo,
		Comment:   comment,
		IsActive:  isActive,
	})
}

func (s *DriverShiftService) GetByID(ctx context.Context, currentUserID int64, roleID int64, id int64) (*dto.DriverShiftResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if !s.canView(roleID) {
		return nil, ErrDriverShiftAccessDenied
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return nil, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := mapDriverShiftToDTO(*item, true)
	return &resp, nil
}

func (s *DriverShiftService) List(ctx context.Context, currentUserID int64, roleID int64, q dto.DriverShiftListQuery) (*dto.DriverShiftListResponse, error) {
	if currentUserID <= 0 {
		return nil, fmt.Errorf("invalid current user id")
	}
	if !s.canView(roleID) {
		return nil, ErrDriverShiftAccessDenied
	}

	if q.DriverID < 0 {
		return nil, fmt.Errorf("driver_id cannot be negative")
	}

	dateFrom, err := normalizeOptionalShiftDate(q.DateFrom, "date_from")
	if err != nil {
		return nil, err
	}

	dateTo, err := normalizeOptionalShiftDate(q.DateTo, "date_to")
	if err != nil {
		return nil, err
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return nil, err
	}

	items, total, err := s.repo.List(ctx, repository.ListDriverShiftsParams{
		DriverID: q.DriverID,
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

	out := make([]dto.DriverShiftResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapDriverShiftToDTO(item, false))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.DriverShiftListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *DriverShiftService) UpdateByID(ctx context.Context, currentUserID int64, roleID int64, id int64, req dto.UpdateDriverShiftRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if !s.canManage(roleID) {
		return false, ErrDriverShiftAccessDenied
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return false, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	params := repository.UpdateDriverShiftParams{}

	if req.DriverID != nil {
		if *req.DriverID <= 0 {
			return false, fmt.Errorf("driver_id must be greater than 0")
		}
		if err := s.ensureDriverExists(ctx, *req.DriverID); err != nil {
			return false, err
		}
		params.DriverID = req.DriverID
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

	if err := validateDriverShiftTimeRange(timeFromForValidation, timeToForValidation); err != nil {
		return false, err
	}

	return s.repo.UpdateByID(ctx, id, params)
}

func (s *DriverShiftService) UpdateActivityByID(ctx context.Context, currentUserID int64, roleID int64, id int64, req dto.UpdateDriverShiftActivityRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if !s.canManage(roleID) {
		return false, ErrDriverShiftAccessDenied
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return false, err
	}

	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return false, err
	}

	return s.repo.UpdateActivityByID(ctx, id, req.IsActive)
}

func (s *DriverShiftService) DeleteByID(ctx context.Context, currentUserID int64, roleID int64, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if !s.canManage(roleID) {
		return false, ErrDriverShiftAccessDenied
	}

	if err := s.repo.RefreshActivity(ctx); err != nil {
		return false, err
	}

	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return false, err
	}

	return s.repo.SoftDeleteByID(ctx, id, currentUserID)
}

func (s *DriverShiftService) ensureDriverExists(ctx context.Context, driverID int64) error {
	if driverID <= 0 {
		return fmt.Errorf("driver_id is required")
	}

	exists, err := s.repo.DriverExists(ctx, driverID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrDriverShiftDriverNotFound
	}

	return nil
}

func (s *DriverShiftService) canView(roleID int64) bool {
	switch roleID {
	case driverShiftRoleAdmin, driverShiftRoleManager, driverShiftRoleDispatcher:
		return true
	default:
		return false
	}
}

func (s *DriverShiftService) canManage(roleID int64) bool {
	switch roleID {
	case driverShiftRoleAdmin, driverShiftRoleManager, driverShiftRoleDispatcher:
		return true
	default:
		return false
	}
}

func validateDriverShiftTimeRange(timeFrom string, timeTo *string) error {
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
		return ErrDriverShiftInvalidTime
	}
	return nil
}

func mapDriverShiftToDTO(item models.DriverShift, includeTripsheets bool) dto.DriverShiftResponse {
	tripsheets := []dto.DriverShiftTripsheetBriefResponse{}
	if includeTripsheets {
		tripsheets = mapDriverShiftTripsheetsToDTO(item.Tripsheets)
	}

	return dto.DriverShiftResponse{
		ID:       item.ID,
		DriverID: item.DriverID,
		Driver: dto.DriverShiftDriverBriefResponse{
			ID:         item.DriverID,
			IIN:        item.DriverIIN,
			Name:       item.DriverName,
			Surname:    item.DriverSurname,
			Middlename: item.DriverMiddlename,
			Phone:      item.DriverPhone,
			Mail:       item.DriverMail,
			StatusID:   item.DriverStatusID,
			StatusCode: item.DriverStatusCode,
			StatusName: item.DriverStatusName,
		},
		ShiftDate:       item.ShiftDate.Format("2006-01-02"),
		TimeFrom:        trimSeconds(item.TimeFrom),
		TimeTo:          trimTimePtr(item.TimeTo),
		Comment:         item.Comment,
		IsActive:        item.IsActive,
		TripsheetsCount: item.TripsheetsCount,
		Tripsheets:      tripsheets,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func mapDriverShiftTripsheetsToDTO(items []models.DriverShiftTripsheet) []dto.DriverShiftTripsheetBriefResponse {
	out := make([]dto.DriverShiftTripsheetBriefResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.DriverShiftTripsheetBriefResponse{
			ID:                 item.ID,
			TripsheetNumber:    item.TripsheetNumber,
			TripsheetDate:      item.TripsheetDate.Format("2006-01-02"),
			VehicleID:          item.VehicleID,
			VehicleBrand:       item.VehicleBrand,
			VehiclePlateNumber: item.VehiclePlateNumber,
			StartTime:          formatOptionalTimeRFC3339(item.StartTime),
			EndTime:            formatOptionalTimeRFC3339(item.EndTime),
			StatusID:           item.StatusID,
			StatusName:         item.StatusName,
			CreatedAt:          item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          item.UpdatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func formatOptionalTimeRFC3339(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}
