package service

import (
	"context"

	"auto_park/modules/vehicle_module/dto"
	"auto_park/modules/vehicle_module/models"
	"auto_park/modules/vehicle_module/repository"
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

	vehicleDrivers, err := s.repo.ListDriversByIDs(ctx, v.DriversIDs)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

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

	incidents, err := s.repo.ListIncidentsByVehicleID(ctx, v.ID)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	resp.Drivers = make([]dto.VehicleDriverInfo, 0, len(vehicleDrivers))
	for _, item := range vehicleDrivers {
		resp.Drivers = append(resp.Drivers, dto.VehicleDriverInfo{
			ID:         item.ID,
			IIN:        item.IIN,
			FirstName:  item.FirstName,
			LastName:   item.LastName,
			MiddleName: item.MiddleName,
			Phone:      item.Phone,
			Email:      item.Email,
		})
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

	resp.Incidents = make([]dto.VehicleIncidentHistoryItem, 0, len(incidents))
	for _, item := range incidents {
		resp.Incidents = append(resp.Incidents, dto.VehicleIncidentHistoryItem{
			ID:               item.ID,
			IncidentTypeID:   item.IncidentTypeID,
			IncidentTypeName: item.IncidentTypeName,
			VehicleID:        item.VehicleID,
			TripsheetID:      item.TripsheetID,
			TripsheetNumber:  item.TripsheetNumber,
			Date:             item.IncidentDate,
			Time:             item.IncidentTime,
			Location:         item.Location,
			Text:             item.Description,
			Driver: dto.VehicleIncidentDriverInfo{
				ID:         item.DriverID,
				IIN:        item.DriverIIN,
				FirstName:  item.DriverFirstName,
				LastName:   item.DriverLastName,
				MiddleName: item.DriverMiddleName,
				Phone:      item.DriverPhone,
				Email:      item.DriverEmail,
			},
			Mechanic: dto.VehicleIncidentMechanicInfo{
				ID:         item.MechanicID,
				IIN:        item.MechanicIIN,
				FirstName:  item.MechanicFirstName,
				LastName:   item.MechanicLastName,
				MiddleName: item.MechanicMiddleName,
				Phone:      item.MechanicPhone,
				Email:      item.MechanicEmail,
				RoleID:     item.MechanicRoleID,
				RoleName:   item.MechanicRoleName,
			},
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
		Drivers:    []dto.VehicleDriverInfo{},

		PhotoPath: v.PhotoPath,
		PhotoURL:  "",

		Insurances:           []dto.VehicleInsuranceHistoryItem{},
		TechnicalInspections: []dto.VehicleTechnicalInspectionHistoryItem{},
		Incidents:            []dto.VehicleIncidentHistoryItem{},

		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
