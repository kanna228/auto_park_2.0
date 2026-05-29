package service

import (
	"context"
	"fmt"
	"strings"

	"auto_park/internal/apierrors"
	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/repository"
)

func (s *vehicleService) UpdateByID(ctx context.Context, id int64, req dto.VehicleUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	current, err := s.repo.GetByID(ctx, id, true)
	if err != nil {
		return false, err
	}
	if current == nil {
		return false, nil
	}
	if current.IsArchived {
		return false, apierrors.ErrEntityArchived
	}

	req.BoardNumber = strings.TrimSpace(req.BoardNumber)
	req.TechnicalPassportNumber = strings.TrimSpace(req.TechnicalPassportNumber)
	req.StateNumber = strings.TrimSpace(req.StateNumber)
	req.VIN = strings.TrimSpace(req.VIN)
	req.BrandModel = strings.TrimSpace(req.BrandModel)

	if req.ReceivedDate.IsZero() {
		return false, fmt.Errorf("received_date is required")
	}
	if req.ManufactureYear < 1900 {
		return false, fmt.Errorf("manufacture_year must be >= 1900")
	}
	if req.Mileage < 0 {
		return false, fmt.Errorf("mileage must be >= 0")
	}
	if req.CurrentFuel < 0 {
		return false, fmt.Errorf("current_fuel must be >= 0")
	}
	if req.StatusID <= 0 {
		return false, fmt.Errorf("status_id must be > 0")
	}

	received := req.ReceivedDate.Format("2006-01-02")

	var expiry *string
	if req.InsuranceExpiryDate != nil {
		tmp := req.InsuranceExpiryDate.Format("2006-01-02")
		expiry = &tmp
	}

	params := repository.UpdateVehicleParams{
		BoardNumber:             req.BoardNumber,
		TechnicalPassportNumber: req.TechnicalPassportNumber,
		StateNumber:             req.StateNumber,
		VIN:                     req.VIN,

		BrandModel:      req.BrandModel,
		ManufactureYear: req.ManufactureYear,
		ReceivedDate:    received,

		EmptyWeightKG:  req.EmptyWeightKG,
		MaxWeightKG:    req.MaxWeightKG,
		EngineVolumeCC: req.EngineVolumeCC,

		InsurancePolicyNumber: req.InsurancePolicyNumber,
		InsuranceExpiryDate:   expiry,

		Mileage:     req.Mileage,
		CurrentFuel: req.CurrentFuel,

		StatusID: req.StatusID,

		DriversIDs: req.DriversIDs,
	}

	return s.repo.UpdateByID(ctx, id, params)
}
