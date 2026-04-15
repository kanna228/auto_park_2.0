package service

import (
	"context"
	"fmt"
)

func (s *tireService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeleteByID(ctx, id)
}
