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
		Search:      q.Search,
		BoardNumber: q.BoardNumber,
		StateNumber: q.StateNumber,
		VIN:         q.VIN,
		BrandModel:  q.BrandModel,

		StatusID:   q.StatusID,
		StatusName: q.StatusName,

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

	installedParts, err := s.repo.ListInstalledPartsByVehicleID(ctx, v.ID)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	statusHistory, _, err := s.repo.ListVehicleStatusHistory(ctx, repository.ListVehicleStatusHistoryParams{
		VehicleID: v.ID,
		Limit:     8,
		Offset:    0,
		SortBy:    "start_date",
		Order:     "desc",
	})
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	tripsheets, err := s.repo.ListTripsheetsByVehicleID(ctx, v.ID)
	if err != nil {
		return dto.VehicleResponse{}, err
	}

	vehicleServices, err := s.repo.ListVehicleServicesByVehicleID(ctx, v.ID)
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
			StatusID:   item.StatusID,
			StatusName: item.StatusName,
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
				StatusID:   item.DriverStatusID,
				StatusName: item.DriverStatusName,
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

	resp.InstalledParts = make([]dto.VehicleInstalledPartHistoryItem, 0, len(installedParts))
	for _, item := range installedParts {
		resp.InstalledParts = append(resp.InstalledParts, dto.VehicleInstalledPartHistoryItem{
			ID:                   item.ID,
			PartID:               item.PartID,
			CatalogPartID:        item.CatalogPartID,
			PartName:             item.PartName,
			PartCategory:         item.PartCategory,
			IsConsumable:         item.IsConsumable,
			VehicleID:            item.VehicleID,
			InstalledAt:          item.InstalledAt,
			PlannedReplacementAt: item.PlannedReplacementAt,
			Quantity:             item.Quantity,
			UnitPrice:            item.UnitPrice,
			TotalPrice:           item.TotalPrice,
			InstalledByUserID:    item.InstalledByUserID,
			InstallerEmail:       item.InstallerEmail,
			InstallerFullName:    item.InstallerFullName,
			IsActive:             item.IsActive,
			CreatedAt:            item.CreatedAt,
			UpdatedAt:            item.UpdatedAt,
		})
	}

	resp.StatusHistory = make([]dto.VehicleStatusHistoryItem, 0, len(statusHistory))
	for _, item := range statusHistory {
		resp.StatusHistory = append(resp.StatusHistory, mapVehicleStatusHistoryToVehicleItemDTO(item))
	}

	resp.Tripsheets = make([]dto.VehicleTripsheetHistoryItem, 0, len(tripsheets))
	for _, item := range tripsheets {
		resp.Tripsheets = append(resp.Tripsheets, mapVehicleTripsheetToDTO(item))
	}

	resp.Services = make([]dto.VehicleServiceHistoryItem, 0, len(vehicleServices))
	for _, item := range vehicleServices {
		resp.Services = append(resp.Services, mapVehicleServiceHistoryToDTO(item))
	}

	return resp, nil
}

func mapVehicleStatusHistoryToVehicleItemDTO(item models.VehicleStatusHistory) dto.VehicleStatusHistoryItem {
	isCurrent := item.EndDate == nil
	endDateDisplay := ""
	if isCurrent {
		endDateDisplay = "По настоящее время"
	} else if item.EndDate != nil {
		endDateDisplay = item.EndDate.Format("2006-01-02 15:04:05")
	}

	return dto.VehicleStatusHistoryItem{
		ID:        item.ID,
		VehicleID: item.VehicleID,
		StatusID:  item.StatusID,
		Status: dto.VehicleStatusBriefInfo{
			ID:   item.StatusID,
			Name: item.StatusName,
		},
		StartDate:      item.StartDate,
		EndDate:        item.EndDate,
		EndDateDisplay: endDateDisplay,
		IsCurrent:      isCurrent,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func mapVehicleTripsheetToDTO(item repository.VehicleTripsheetHistoryRow) dto.VehicleTripsheetHistoryItem {
	trips := make([]dto.VehicleTripsheetTripHistoryItem, 0, len(item.Trips))
	for _, trip := range item.Trips {
		trips = append(trips, dto.VehicleTripsheetTripHistoryItem{
			ID:               trip.ID,
			TripsheetID:      trip.TripsheetID,
			RouteDescription: trip.RouteDescription,
			StartTime:        trip.StartTime,
			EndTime:          trip.EndTime,
			DistancePassed:   trip.DistancePassed,
			StatusID:         trip.StatusID,
			Status: dto.VehicleTripsheetStatusBriefInfo{
				ID:   trip.StatusID,
				Name: trip.StatusName,
			},
			CreatedAt: trip.CreatedAt,
			UpdatedAt: trip.UpdatedAt,
		})
	}

	return dto.VehicleTripsheetHistoryItem{
		ID:                         item.ID,
		TripsheetNumber:            item.TripsheetNumber,
		TripsheetDate:              item.TripsheetDate.Format("2006-01-02"),
		VehicleID:                  item.VehicleID,
		VehicleBrand:               item.VehicleBrand,
		VehiclePlateNumber:         item.VehiclePlateNumber,
		DriverID:                   item.DriverID,
		DriverShiftID:              item.DriverShiftID,
		Driver:                     mapVehicleTripsheetDriverToDTO(item),
		DriverLastName:             firstStringPtr(item.DriverLastName, item.DriverSnapshotLastName),
		DriverFirstName:            firstStringPtr(item.DriverFirstName, item.DriverSnapshotFirstName),
		DriverMiddleName:           firstStringPtr(item.DriverMiddleName, item.DriverSnapshotMiddleName),
		StartTime:                  item.StartTime,
		EndTime:                    item.EndTime,
		MileageStart:               item.MileageStart,
		MileageEnd:                 item.MileageEnd,
		FuelStart:                  item.FuelStart,
		FuelIssued:                 item.FuelIssued,
		FuelConsumptionTheoretical: item.FuelConsumptionTheoretical,
		FuelConsumptionActual:      item.FuelConsumptionActual,
		StatusID:                   item.StatusID,
		Status: dto.VehicleTripsheetStatusBriefInfo{
			ID:   item.StatusID,
			Name: item.StatusName,
		},
		Trips:     trips,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapVehicleTripsheetDriverToDTO(item repository.VehicleTripsheetHistoryRow) *dto.VehicleTripsheetDriverInfo {
	if item.DriverID == nil && item.DriverIIN == nil && item.DriverFirstName == nil && item.DriverLastName == nil && item.DriverSnapshotFirstName == nil && item.DriverSnapshotLastName == nil {
		return nil
	}

	return &dto.VehicleTripsheetDriverInfo{
		ID:         item.DriverID,
		IIN:        item.DriverIIN,
		FirstName:  firstStringPtr(item.DriverFirstName, item.DriverSnapshotFirstName),
		LastName:   firstStringPtr(item.DriverLastName, item.DriverSnapshotLastName),
		MiddleName: firstStringPtr(item.DriverMiddleName, item.DriverSnapshotMiddleName),
		Phone:      item.DriverPhone,
		Email:      item.DriverEmail,
		StatusID:   item.DriverStatusID,
		StatusName: derefString(item.DriverStatusName),
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func firstStringPtr(values ...*string) *string {
	for _, value := range values {
		if value != nil && *value != "" {
			return value
		}
	}
	return nil
}

func mapVehicleServiceHistoryToDTO(item repository.VehicleServiceHistoryRow) dto.VehicleServiceHistoryItem {
	return dto.VehicleServiceHistoryItem{
		ID:     item.ID,
		TypeID: item.TypeID,
		Type: dto.VehicleServiceTypeInfo{
			ID:          item.TypeID,
			Name:        item.TypeName,
			Description: item.TypeDescription,
		},
		PartID: item.PartID,
		Part: dto.VehicleServicePartInfo{
			ID:          item.PartID,
			Name:        item.PartName,
			Description: item.PartDescription,
		},
		VehicleID: item.VehicleID,
		Date:      item.ServiceDate,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
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
		InstalledParts:       []dto.VehicleInstalledPartHistoryItem{},
		StatusHistory:        []dto.VehicleStatusHistoryItem{},
		Tripsheets:           []dto.VehicleTripsheetHistoryItem{},
		Services:             []dto.VehicleServiceHistoryItem{},

		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
