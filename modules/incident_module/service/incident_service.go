package service

import (
	"auto_park/modules/incident_module/dto"
	"auto_park/modules/incident_module/models"
	"auto_park/modules/incident_module/repository"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type IncidentService interface {
	Create(ctx context.Context, req dto.IncidentCreateRequest) (*dto.IncidentCreateResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.IncidentResponse, error)
	GetAll(ctx context.Context, filter dto.IncidentListQuery) ([]dto.IncidentResponse, int, error)
	Update(ctx context.Context, id int64, req dto.IncidentUpdateRequest) (*dto.IncidentResponse, error)
	Delete(ctx context.Context, id int64) error
	ListIncidentTypes(ctx context.Context, limit int, offset int) ([]dto.IncidentTypeResponse, int, int, int, error)
}

type incidentService struct {
	repo repository.IncidentRepository
}

func NewIncidentService(repo repository.IncidentRepository) IncidentService {
	return &incidentService{repo: repo}
}

func parseIncidentDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	parsed, err := time.Parse("2006-01-02", val)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be in format YYYY-MM-DD")
	}
	return parsed, nil
}

func parseIncidentTime(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return time.Time{}, fmt.Errorf("time is required")
	}
	layouts := []string{"15:04:05", "15:04"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, val); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("time must be in format HH:MM or HH:MM:SS")
}

func timeToStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func clockTimeToStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("15:04:05")
	return &s
}

func mapTripsheetToDTO(item *models.IncidentTripsheet) *dto.IncidentTripsheetResponse {
	if item == nil {
		return nil
	}

	return &dto.IncidentTripsheetResponse{
		ID:                         item.ID,
		TripsheetNumber:            item.TripsheetNumber,
		TripsheetDate:              item.TripsheetDate.Format("2006-01-02"),
		VehicleBrand:               item.VehicleBrand,
		VehiclePlateNumber:         item.VehiclePlateNumber,
		DriverID:                   item.DriverID,
		DriverLastName:             item.DriverLastName,
		DriverFirstName:            item.DriverFirstName,
		DriverMiddleName:           item.DriverMiddleName,
		StatusID:                   item.StatusID,
		StatusName:                 item.StatusName,
		StartTime:                  timeToStringPtr(item.StartTime),
		EndTime:                    timeToStringPtr(item.EndTime),
		MileageStart:               item.MileageStart,
		MileageEnd:                 item.MileageEnd,
		FuelStart:                  item.FuelStart,
		FuelIssued:                 item.FuelIssued,
		FuelConsumptionTheoretical: item.FuelConsumptionTheoretical,
		FuelConsumptionActual:      item.FuelConsumptionActual,
	}
}

func mapMechanicShiftToDTO(item *models.IncidentMechanicShift) *dto.IncidentMechanicShiftResponse {
	if item == nil {
		return nil
	}

	return &dto.IncidentMechanicShiftResponse{
		ID:               item.ID,
		UserID:           item.UserID,
		ShiftDate:        item.ShiftDate.Format("2006-01-02"),
		TimeFrom:         item.TimeFrom.Format("15:04:05"),
		TimeTo:           clockTimeToStringPtr(item.TimeTo),
		Comment:          item.Comment,
		IsActive:         item.IsActive,
		MechanicEmail:    item.MechanicEmail,
		MechanicFullName: item.MechanicFullName,
		CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        item.UpdatedAt.Format(time.RFC3339),
	}
}

func mapIncidentToDTO(item *models.Incident) *dto.IncidentResponse {
	if item == nil {
		return nil
	}

	return &dto.IncidentResponse{
		ID:                 item.ID,
		IncidentTypeID:     item.IncidentTypeID,
		IncidentTypeName:   item.IncidentTypeName,
		VehicleID:          item.VehicleID,
		VehicleStateNumber: item.VehicleStateNumber,
		DriverID:           item.DriverID,
		DriverFullName:     item.DriverFullName,
		MechanicID:         item.MechanicID,
		MechanicFullName:   item.MechanicFullName,
		MechanicShiftID:    item.MechanicShiftID,
		MechanicShift:      mapMechanicShiftToDTO(item.MechanicShift),
		TripsheetID:        item.TripsheetID,
		Tripsheet:          mapTripsheetToDTO(item.Tripsheet),
		Date:               item.IncidentDate.Format("2006-01-02"),
		Time:               item.IncidentTime.Format("15:04:05"),
		Location:           item.Location,
		Description:        item.Description,
		CreatedAt:          item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          item.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *incidentService) Create(ctx context.Context, req dto.IncidentCreateRequest) (*dto.IncidentCreateResponse, error) {
	dateVal, err := parseIncidentDate(req.Date)
	if err != nil {
		return nil, err
	}
	timeVal, err := parseIncidentTime(req.Time)
	if err != nil {
		return nil, err
	}

	if req.IncidentTypeID <= 0 || req.VehicleID <= 0 || req.DriverID <= 0 || req.MechanicID <= 0 || req.MechanicShiftID <= 0 {
		return nil, fmt.Errorf("incident_type_id, vehicle_id, driver_id, mechanic_id and mechanic_shift_id must be > 0")
	}
	if req.TripsheetID != nil && *req.TripsheetID <= 0 {
		return nil, fmt.Errorf("tripsheet_id must be > 0")
	}

	location := strings.TrimSpace(req.Location)
	if location == "" {
		return nil, fmt.Errorf("location is required")
	}

	item, err := s.repo.Create(ctx, models.CreateIncidentInput{
		IncidentTypeID:  req.IncidentTypeID,
		VehicleID:       req.VehicleID,
		DriverID:        req.DriverID,
		MechanicID:      req.MechanicID,
		MechanicShiftID: req.MechanicShiftID,
		TripsheetID:     req.TripsheetID,
		IncidentDate:    dateVal,
		IncidentTime:    timeVal,
		Location:        location,
		Description:     strings.TrimSpace(req.Description),
	})
	if err != nil {
		return nil, err
	}

	return &dto.IncidentCreateResponse{ID: item.ID}, nil
}

func (s *incidentService) GetByID(ctx context.Context, id int64) (*dto.IncidentResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapIncidentToDTO(item), nil
}

func (s *incidentService) GetAll(ctx context.Context, filter dto.IncidentListQuery) ([]dto.IncidentResponse, int, error) {
	if filter.DateFrom != "" {
		if _, err := parseIncidentDate(filter.DateFrom); err != nil {
			return nil, 0, err
		}
	}
	if filter.DateTo != "" {
		if _, err := parseIncidentDate(filter.DateTo); err != nil {
			return nil, 0, err
		}
	}
	if filter.TripsheetID != nil && *filter.TripsheetID <= 0 {
		return nil, 0, fmt.Errorf("tripsheet_id must be > 0")
	}
	if filter.MechanicShiftID != nil && *filter.MechanicShiftID <= 0 {
		return nil, 0, fmt.Errorf("mechanic_shift_id must be > 0")
	}

	items, total, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.IncidentResponse, 0, len(items))
	for i := range items {
		mapped := mapIncidentToDTO(&items[i])
		if mapped != nil {
			out = append(out, *mapped)
		}
	}

	return out, total, nil
}

func (s *incidentService) Update(ctx context.Context, id int64, req dto.IncidentUpdateRequest) (*dto.IncidentResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	dateVal, err := parseIncidentDate(req.Date)
	if err != nil {
		return nil, err
	}
	timeVal, err := parseIncidentTime(req.Time)
	if err != nil {
		return nil, err
	}

	if req.IncidentTypeID <= 0 || req.VehicleID <= 0 || req.DriverID <= 0 || req.MechanicID <= 0 || req.MechanicShiftID <= 0 {
		return nil, fmt.Errorf("incident_type_id, vehicle_id, driver_id, mechanic_id and mechanic_shift_id must be > 0")
	}
	if req.TripsheetID != nil && *req.TripsheetID <= 0 {
		return nil, fmt.Errorf("tripsheet_id must be > 0")
	}

	location := strings.TrimSpace(req.Location)
	if location == "" {
		return nil, fmt.Errorf("location is required")
	}

	item, err := s.repo.Update(ctx, models.UpdateIncidentInput{
		ID:              id,
		IncidentTypeID:  req.IncidentTypeID,
		VehicleID:       req.VehicleID,
		DriverID:        req.DriverID,
		MechanicID:      req.MechanicID,
		MechanicShiftID: req.MechanicShiftID,
		TripsheetID:     req.TripsheetID,
		IncidentDate:    dateVal,
		IncidentTime:    timeVal,
		Location:        location,
		Description:     strings.TrimSpace(req.Description),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return mapIncidentToDTO(item), nil
}

func (s *incidentService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *incidentService) ListIncidentTypes(ctx context.Context, limit int, offset int) ([]dto.IncidentTypeResponse, int, int, int, error) {
	items, err := s.repo.ListIncidentTypes(ctx)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	total := len(items)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	out := make([]dto.IncidentTypeResponse, 0, end-start)
	for i := range items[start:end] {
		out = append(out, dto.IncidentTypeResponse{
			ID:   items[start+i].ID,
			Name: items[start+i].Name,
		})
	}

	return out, total, limit, offset, nil
}
