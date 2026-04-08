package repository

import (
	"context"
	"fmt"
	"strings"
)

func (r *vehicleRepo) UpdatePhotoPath(ctx context.Context, id int64, photoPath string) (bool, error) {
	const q = `
		UPDATE vehicles
		SET
			photo_path = $1,
			updated_at = NOW()
		WHERE id = $2;
	`

	var value any
	if strings.TrimSpace(photoPath) == "" {
		value = nil
	} else {
		value = strings.TrimSpace(photoPath)
	}

	res, err := r.db.ExecContext(ctx, q, value, id)
	if err != nil {
		return false, fmt.Errorf("update vehicle photo_path: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update vehicle photo_path rows affected: %w", err)
	}

	return aff > 0, nil
}
