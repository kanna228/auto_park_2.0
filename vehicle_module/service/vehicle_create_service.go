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
	ListVehicleStatuses(ctx context.Context) (*dto.VehicleStatusListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.VehicleUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	UploadPhoto(ctx context.Context, id int64, fileHeader any) (*dto.VehicleResponse, error)
	DeletePhoto(ctx context.Context, id int64) (*dto.VehicleResponse, error)
}

type vehicleService struct {
	repo                    repository.VehicleRepository
	storage                 *VehiclePhotoStorage
	insuranceRepo           repository.InsuranceRepository
	technicalInspectionRepo repository.TechnicalInspectionRepository
}

func NewVehicleService(
	repo repository.VehicleRepository,
	storage *VehiclePhotoStorage,
	insuranceRepo repository.InsuranceRepository,
	technicalInspectionRepo repository.TechnicalInspectionRepository,
) VehicleService {
	return &vehicleService{
		repo:                    repo,
		storage:                 storage,
		insuranceRepo:           insuranceRepo,
		technicalInspectionRepo: technicalInspectionRepo,
	}
}

func (s *vehicleService) Create(ctx context.Context, req dto.VehicleCreateRequest) (int64, error) {
	req.BoardNumber = strings.TrimSpace(req.BoardNumber)
	req.TechnicalPassportNumber = strings.TrimSpace(req.TechnicalPassportNumber)
	req.StateNumber = strings.TrimSpace(req.StateNumber)
	req.VIN = strings.TrimSpace(req.VIN)
	req.BrandModel = strings.TrimSpace(req.BrandModel)

	received := req.ReceivedDate.Format("2006-01-02")

	var expiry *string
	if req.InsuranceExpiryDate != nil {
		val := req.InsuranceExpiryDate.Format("2006-01-02")
		expiry = &val
	}

	if req.ManufactureYear < 1900 {
		return 0, fmt.Errorf("manufacture_year must be >= 1900")
	}
	if req.Mileage < 0 {
		return 0, fmt.Errorf("mileage must be >= 0")
	}
	if req.CurrentFuel < 0 {
		return 0, fmt.Errorf("current_fuel must be >= 0")
	}
	if req.StatusID <= 0 {
		return 0, fmt.Errorf("status_id must be > 0")
	}

	params := repository.CreateVehicleParams{
		BoardNumber:             req.BoardNumber,
		TechnicalPassportNumber: req.TechnicalPassportNumber,
		StateNumber:             req.StateNumber,
		VIN:                     req.VIN,
		BrandModel:              req.BrandModel,
		ManufactureYear:         req.ManufactureYear,
		ReceivedDate:            received,
		EmptyWeightKG:           req.EmptyWeightKG,
		MaxWeightKG:             req.MaxWeightKG,
		EngineVolumeCC:          req.EngineVolumeCC,
		InsurancePolicyNumber:   req.InsurancePolicyNumber,
		InsuranceExpiryDate:     expiry,
		Mileage:                 req.Mileage,
		CurrentFuel:             req.CurrentFuel,
		StatusID:                req.StatusID,
		DriversIDs:              req.DriversIDs,
	}

	_ = time.Now()

	return s.repo.Create(ctx, params)
}
