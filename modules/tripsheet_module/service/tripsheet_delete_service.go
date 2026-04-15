package service

import (
	"context"
	"fmt"
)

func (s *tripsheetService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}

	return s.repo.DeleteTripsWithTripsheet(ctx, id)
}
