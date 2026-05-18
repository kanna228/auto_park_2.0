package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/models"
	"auto_park/modules/tripsheet_module/repository"
)

type TripsheetService interface {
	Create(ctx context.Context, req dto.CreateTripsheetRequest) (*dto.CreateTripsheetResponse, error)
	GetAll(ctx context.Context, f dto.TripsheetFilter) ([]dto.TripsheetResponse, int, error)
	GetByID(ctx context.Context, id int64) (*dto.TripsheetResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateTripsheetRequest) (*dto.CreateTripsheetResponse, error)
	Delete(ctx context.Context, id int64) error
}

type tripsheetService struct {
	repo repository.TripsheetRepo
}

func NewTripsheetService(repo repository.TripsheetRepo) TripsheetService {
	return &tripsheetService{
		repo: repo,
	}
}

func (s *tripsheetService) Create(ctx context.Context, req dto.CreateTripsheetRequest) (*dto.CreateTripsheetResponse, error) {
	req.TripsheetNumber = strings.TrimSpace(req.TripsheetNumber)
	req.TripsheetDate = strings.TrimSpace(req.TripsheetDate)
	req.VehiclePlateNumber = strings.TrimSpace(req.VehiclePlateNumber)

	if req.TripsheetNumber == "" {
		return nil, fmt.Errorf("tripsheet_number is required")
	}
	if req.TripsheetDate == "" {
		return nil, fmt.Errorf("tripsheet_date is required")
	}
	if req.VehiclePlateNumber == "" {
		return nil, fmt.Errorf("vehicle_plate_number is required")
	}

	tripsheetDate, err := time.Parse("2006-01-02", req.TripsheetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid tripsheet_date format, expected YYYY-MM-DD")
	}

	var startTime *time.Time
	if req.StartTime != nil && strings.TrimSpace(*req.StartTime) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.StartTime))
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format, expected RFC3339")
		}
		startTime = &parsed
	}

	var endTime *time.Time
	if req.EndTime != nil && strings.TrimSpace(*req.EndTime) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.EndTime))
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format, expected RFC3339")
		}
		endTime = &parsed
	}

	mileageStart := getIntOrZero(req.MileageStart)
	mileageEnd := getIntOrZero(req.MileageEnd)
	fuelStart := getIntOrZero(req.FuelStart)
	fuelIssued := getIntOrZero(req.FuelIssued)
	fuelTheo := getIntOrZero(req.FuelConsumptionTheoretical)
	fuelActual := getIntOrZero(req.FuelConsumptionActual)

	if mileageStart < 0 || mileageEnd < 0 || fuelStart < 0 || fuelIssued < 0 || fuelTheo < 0 || fuelActual < 0 {
		return nil, fmt.Errorf("numeric fields cannot be negative")
	}

	if mileageEnd < mileageStart {
		return nil, fmt.Errorf("mileage_end cannot be less than mileage_start")
	}

	if req.DriverShiftID != nil && *req.DriverShiftID > 0 {
		if err := s.repo.ValidateDriverShiftForTripsheet(ctx, *req.DriverShiftID, req.DriverID, tripsheetDate); err != nil {
			return nil, err
		}
	}

	statusID := req.StatusID
	if statusID == nil {
		defaultStatusID, err := s.repo.GetStatusIDByName(ctx, "Создано")
		if err != nil {
			return nil, fmt.Errorf("get default status: %w", err)
		}
		statusID = &defaultStatusID
	}

	input := models.CreateTripsheetInput{
		TripsheetNumber:            req.TripsheetNumber,
		TripsheetDate:              tripsheetDate,
		VehicleID:                  req.VehicleID,
		VehicleBrand:               req.VehicleBrand,
		VehiclePlateNumber:         req.VehiclePlateNumber,
		DriverLastName:             req.DriverLastName,
		DriverFirstName:            req.DriverFirstName,
		DriverMiddleName:           req.DriverMiddleName,
		DriverID:                   req.DriverID,
		DriverShiftID:              req.DriverShiftID,
		StartTime:                  startTime,
		EndTime:                    endTime,
		MileageStart:               mileageStart,
		MileageEnd:                 mileageEnd,
		FuelStart:                  fuelStart,
		FuelIssued:                 fuelIssued,
		FuelConsumptionTheoretical: fuelTheo,
		FuelConsumptionActual:      fuelActual,
		StatusID:                   statusID,
	}

	created, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := s.syncDriverStatusForTripsheet(ctx, created.DriverID, created.DriverShiftID, created.EndTime, created.StatusID); err != nil {
		return nil, err
	}

	resp := &dto.CreateTripsheetResponse{
		ID:                         created.ID,
		TripsheetNumber:            created.TripsheetNumber,
		TripsheetDate:              created.TripsheetDate.Format("2006-01-02"),
		VehicleID:                  created.VehicleID,
		VehicleBrand:               created.VehicleBrand,
		VehiclePlateNumber:         created.VehiclePlateNumber,
		DriverLastName:             created.DriverLastName,
		DriverFirstName:            created.DriverFirstName,
		DriverMiddleName:           created.DriverMiddleName,
		DriverID:                   created.DriverID,
		DriverShiftID:              created.DriverShiftID,
		MileageStart:               created.MileageStart,
		MileageEnd:                 created.MileageEnd,
		FuelStart:                  created.FuelStart,
		FuelIssued:                 created.FuelIssued,
		FuelConsumptionTheoretical: created.FuelConsumptionTheoretical,
		FuelConsumptionActual:      created.FuelConsumptionActual,
		StatusID:                   created.StatusID,
		CreatedAt:                  created.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                  created.UpdatedAt.Format(time.RFC3339),
	}

	if created.StartTime != nil {
		v := created.StartTime.Format(time.RFC3339)
		resp.StartTime = &v
	}
	if created.EndTime != nil {
		v := created.EndTime.Format(time.RFC3339)
		resp.EndTime = &v
	}

	return resp, nil
}

func (s *tripsheetService) syncDriverStatusForTripsheet(ctx context.Context, driverID *int64, driverShiftID *int64, endTime *time.Time, statusID int64) error {
	resolvedDriverID, err := s.repo.ResolveDriverID(ctx, driverID, driverShiftID)
	if err != nil {
		return err
	}
	if resolvedDriverID == nil || *resolvedDriverID <= 0 {
		return nil
	}
	if endTime == nil && !isClosedTripsheetStatus(statusID) {
		return s.repo.SetDriverStatusByCode(ctx, *resolvedDriverID, "on_trip")
	}
	return s.repo.RefreshDriverAvailability(ctx, *resolvedDriverID)
}

func isClosedTripsheetStatus(statusID int64) bool {
	return statusID == 4 || statusID == 5
}

func getIntOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
