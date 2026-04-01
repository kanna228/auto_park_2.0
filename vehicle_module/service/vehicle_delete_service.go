package service

import (
	"context"
	"fmt"
)

// DeleteByID — удаляет машину по id
func (s *vehicleService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeleteByID(ctx, id)
}
