package repository

import (
	"context"
	"fmt"
)

// DeleteByID удаляет машину по id. Возвращает true если удалили, false если не нашли.
func (r *vehicleRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	const q = `DELETE FROM vehicles WHERE id = $1;`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete vehicle by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete vehicle by id rows affected: %w", err)
	}

	return aff > 0, nil
}
