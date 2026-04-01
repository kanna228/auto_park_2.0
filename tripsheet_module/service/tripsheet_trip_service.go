package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"auto_park/tripsheet_module/dto"
	"auto_park/tripsheet_module/models"
	"auto_park/tripsheet_module/repository"
)

type TripsheetTripService interface {
	Create(ctx context.Context, req dto.CreateTripsheetTripRequest) (*dto.TripsheetTripResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateTripsheetTripRequest) (*dto.TripsheetTripResponse, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dto.TripsheetTripResponse, error)
	GetAll(ctx context.Context, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error)
	GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error)
}

type tripsheetTripService struct {
	repo repository.TripsheetTripRepo
}

func NewTripsheetTripService(repo repository.TripsheetTripRepo) TripsheetTripService {
	return &tripsheetTripService{
		repo: repo,
	}
}

func (s *tripsheetTripService) Create(ctx context.Context, req dto.CreateTripsheetTripRequest) (*dto.TripsheetTripResponse, error) {
	req.RouteDescription = strings.TrimSpace(req.RouteDescription)

	if req.TripsheetID <= 0 {
		return nil, fmt.Errorf("tripsheet_id is required")
	}
	if req.RouteDescription == "" {
		return nil, fmt.Errorf("route_description is required")
	}
	if req.StatusID <= 0 {
		return nil, fmt.Errorf("status_id is required")
	}

	startTime, endTime, err := parseTripTimes(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	distancePassed := 0
	if req.DistancePassed != nil {
		distancePassed = *req.DistancePassed
	}
	if distancePassed < 0 {
		return nil, fmt.Errorf("distance_passed cannot be negative")
	}

	input := models.CreateTripsheetTripInput{
		TripsheetID:      req.TripsheetID,
		RouteDescription: req.RouteDescription,
		StartTime:        startTime,
		EndTime:          endTime,
		DistancePassed:   distancePassed,
		StatusID:         req.StatusID,
	}

	created, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	return mapTripsheetTripModelToResponse(created), nil
}

func (s *tripsheetTripService) Update(ctx context.Context, id int64, req dto.UpdateTripsheetTripRequest) (*dto.TripsheetTripResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}

	req.RouteDescription = strings.TrimSpace(req.RouteDescription)

	if req.TripsheetID <= 0 {
		return nil, fmt.Errorf("tripsheet_id is required")
	}
	if req.RouteDescription == "" {
		return nil, fmt.Errorf("route_description is required")
	}
	if req.StatusID <= 0 {
		return nil, fmt.Errorf("status_id is required")
	}

	startTime, endTime, err := parseTripTimes(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	distancePassed := 0
	if req.DistancePassed != nil {
		distancePassed = *req.DistancePassed
	}
	if distancePassed < 0 {
		return nil, fmt.Errorf("distance_passed cannot be negative")
	}

	input := models.UpdateTripsheetTripInput{
		ID:               id,
		TripsheetID:      req.TripsheetID,
		RouteDescription: req.RouteDescription,
		StartTime:        startTime,
		EndTime:          endTime,
		DistancePassed:   distancePassed,
		StatusID:         req.StatusID,
	}

	updated, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	return mapTripsheetTripModelToResponse(updated), nil
}

func (s *tripsheetTripService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	return s.repo.Delete(ctx, id)
}

func (s *tripsheetTripService) GetByID(ctx context.Context, id int64) (*dto.TripsheetTripResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *tripsheetTripService) GetAll(ctx context.Context, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	return s.repo.GetAll(ctx, filter)
}

func (s *tripsheetTripService) GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	if tripsheetID <= 0 {
		return nil, 0, fmt.Errorf("invalid tripsheet_id")
	}
	return s.repo.GetAllByTripsheetID(ctx, tripsheetID, filter)
}

func parseTripTimes(startTimeStr, endTimeStr *string) (*time.Time, *time.Time, error) {
	var startTime *time.Time
	if startTimeStr != nil && strings.TrimSpace(*startTimeStr) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*startTimeStr))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_time format, expected RFC3339")
		}
		startTime = &parsed
	}

	var endTime *time.Time
	if endTimeStr != nil && strings.TrimSpace(*endTimeStr) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*endTimeStr))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end_time format, expected RFC3339")
		}
		endTime = &parsed
	}

	return startTime, endTime, nil
}

func mapTripsheetTripModelToResponse(item *models.TripsheetTrip) *dto.TripsheetTripResponse {
	resp := &dto.TripsheetTripResponse{
		ID:               item.ID,
		TripsheetID:      item.TripsheetID,
		RouteDescription: item.RouteDescription,
		DistancePassed:   item.DistancePassed,
		StatusID:         item.StatusID,
		CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        item.UpdatedAt.Format(time.RFC3339),
	}

	if item.StartTime != nil {
		v := item.StartTime.Format(time.RFC3339)
		resp.StartTime = &v
	}
	if item.EndTime != nil {
		v := item.EndTime.Format(time.RFC3339)
		resp.EndTime = &v
	}

	return resp
}
