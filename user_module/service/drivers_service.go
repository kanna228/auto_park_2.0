package service

import (
	"context"

	"auto_park/user_module/dto"
	"auto_park/user_module/models"
	"auto_park/user_module/repository"
)

type DriversService struct {
	repo *repository.DriverRepo
}

func NewDriversService(repo *repository.DriverRepo) *DriversService {
	return &DriversService{repo: repo}
}

func (s *DriversService) Create(ctx context.Context, req dto.CreateDriverRequest) (*models.Driver, error) {
	return s.repo.Create(ctx, &models.Driver{
		IIN:        req.IIN,
		Name:       req.Name,
		Surname:    req.Surname,
		Middlename: req.Middlename,
		Phone:      req.Phone,
		Mail:       req.Mail,
	})
}

func (s *DriversService) GetByID(ctx context.Context, id int64) (*models.Driver, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DriversService) List(ctx context.Context, limit, offset int) ([]models.Driver, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *DriversService) Update(ctx context.Context, id int64, req dto.UpdateDriverRequest) (*models.Driver, error) {
	upd := map[string]any{}
	if req.IIN != nil {
		upd["iin"] = *req.IIN
	}
	if req.Name != nil {
		upd["name"] = *req.Name
	}
	if req.Surname != nil {
		upd["surname"] = *req.Surname
	}
	if req.Middlename != nil {
		upd["middlename"] = *req.Middlename
	}
	if req.Phone != nil {
		upd["phone"] = *req.Phone
	}
	if req.Mail != nil {
		upd["mail"] = *req.Mail
	}
	return s.repo.Update(ctx, id, upd)
}

func (s *DriversService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
