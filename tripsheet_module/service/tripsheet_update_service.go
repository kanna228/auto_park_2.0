package service

import (
	"auto_park/tripsheet_module/dto"
	"auto_park/tripsheet_module/models"
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *tripsheetService) Update(ctx context.Context, id int64, req dto.UpdateTripsheetRequest) (*dto.CreateTripsheetResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

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

	statusID := req.StatusID
	if statusID == nil {
		current, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		statusID = &current.StatusID
	}

	input := models.UpdateTripsheetInput{
		ID:                         id,
		TripsheetNumber:            req.TripsheetNumber,
		TripsheetDate:              tripsheetDate,
		VehicleID:                  req.VehicleID,
		VehicleBrand:               req.VehicleBrand,
		VehiclePlateNumber:         req.VehiclePlateNumber,
		DriverLastName:             req.DriverLastName,
		DriverFirstName:            req.DriverFirstName,
		DriverMiddleName:           req.DriverMiddleName,
		DriverID:                   req.DriverID,
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

	updated, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	return mapTripsheetModelToCreateResponse(updated), nil
}

func mapTripsheetModelToCreateResponse(item *models.Tripsheet) *dto.CreateTripsheetResponse {
	resp := &dto.CreateTripsheetResponse{
		ID:                         item.ID,
		TripsheetNumber:            item.TripsheetNumber,
		TripsheetDate:              item.TripsheetDate.Format("2006-01-02"),
		VehicleID:                  item.VehicleID,
		VehicleBrand:               item.VehicleBrand,
		VehiclePlateNumber:         item.VehiclePlateNumber,
		DriverLastName:             item.DriverLastName,
		DriverFirstName:            item.DriverFirstName,
		DriverMiddleName:           item.DriverMiddleName,
		DriverID:                   item.DriverID,
		MileageStart:               item.MileageStart,
		MileageEnd:                 item.MileageEnd,
		FuelStart:                  item.FuelStart,
		FuelIssued:                 item.FuelIssued,
		FuelConsumptionTheoretical: item.FuelConsumptionTheoretical,
		FuelConsumptionActual:      item.FuelConsumptionActual,
		StatusID:                   item.StatusID,
		CreatedAt:                  item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                  item.UpdatedAt.Format(time.RFC3339),
	}

	if item.StartTime != nil {
		v := item.StartTime.Format(time.RFC3339)
		resp.StartTime = &v
	}
	if item.EndTime != nil {
		v := item.EndTime.Format(time.RFC3339)
		resp.EndTime = &v
	}

	return resp
}
