package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/models"
	"auto_park/modules/tripsheet_module/repository"
)

type WaybillRoutePointService interface {
	List(ctx context.Context, waybillID int64) ([]dto.WaybillRoutePointResponse, error)
	Create(ctx context.Context, waybillID int64, req dto.WaybillRoutePointCreateRequest) (int64, error)
	Update(ctx context.Context, waybillID int64, routeID int64, req dto.WaybillRoutePointUpdateRequest) (bool, error)
	Delete(ctx context.Context, waybillID int64, routeID int64) (bool, error)
}

type waybillRoutePointService struct {
	repo repository.WaybillRoutePointRepository
}

func NewWaybillRoutePointService(repo repository.WaybillRoutePointRepository) WaybillRoutePointService {
	return &waybillRoutePointService{repo: repo}
}

func (s *waybillRoutePointService) List(ctx context.Context, waybillID int64) ([]dto.WaybillRoutePointResponse, error) {
	if waybillID <= 0 {
		return nil, fmt.Errorf("invalid waybill id")
	}
	items, err := s.repo.List(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.WaybillRoutePointResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapRoutePointToDTO(item))
	}
	return out, nil
}

func (s *waybillRoutePointService) Create(ctx context.Context, waybillID int64, req dto.WaybillRoutePointCreateRequest) (int64, error) {
	if waybillID <= 0 {
		return 0, fmt.Errorf("invalid waybill id")
	}
	if req.SeqNumber <= 0 {
		return 0, fmt.Errorf("seq_number must be greater than 0")
	}
	destination := strings.TrimSpace(req.Destination)
	if destination == "" {
		return 0, fmt.Errorf("destination is required")
	}
	if err := validateRoutePointTimes(req.ArrivalTime, req.HospitalizationTime, req.LPUArrivalTime, req.ReleaseTime); err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, waybillID, repository.CreateWaybillRoutePointParams{
		SeqNumber:           req.SeqNumber,
		Destination:         destination,
		ArrivalTime:         normalizeOptionalTimeString(req.ArrivalTime),
		HospitalizationTime: normalizeOptionalTimeString(req.HospitalizationTime),
		LPUArrivalTime:      normalizeOptionalTimeString(req.LPUArrivalTime),
		ReleaseTime:         normalizeOptionalTimeString(req.ReleaseTime),
	})
}

func (s *waybillRoutePointService) Update(ctx context.Context, waybillID int64, routeID int64, req dto.WaybillRoutePointUpdateRequest) (bool, error) {
	if waybillID <= 0 {
		return false, fmt.Errorf("invalid waybill id")
	}
	if routeID <= 0 {
		return false, fmt.Errorf("invalid route id")
	}
	if req.SeqNumber != nil && *req.SeqNumber <= 0 {
		return false, fmt.Errorf("seq_number must be greater than 0")
	}
	if req.Destination != nil && strings.TrimSpace(*req.Destination) == "" {
		return false, fmt.Errorf("destination cannot be empty")
	}
	if err := validateRoutePointTimes(req.ArrivalTime, req.HospitalizationTime, req.LPUArrivalTime, req.ReleaseTime); err != nil {
		return false, err
	}

	var destination *string
	if req.Destination != nil {
		v := strings.TrimSpace(*req.Destination)
		destination = &v
	}

	return s.repo.Update(ctx, waybillID, routeID, repository.UpdateWaybillRoutePointParams{
		SeqNumber:           req.SeqNumber,
		Destination:         destination,
		ArrivalTime:         normalizeOptionalTimeString(req.ArrivalTime),
		HospitalizationTime: normalizeOptionalTimeString(req.HospitalizationTime),
		LPUArrivalTime:      normalizeOptionalTimeString(req.LPUArrivalTime),
		ReleaseTime:         normalizeOptionalTimeString(req.ReleaseTime),
	})
}

func (s *waybillRoutePointService) Delete(ctx context.Context, waybillID int64, routeID int64) (bool, error) {
	if waybillID <= 0 {
		return false, fmt.Errorf("invalid waybill id")
	}
	if routeID <= 0 {
		return false, fmt.Errorf("invalid route id")
	}
	return s.repo.Delete(ctx, waybillID, routeID)
}

func mapRoutePointToDTO(item models.WaybillRoutePoint) dto.WaybillRoutePointResponse {
	return dto.WaybillRoutePointResponse{
		ID:                  item.ID,
		WaybillID:           item.WaybillID,
		SeqNumber:           item.SeqNumber,
		Destination:         item.Destination,
		ArrivalTime:         item.ArrivalTime,
		HospitalizationTime: item.HospitalizationTime,
		LPUArrivalTime:      item.LPUArrivalTime,
		ReleaseTime:         item.ReleaseTime,
		CreatedAt:           item.CreatedAt.Format(time.RFC3339),
	}
}

func validateRoutePointTimes(values ...*string) error {
	for _, value := range values {
		if value == nil || strings.TrimSpace(*value) == "" {
			continue
		}
		if _, err := parseRoutePointTime(*value); err != nil {
			return err
		}
	}
	return nil
}

func normalizeOptionalTimeString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		empty := ""
		return &empty
	}
	parsed, _ := parseRoutePointTime(trimmed)
	normalized := parsed.Format("15:04:05")
	return &normalized
}

func parseRoutePointTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{"15:04:05", "15:04"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("time fields must be in HH:MM or HH:MM:SS format")
}
