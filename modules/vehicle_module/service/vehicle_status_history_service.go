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

func (s *vehicleService) GetVehicleStatusHistoryByID(ctx context.Context, id int64) (*dto.VehicleStatusHistoryResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	item, err := s.repo.GetVehicleStatusHistoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	resp := mapVehicleStatusHistoryToDTO(*item)
	return &resp, nil
}

func (s *vehicleService) ListVehicleStatusHistory(ctx context.Context, q dto.VehicleStatusHistoryListQuery) (*dto.VehicleStatusHistoryListResponse, error) {
	if q.VehicleID < 0 {
		return nil, fmt.Errorf("vehicle_id cannot be negative")
	}
	if q.StatusID < 0 {
		return nil, fmt.Errorf("status_id cannot be negative")
	}

	startFrom, err := normalizeVehicleStatusHistoryOptionalDate(q.StartFrom, "start_from")
	if err != nil {
		return nil, err
	}
	startTo, err := normalizeVehicleStatusHistoryOptionalDate(q.StartTo, "start_to")
	if err != nil {
		return nil, err
	}
	endFrom, err := normalizeVehicleStatusHistoryOptionalDate(q.EndFrom, "end_from")
	if err != nil {
		return nil, err
	}
	endTo, err := normalizeVehicleStatusHistoryOptionalDate(q.EndTo, "end_to")
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListVehicleStatusHistory(ctx, repository.ListVehicleStatusHistoryParams{
		VehicleID: q.VehicleID,
		StatusID:  q.StatusID,
		StartFrom: startFrom,
		StartTo:   startTo,
		EndFrom:   endFrom,
		EndTo:     endTo,
		IsCurrent: q.IsCurrent,
		Limit:     q.Limit,
		Offset:    q.Offset,
		SortBy:    q.SortBy,
		Order:     q.Order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.VehicleStatusHistoryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapVehicleStatusHistoryToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.VehicleStatusHistoryListResponse{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func normalizeVehicleStatusHistoryOptionalDate(value string, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}

	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}

	return trimmed, nil
}

func mapVehicleStatusHistoryToDTO(item models.VehicleStatusHistory) dto.VehicleStatusHistoryResponse {
	isCurrent := item.EndDate == nil
	endDateDisplay := ""
	if isCurrent {
		endDateDisplay = "По настоящее время"
	} else if item.EndDate != nil {
		endDateDisplay = item.EndDate.Format("2006-01-02 15:04:05")
	}

	return dto.VehicleStatusHistoryResponse{
		ID:        item.ID,
		VehicleID: item.VehicleID,
		Vehicle: dto.VehicleStatusHistoryVehicleBriefResponse{
			ID:          item.VehicleID,
			StateNumber: item.VehicleStateNumber,
			BrandModel:  item.VehicleBrandModel,
		},
		StatusID: item.StatusID,
		Status: dto.VehicleStatusHistoryStatusBriefResponse{
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
