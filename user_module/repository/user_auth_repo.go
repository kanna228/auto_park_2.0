package repository

import (
	"context"
	"fmt"
	"time"

	"auto_park/user_module/models"
)

func (r *UserRepo) GetAuthByEmail(ctx context.Context, email string) (*models.UserAuth, error) {
	q := fmt.Sprintf(`
SELECT id, email, password, iin, role_id, first_name, last_name, middle_name, session_token, last_seen
FROM %s
WHERE LOWER(email) = LOWER($1)
LIMIT 1;
`, r.usersTable())

	var ua models.UserAuth
	err := r.DB.QueryRowContext(ctx, q, email).Scan(
		&ua.ID,
		&ua.Email,
		&ua.PassHash,
		&ua.IIN,
		&ua.RoleID,
		&ua.FirstName,
		&ua.LastName,
		&ua.MiddleName,
		&ua.SessionToken,
		&ua.LastSeen,
	)
	if err != nil {
		return nil, err
	}
	return &ua, nil
}

func (r *UserRepo) UpdateSession(ctx context.Context, userID int64, token string, when time.Time) error {
	q := fmt.Sprintf(`
UPDATE %s
SET session_token = $1,
    last_seen = $2,
    updated_at = $2
WHERE id = $3;
`, r.usersTable())

	_, err := r.DB.ExecContext(ctx, q, token, when.UTC(), userID)
	return err
}
