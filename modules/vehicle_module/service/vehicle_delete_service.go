package service

import (
	"context"
	"fmt"
)

func (s *vehicleService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}

	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if vehicle == nil {
		return false, nil
	}

	ok, err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	if vehicle.PhotoPath != "" {
		_ = s.storage.Delete(vehicle.PhotoPath)
	}

	return true, nil
}
