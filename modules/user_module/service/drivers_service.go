package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/models"
	"auto_park/modules/user_module/repository"
)

type DriversService struct {
	repo    *repository.DriverRepo
	storage *DriverPhotoStorage
}

func NewDriversService(repo *repository.DriverRepo, storage *DriverPhotoStorage) *DriversService {
	return &DriversService{
		repo:    repo,
		storage: storage,
	}
}

func (s *DriversService) Create(ctx context.Context, req dto.CreateDriverRequest) (*models.Driver, error) {
	birthDate, err := parseOptionalDriverDate(req.BirthDate, "birth_date")
	if err != nil {
		return nil, err
	}
	if err := validateExperienceYears(req.ExperienceYears); err != nil {
		return nil, err
	}
	statusID := req.StatusID
	if statusID <= 0 {
		statusID = 1
	}
	if err := s.validateDriverStatus(ctx, statusID); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &models.Driver{
		IIN:             req.IIN,
		Name:            req.Name,
		Surname:         req.Surname,
		Middlename:      req.Middlename,
		Phone:           req.Phone,
		Mail:            req.Mail,
		BirthDate:       birthDate,
		LicenseNumber:   strings.TrimSpace(req.LicenseNumber),
		LicenseCategory: strings.TrimSpace(req.LicenseCategory),
		ExperienceYears: req.ExperienceYears,
		StatusID:        statusID,
	})
}

func (s *DriversService) GetByID(ctx context.Context, id int64) (*models.Driver, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DriversService) GetPassport(ctx context.Context, id int64, q dto.DriverPassportQuery) (*models.DriverPassport, error) {
	return s.repo.GetPassport(ctx, id, repository.DriverPassportParams{
		TripsLimit:      q.TripsLimit,
		TripsOffset:     q.TripsOffset,
		IncidentsLimit:  q.IncidentsLimit,
		IncidentsOffset: q.IncidentsOffset,
	})
}

func (s *DriversService) List(ctx context.Context, q dto.DriverListQuery) ([]models.Driver, int64, error) {
	return s.repo.List(ctx, repository.DriverListParams{
		Search:      strings.TrimSpace(q.Search),
		Status:      strings.TrimSpace(q.Status),
		BoardNumber: strings.TrimSpace(q.BoardNumber),
		Limit:       q.Limit,
		Offset:      q.Offset,
	})
}

func (s *DriversService) ListStatuses(ctx context.Context, limit, offset int) ([]models.DriverStatus, int64, error) {
	return s.repo.ListStatuses(ctx, limit, offset)
}

func (s *DriversService) UpdateStatus(ctx context.Context, id int64, req dto.UpdateDriverStatusRequest) (*models.Driver, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if err := s.validateDriverStatus(ctx, req.StatusID); err != nil {
		return nil, err
	}
	return s.repo.UpdateStatus(ctx, id, req.StatusID)
}

func (s *DriversService) Update(ctx context.Context, id int64, req dto.UpdateDriverRequest) (*models.Driver, error) {
	upd := map[string]any{}
	if req.IIN != nil {
		upd["iin"] = strings.TrimSpace(*req.IIN)
	}
	if req.Name != nil {
		upd["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Surname != nil {
		upd["surname"] = strings.TrimSpace(*req.Surname)
	}
	if req.Middlename != nil {
		upd["middlename"] = nullText(strings.TrimSpace(*req.Middlename))
	}
	if req.Phone != nil {
		upd["phone"] = nullText(strings.TrimSpace(*req.Phone))
	}
	if req.Mail != nil {
		upd["mail"] = nullText(strings.TrimSpace(*req.Mail))
	}
	if req.BirthDate != nil {
		birthDate, err := parseOptionalDriverDate(*req.BirthDate, "birth_date")
		if err != nil {
			return nil, err
		}
		if birthDate == nil {
			upd["birth_date"] = nil
		} else {
			upd["birth_date"] = *birthDate
		}
	}
	if req.LicenseNumber != nil {
		upd["license_number"] = nullText(strings.TrimSpace(*req.LicenseNumber))
	}
	if req.LicenseCategory != nil {
		upd["license_category"] = nullText(strings.TrimSpace(*req.LicenseCategory))
	}
	if req.ExperienceYears != nil {
		if err := validateExperienceYears(req.ExperienceYears); err != nil {
			return nil, err
		}
		upd["experience_years"] = *req.ExperienceYears
	}
	if req.StatusID != nil {
		if err := s.validateDriverStatus(ctx, *req.StatusID); err != nil {
			return nil, err
		}
		upd["status_id"] = *req.StatusID
	}
	return s.repo.Update(ctx, id, upd)
}

func (s *DriversService) UploadPhoto(ctx context.Context, id int64, file *multipart.FileHeader) (*models.Driver, error) {
	driver, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	relativePath, err := s.storage.Save(id, file, driver.PhotoPath)
	if err != nil {
		return nil, err
	}

	return s.repo.UpdatePhotoPath(ctx, id, relativePath)
}

func (s *DriversService) DeletePhoto(ctx context.Context, id int64) (*models.Driver, error) {
	driver, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if driver.PhotoPath != "" {
		if err := s.storage.Delete(driver.PhotoPath); err != nil {
			return nil, err
		}
	}

	return s.repo.UpdatePhotoPath(ctx, id, "")
}

func (s *DriversService) Delete(ctx context.Context, id int64) error {
	driver, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if driver.PhotoPath != "" {
		_ = s.storage.Delete(driver.PhotoPath)
	}

	return nil
}

func (s *DriversService) AssignVehicle(ctx context.Context, driverID int64, vehicleID int64) error {
	if driverID <= 0 {
		return fmt.Errorf("invalid driver id")
	}
	if vehicleID <= 0 {
		return fmt.Errorf("invalid vehicle_id")
	}
	return s.repo.AssignVehicle(ctx, driverID, vehicleID)
}

func (s *DriversService) UnassignVehicle(ctx context.Context, driverID int64, vehicleID int64) error {
	if driverID <= 0 {
		return fmt.Errorf("invalid driver id")
	}
	if vehicleID <= 0 {
		return fmt.Errorf("invalid vehicle_id")
	}
	return s.repo.UnassignVehicle(ctx, driverID, vehicleID)
}

func parseOptionalDriverDate(value string, field string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}

	return &parsed, nil
}

func validateExperienceYears(value *int) error {
	if value == nil {
		return nil
	}
	if *value < 0 {
		return fmt.Errorf("experience_years cannot be negative")
	}
	if *value > 80 {
		return fmt.Errorf("experience_years must be less than or equal to 80")
	}
	return nil
}

func (s *DriversService) validateDriverStatus(ctx context.Context, statusID int64) error {
	if statusID <= 0 {
		return fmt.Errorf("status_id is required")
	}
	exists, err := s.repo.StatusExists(ctx, statusID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("driver status not found")
	}
	return nil
}

func nullText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
