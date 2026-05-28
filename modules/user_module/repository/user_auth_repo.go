package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"auto_park/modules/user_module/models"
)

func (r *UserRepo) GetAuthByEmail(ctx context.Context, email string) (*models.UserAuth, error) {
	q := fmt.Sprintf(`
SELECT
    u.id,
    'user' AS account_type,
    u.email,
    u.password,
    u.iin,
    u.role_id,
    COALESCE(roles.name, '') AS role_name,
    u.first_name,
    u.last_name,
    u.middle_name,
    NULL::bigint AS driver_id,
    u.session_token,
    u.last_seen
FROM %s u
LEFT JOIN roles ON roles.id = u.role_id
WHERE LOWER(u.email) = LOWER($1)
LIMIT 1;
`, r.usersTable())

	var ua models.UserAuth
	err := r.DB.QueryRowContext(ctx, q, email).Scan(
		&ua.ID,
		&ua.AccountType,
		&ua.Email,
		&ua.PassHash,
		&ua.IIN,
		&ua.RoleID,
		&ua.RoleName,
		&ua.FirstName,
		&ua.LastName,
		&ua.MiddleName,
		&ua.DriverID,
		&ua.SessionToken,
		&ua.LastSeen,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.getDriverAuthByEmail(ctx, email)
		}
		return nil, err
	}
	return &ua, nil
}

func (r *UserRepo) getDriverAuthByEmail(ctx context.Context, email string) (*models.UserAuth, error) {
	const q = `
SELECT
    d.id,
    'driver' AS account_type,
    d.mail AS email,
    d.password,
    d.iin,
    d.role_id,
    COALESCE(roles.name, '') AS role_name,
    d.name AS first_name,
    d.surname AS last_name,
    NULLIF(d.middlename, '') AS middle_name,
    d.id AS driver_id,
    d.session_token,
    d.last_seen
FROM drivers d
LEFT JOIN roles ON roles.id = d.role_id
WHERE d.mail IS NOT NULL
  AND BTRIM(d.mail) <> ''
  AND d.password IS NOT NULL
  AND d.is_archived = FALSE
  AND LOWER(d.mail) = LOWER($1)
ORDER BY d.id ASC
LIMIT 1;
`

	var ua models.UserAuth
	err := r.DB.QueryRowContext(ctx, q, email).Scan(
		&ua.ID,
		&ua.AccountType,
		&ua.Email,
		&ua.PassHash,
		&ua.IIN,
		&ua.RoleID,
		&ua.RoleName,
		&ua.FirstName,
		&ua.LastName,
		&ua.MiddleName,
		&ua.DriverID,
		&ua.SessionToken,
		&ua.LastSeen,
	)
	if err != nil {
		return nil, err
	}
	return &ua, nil
}

func (r *UserRepo) UpdateSession(ctx context.Context, accountType string, accountID int64, token string, when time.Time) error {
	if accountType == "driver" {
		const q = `
UPDATE drivers
SET session_token = $1,
    last_seen = $2,
    updated_at = $2
WHERE id = $3;
`
		_, err := r.DB.ExecContext(ctx, q, token, when.UTC(), accountID)
		return err
	}

	q := fmt.Sprintf(`
UPDATE %s
SET session_token = $1,
    last_seen = $2,
    updated_at = $2
WHERE id = $3;
`, r.usersTable())

	_, err := r.DB.ExecContext(ctx, q, token, when.UTC(), accountID)
	return err
}
