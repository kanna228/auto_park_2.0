package repository

import (
	"context"
	"fmt"
)

func (r *UserRepo) DeleteUserByID(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id=$1`, r.usersTable())

	res, err := r.DB.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrUserNotFound
	}
	return nil
}
