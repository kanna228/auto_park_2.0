package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/vehicle_module/dto"
	"auto_park/vehicle_module/repository"
)

type VehicleService interface {
	Create(ctx context.Context, req dto.VehicleCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64) (*dto.VehicleResponse, error)
	List(ctx context.Context, q dto.VehicleListQuery) (*dto.VehicleListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.VehicleUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type vehicleService struct {
	repo repository.VehicleRepository
}

func NewVehicleService(repo repository.VehicleRepository) VehicleService {
	return &vehicleService{repo: repo}
}

func (s *vehicleService) Create(ctx context.Context, req dto.VehicleCreateRequest) (int64, error) {
	// Мини-санитайзинг строк
	req.BoardNumber = strings.TrimSpace(req.BoardNumber)
	req.TechnicalPassportNumber = strings.TrimSpace(req.TechnicalPassportNumber)
	req.StateNumber = strings.TrimSpace(req.StateNumber)
	req.VIN = strings.TrimSpace(req.VIN)
	req.BrandModel = strings.TrimSpace(req.BrandModel)

	// Даты в "YYYY-MM-DD" чтобы не поймать смещение по времени
	received := req.ReceivedDate.Format("2006-01-02")

	var expiry *string
	if req.InsuranceExpiryDate != nil {
		s := req.InsuranceExpiryDate.Format("2006-01-02")
		expiry = &s
	}

	// если кто-то не передал mileage/current_fuel — оставим как есть (у тебя default в БД тоже есть, но мы принимаем всё)
	if req.Mileage < 0 {
		return 0, fmt.Errorf("mileage must be >= 0")
	}
	if req.CurrentFuel < 0 {
		return 0, fmt.Errorf("current_fuel must be >= 0")
	}

	params := repository.CreateVehicleParams{
		BoardNumber:             req.BoardNumber,
		TechnicalPassportNumber: req.TechnicalPassportNumber,
		StateNumber:             req.StateNumber,
		VIN:                     req.VIN,
		BrandModel:              req.BrandModel,
		ManufactureYear:         req.ManufactureYear,
		ReceivedDate:            received,

		EmptyWeightKG:  req.EmptyWeightKG,
		MaxWeightKG:    req.MaxWeightKG,
		EngineVolumeCC: req.EngineVolumeCC,

		InsurancePolicyNumber: req.InsurancePolicyNumber,
		InsuranceExpiryDate:   expiry,

		Mileage:     req.Mileage,
		CurrentFuel: req.CurrentFuel,

		DriversIDs: req.DriversIDs,
	}

	// На всякий случай: если в запросе date-time — отрежем время
	_ = time.Now()

	return s.repo.Create(ctx, params)
}
