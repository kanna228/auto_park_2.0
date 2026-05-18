package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/models"
	"auto_park/modules/vehicle_module/repository"
)

func (s *tireService) ListVehicleTires(ctx context.Context, vehicleID int64) ([]dto.VehicleTireResponse, error) {
	if vehicleID <= 0 {
		return nil, fmt.Errorf("invalid vehicle_id")
	}

	items, err := s.repo.ListByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	out := make([]dto.VehicleTireResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapVehicleTireToDTO(item))
	}
	return out, nil
}

func (s *tireService) CreateVehicleTire(ctx context.Context, vehicleID int64, req dto.VehicleTireCreateRequest) (int64, error) {
	if vehicleID <= 0 {
		return 0, fmt.Errorf("invalid vehicle_id")
	}
	position := strings.TrimSpace(req.Position)
	tireType := strings.TrimSpace(req.Type)
	if position == "" {
		return 0, fmt.Errorf("position is required")
	}
	if tireType == "" {
		return 0, fmt.Errorf("type is required")
	}
	if req.MileageKM < 0 {
		return 0, fmt.Errorf("mileage_km must be >= 0")
	}
	if req.MaxMileageKM <= 0 {
		return 0, fmt.Errorf("max_mileage_km must be greater than 0")
	}

	installedAt, err := normalizeVehicleTireDate(req.InstalledAt)
	if err != nil {
		return 0, err
	}
	placeID, err := s.repo.EnsurePlace(ctx, position)
	if err != nil {
		return 0, err
	}

	return s.repo.CreateForVehicle(ctx, vehicleID, repository.CreateVehicleTireParams{
		PlaceID:     placeID,
		Tire:        tireType,
		Mileage:     req.MileageKM,
		MaxUsage:    req.MaxMileageKM,
		InstalledAt: installedAt,
	})
}

func (s *tireService) UpdateVehicleTire(ctx context.Context, vehicleID int64, tireID int64, req dto.VehicleTireUpdateRequest) (bool, error) {
	if vehicleID <= 0 {
		return false, fmt.Errorf("invalid vehicle_id")
	}
	if tireID <= 0 {
		return false, fmt.Errorf("invalid tire id")
	}

	current, err := s.repo.GetByID(ctx, tireID)
	if err != nil {
		return false, err
	}
	if current == nil || current.VehicleID == nil || *current.VehicleID != vehicleID {
		return false, nil
	}

	position := strings.TrimSpace(current.PlaceName)
	if req.Position != nil {
		position = strings.TrimSpace(*req.Position)
	}
	if position == "" {
		return false, fmt.Errorf("position is required")
	}

	tireType := strings.TrimSpace(current.Tire)
	if req.Type != nil {
		tireType = strings.TrimSpace(*req.Type)
	}
	if tireType == "" {
		return false, fmt.Errorf("type is required")
	}

	mileage := current.Mileage
	if req.MileageKM != nil {
		mileage = *req.MileageKM
	}
	if mileage < 0 {
		return false, fmt.Errorf("mileage_km must be >= 0")
	}

	maxMileage := current.MaxUsage
	if req.MaxMileageKM != nil {
		maxMileage = *req.MaxMileageKM
	}
	if maxMileage <= 0 {
		return false, fmt.Errorf("max_mileage_km must be greater than 0")
	}

	installedAt := current.InstalledAt.Format("2006-01-02")
	if req.InstalledAt != nil {
		parsed, err := normalizeVehicleTireDate(*req.InstalledAt)
		if err != nil {
			return false, err
		}
		installedAt = parsed
	}

	placeID, err := s.repo.EnsurePlace(ctx, position)
	if err != nil {
		return false, err
	}

	return s.repo.UpdateForVehicle(ctx, vehicleID, tireID, repository.UpdateVehicleTireParams{
		PlaceID:     placeID,
		Tire:        tireType,
		Mileage:     mileage,
		MaxUsage:    maxMileage,
		InstalledAt: installedAt,
	})
}

func (s *tireService) DetachVehicleTire(ctx context.Context, vehicleID int64, tireID int64) (bool, error) {
	if vehicleID <= 0 {
		return false, fmt.Errorf("invalid vehicle_id")
	}
	if tireID <= 0 {
		return false, fmt.Errorf("invalid tire id")
	}
	return s.repo.DetachFromVehicle(ctx, vehicleID, tireID)
}

func mapVehicleTireToDTO(item models.Tire) dto.VehicleTireResponse {
	vehicleID := int64(0)
	if item.VehicleID != nil {
		vehicleID = *item.VehicleID
	}

	return dto.VehicleTireResponse{
		ID:               item.ID,
		VehicleID:        vehicleID,
		Position:         item.PlaceName,
		Type:             item.Tire,
		MileageKM:        item.Mileage,
		MaxMileageKM:     item.MaxUsage,
		RemainingPercent: tireRemainingPercent(item.Mileage, item.MaxUsage),
		InstalledAt:      item.InstalledAt.Format("2006-01-02"),
	}
}

func tireRemainingPercent(mileage int64, maxMileage int64) int64 {
	if maxMileage <= 0 {
		return 0
	}
	remaining := float64(maxMileage-mileage) / float64(maxMileage) * 100
	if remaining < 0 {
		return 0
	}
	if remaining > 100 {
		return 100
	}
	return int64(math.Round(remaining))
}

func normalizeVehicleTireDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now().Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return "", fmt.Errorf("installed_at must be in YYYY-MM-DD format")
	}
	return value, nil
}
