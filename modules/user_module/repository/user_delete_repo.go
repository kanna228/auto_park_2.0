package repository

import (
	"context"
	"fmt"
)

func (r *UserRepo) DeleteUserByID(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`
		UPDATE %s
		SET email = CONCAT('__archived_user_', id, '@auto-park.local'),
			iin = CONCAT('ARCH', id),
			is_archived = TRUE,
			is_active = FALSE,
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND is_archived = FALSE;
	`, r.usersTable())

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
