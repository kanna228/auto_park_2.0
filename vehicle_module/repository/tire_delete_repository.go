package repository

import (
	"context"
	"fmt"
)

func (r *tireRepoImpl) DeleteByID(ctx context.Context, id int64) (bool, error) {
	const q = `DELETE FROM tires WHERE id = $1;`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete tire by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete tire by id rows affected: %w", err)
	}

	return aff > 0, nil
}
