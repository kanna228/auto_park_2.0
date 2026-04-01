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
	resp := mapVehicleToDTO(*v)
	return &resp, nil
}

func (s *vehicleService) List(ctx context.Context, q dto.VehicleListQuery) (*dto.VehicleListResponse, error) {
	params := repository.ListVehiclesParams{
		BoardNumber: q.BoardNumber,
		StateNumber: q.StateNumber,
		VIN:         q.VIN,
		BrandModel:  q.BrandModel,

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
		out = append(out, mapVehicleToDTO(v))
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

		DriversIDs: v.DriversIDs,

		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
