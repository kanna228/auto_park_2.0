package service

import (
	"auto_park/modules/tripsheet_module/dto"
	"context"
)

func (s *tripsheetService) GetAll(ctx context.Context, f dto.TripsheetFilter) ([]dto.TripsheetResponse, int, error) {
	return s.repo.GetAll(ctx, f)
}

func (s *tripsheetService) GetByID(ctx context.Context, id int64) (*dto.TripsheetResponse, error) {
	return s.repo.GetByID(ctx, id)
}
