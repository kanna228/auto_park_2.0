package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/models"
	"auto_park/modules/vehicle_module/repository"
)

type VehicleDocumentService interface {
	List(ctx context.Context, vehicleID int64) ([]dto.VehicleDocumentResponse, error)
	Create(ctx context.Context, vehicleID int64, req dto.VehicleDocumentCreateRequest) (int64, error)
	Delete(ctx context.Context, vehicleID int64, documentID int64) (bool, error)
}

type vehicleDocumentService struct {
	repo repository.VehicleDocumentRepository
}

func NewVehicleDocumentService(repo repository.VehicleDocumentRepository) VehicleDocumentService {
	return &vehicleDocumentService{repo: repo}
}

func (s *vehicleDocumentService) List(ctx context.Context, vehicleID int64) ([]dto.VehicleDocumentResponse, error) {
	if vehicleID <= 0 {
		return nil, fmt.Errorf("invalid vehicle_id")
	}
	items, err := s.repo.ListByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.VehicleDocumentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapVehicleDocumentToDTO(item))
	}
	return out, nil
}

func (s *vehicleDocumentService) Create(ctx context.Context, vehicleID int64, req dto.VehicleDocumentCreateRequest) (int64, error) {
	if vehicleID <= 0 {
		return 0, fmt.Errorf("invalid vehicle_id")
	}
	docType := strings.ToLower(strings.TrimSpace(req.Type))
	if docType != "insurance" && docType != "tachograph" {
		return 0, fmt.Errorf("type must be insurance or tachograph")
	}
	number := strings.TrimSpace(req.Number)
	if number == "" {
		return 0, fmt.Errorf("number is required")
	}
	validFrom, err := normalizeVehicleDocumentDate(req.ValidFrom, "valid_from")
	if err != nil {
		return 0, err
	}
	validTo, err := normalizeVehicleDocumentDate(req.ValidTo, "valid_to")
	if err != nil {
		return 0, err
	}
	if validTo.Before(validFrom) {
		return 0, fmt.Errorf("valid_to cannot be earlier than valid_from")
	}

	return s.repo.Create(ctx, vehicleID, repository.CreateVehicleDocumentParams{
		Type:      docType,
		Number:    number,
		ValidFrom: validFrom.Format("2006-01-02"),
		ValidTo:   validTo.Format("2006-01-02"),
	})
}

func (s *vehicleDocumentService) Delete(ctx context.Context, vehicleID int64, documentID int64) (bool, error) {
	if vehicleID <= 0 {
		return false, fmt.Errorf("invalid vehicle_id")
	}
	if documentID <= 0 {
		return false, fmt.Errorf("invalid document id")
	}
	return s.repo.Delete(ctx, vehicleID, documentID)
}

func mapVehicleDocumentToDTO(item models.VehicleDocument) dto.VehicleDocumentResponse {
	return dto.VehicleDocumentResponse{
		ID:        item.ID,
		VehicleID: item.VehicleID,
		Type:      item.Type,
		Number:    item.Number,
		ValidFrom: item.ValidFrom.Format("2006-01-02"),
		ValidTo:   item.ValidTo.Format("2006-01-02"),
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
}

func normalizeVehicleDocumentDate(value string, field string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return parsed, nil
}
