package service

import (
	"auto_park/vehicle_module/dto"
	"auto_park/vehicle_module/models"
	"auto_park/vehicle_module/repository"
	"context"
)

func (s *vehicleService) GetByID(ctx context.Context, id int64) (*dto.VehicleResponse, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	resp, err := s.buildVehicleResponse(ctx, *v)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *vehicleService) List(ctx context.Context, q dto.VehicleListQuery) (*dto.VehicleListResponse, error) {
	params := repository.ListVehiclesParams{
		BoardNumber: q.BoardNumber,
		StateNumber: q.StateNumber,
		VIN:         q.VIN,
		BrandModel:  q.BrandModel,

		StatusID: q.StatusID,

		ManufactureYearFrom: q.ManufactureYearFrom,
		ManufactureYearTo:   q.ManufactureYearTo,

		DriverID: q.DriverID,

		Limit:  q.Limit,
		Offset: q.Offset,

		SortBy: q.SortBy,
		Order:  q.Order,
	}

	items, total, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	out := make([]dto.VehicleResponse, 0, len(items))
	for _, v := range items {
		resp, err := s.buildVehicleResponse(ctx, v)
		if err != nil {
			return nil, err
		}
		out = append(out, resp)
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.VehicleListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *vehicleService) buildVehicleResponse(ctx context.Context, v models.Vehicle) (dto.VehicleResponse, error) {
	resp := mapVehicleToDTO(v)

	insurances, _, err := s.insuranceRepo.List(ctx, repository.ListInsuranceParams{
		VehicleID: &v.ID,
		Limit:     1000,
		Offset:    0,
		SortBy:    "start_date",
		Order:     "desc",
	})
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	technicalInspections, _, err := s.technicalInspectionRepo.List(ctx, repository.ListTechnicalInspectionParams{
		VehicleID: &v.ID,
		Limit:     1000,
		Offset:    0,
		SortBy:    "start_date",
		Order:     "desc",
	})
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	resp.Insurances = make([]dto.VehicleInsuranceHistoryItem, 0, len(insurances))
	for _, item := range insurances {
		resp.Insurances = append(resp.Insurances, dto.VehicleInsuranceHistoryItem{
			ID:        item.ID,
			VehicleID: item.VehicleID,
			Name:      item.Name,
			StartDate: item.StartDate,
			EndDate:   item.EndDate,
			FilePath:  item.FilePath,
			FileURL:   "",
			IsActive:  item.IsActive,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	resp.TechnicalInspections = make([]dto.VehicleTechnicalInspectionHistoryItem, 0, len(technicalInspections))
	for _, item := range technicalInspections {
		resp.TechnicalInspections = append(resp.TechnicalInspections, dto.VehicleTechnicalInspectionHistoryItem{
			ID:        item.ID,
			VehicleID: item.VehicleID,
			Name:      item.Name,
			StartDate: item.StartDate,
			EndDate:   item.EndDate,
			FilePath:  item.FilePath,
			FileURL:   "",
			IsActive:  item.IsActive,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	return resp, nil
}

func mapVehicleToDTO(v models.Vehicle) dto.VehicleResponse {
	return dto.VehicleResponse{
		ID: v.ID,

		BoardNumber:             v.BoardNumber,
		TechnicalPassportNumber: v.TechnicalPassportNumber,
		StateNumber:             v.StateNumber,
		VIN:                     v.VIN,

		BrandModel:      v.BrandModel,
		ManufactureYear: v.ManufactureYear,
		ReceivedDate:    v.ReceivedDate,

		EmptyWeightKG:  v.EmptyWeightKG,
		MaxWeightKG:    v.MaxWeightKG,
		EngineVolumeCC: v.EngineVolumeCC,

		InsurancePolicyNumber: v.InsurancePolicyNumber,
		InsuranceExpiryDate:   v.InsuranceExpiryDate,

		Mileage:     v.Mileage,
		CurrentFuel: v.CurrentFuel,

		StatusID:   v.StatusID,
		StatusName: v.StatusName,

		DriversIDs: v.DriversIDs,

		PhotoPath: v.PhotoPath,
		PhotoURL:  "",

		Insurances:           []dto.VehicleInsuranceHistoryItem{},
		TechnicalInspections: []dto.VehicleTechnicalInspectionHistoryItem{},

		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
