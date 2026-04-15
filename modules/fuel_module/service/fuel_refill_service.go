package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/fuel_module/dto"
	"auto_park/modules/fuel_module/models"
	"auto_park/modules/fuel_module/repository"
)

type FuelRefillService interface {
	Create(ctx context.Context, req dto.CreateFuelRefillRequest) (*dto.FuelRefillResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateFuelRefillRequest) (*dto.FuelRefillResponse, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dto.FuelRefillResponse, error)
	GetAll(ctx context.Context, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error)
	GetAllByTripsheetID(ctx context.Context, tripsheetID int64) ([]dto.FuelRefillResponse, int, error)
	GetAllByVehicleID(ctx context.Context, vehicleID int64) ([]dto.FuelRefillResponse, int, error)
}

type fuelRefillService struct {
	repo repository.FuelRefillRepository
}

func NewFuelRefillService(repo repository.FuelRefillRepository) FuelRefillService {
	return &fuelRefillService{repo: repo}
}

func (s *fuelRefillService) Create(ctx context.Context, req dto.CreateFuelRefillRequest) (*dto.FuelRefillResponse, error) {
	input, err := validateAndMapCreate(req)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return mapModelToResponse(created), nil
}

func (s *fuelRefillService) Update(ctx context.Context, id int64, req dto.UpdateFuelRefillRequest) (*dto.FuelRefillResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	input, err := validateAndMapUpdate(id, req)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	return mapModelToResponse(updated), nil
}

func (s *fuelRefillService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *fuelRefillService) GetByID(ctx context.Context, id int64) (*dto.FuelRefillResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *fuelRefillService) GetAll(ctx context.Context, filter dto.FuelRefillFilter) ([]dto.FuelRefillResponse, int, error) {
	return s.repo.GetAll(ctx, filter)
}

func (s *fuelRefillService) GetAllByTripsheetID(ctx context.Context, tripsheetID int64) ([]dto.FuelRefillResponse, int, error) {
	if tripsheetID <= 0 {
		return nil, 0, fmt.Errorf("invalid tripsheet_id")
	}
	return s.repo.GetAllByTripsheetID(ctx, tripsheetID)
}

func (s *fuelRefillService) GetAllByVehicleID(ctx context.Context, vehicleID int64) ([]dto.FuelRefillResponse, int, error) {
	if vehicleID <= 0 {
		return nil, 0, fmt.Errorf("invalid vehicle_id")
	}
	return s.repo.GetAllByVehicleID(ctx, vehicleID)
}

func validateAndMapCreate(req dto.CreateFuelRefillRequest) (models.CreateFuelRefillInput, error) {
	dateVal, timeVal, location, err := validateCommon(req.TripsheetID, req.VehicleID, req.FuelAmount, req.Date, req.Time, req.Location)
	if err != nil {
		return models.CreateFuelRefillInput{}, err
	}

	return models.CreateFuelRefillInput{
		TripsheetID: req.TripsheetID,
		VehicleID:   req.VehicleID,
		FuelAmount:  req.FuelAmount,
		Date:        dateVal,
		Time:        timeVal,
		Location:    location,
	}, nil
}

func validateAndMapUpdate(id int64, req dto.UpdateFuelRefillRequest) (models.UpdateFuelRefillInput, error) {
	dateVal, timeVal, location, err := validateCommon(req.TripsheetID, req.VehicleID, req.FuelAmount, req.Date, req.Time, req.Location)
	if err != nil {
		return models.UpdateFuelRefillInput{}, err
	}

	return models.UpdateFuelRefillInput{
		ID:          id,
		TripsheetID: req.TripsheetID,
		VehicleID:   req.VehicleID,
		FuelAmount:  req.FuelAmount,
		Date:        dateVal,
		Time:        timeVal,
		Location:    location,
	}, nil
}

func validateCommon(tripsheetID, vehicleID int64, fuelAmount float64, dateStr, timeStr string, location *string) (time.Time, time.Time, *string, error) {
	if tripsheetID <= 0 {
		return time.Time{}, time.Time{}, nil, fmt.Errorf("tripsheet_id is required")
	}
	if vehicleID <= 0 {
		return time.Time{}, time.Time{}, nil, fmt.Errorf("vehicle_id is required")
	}
	if fuelAmount <= 0 {
		return time.Time{}, time.Time{}, nil, fmt.Errorf("fuel_amount must be greater than 0")
	}

	dateVal, err := time.Parse("2006-01-02", strings.TrimSpace(dateStr))
	if err != nil {
		return time.Time{}, time.Time{}, nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	parsedTime, err := time.Parse("15:04:05", strings.TrimSpace(timeStr))
	if err != nil {
		return time.Time{}, time.Time{}, nil, fmt.Errorf("invalid time format, expected HH:MM:SS")
	}

	var normalizedLocation *string
	if location != nil {
		trimmed := strings.TrimSpace(*location)
		if trimmed != "" {
			normalizedLocation = &trimmed
		}
	}

	return dateVal, parsedTime, normalizedLocation, nil
}

func mapModelToResponse(item *models.FuelRefill) *dto.FuelRefillResponse {
	resp := &dto.FuelRefillResponse{
		ID:          item.ID,
		TripsheetID: item.TripsheetID,
		VehicleID:   item.VehicleID,
		FuelAmount:  item.FuelAmount,
		Date:        item.Date.Format("2006-01-02"),
		Time:        item.Time.Format("15:04:05"),
		Location:    item.Location,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
	return resp
}
