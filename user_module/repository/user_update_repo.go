package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/user_module/models"
)

type UpdateUserAdminParams struct {
	ID         int64
	Email      *string
	FirstName  *string
	LastName   *string
	MiddleName *string
	IIN        *string
	Phone      *string
	RoleID     *int64
	PassHash   *string // bcrypt hash
}

// EmailBelongsToOther checks that email is not used by another user
func (r *UserRepo) EmailBelongsToOther(ctx context.Context, email string, userID int64) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s WHERE LOWER(email)=LOWER($1) AND id <> $2 LIMIT 1`, r.usersTable())
	var one int
	err := r.DB.QueryRowContext(ctx, q, strings.TrimSpace(email), userID).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateUserAdmin partial update (nil => no change)
func (r *UserRepo) UpdateUserAdmin(ctx context.Context, p UpdateUserAdminParams) (*models.UserPublic, error) {
	now := time.Now().UTC()

	q := fmt.Sprintf(`
UPDATE %s SET
	email       = COALESCE($1, email),
	first_name  = COALESCE($2, first_name),
	last_name   = COALESCE($3, last_name),
	middle_name = COALESCE($4, middle_name),
	iin         = COALESCE($5, iin),
	phone       = COALESCE($6, phone),
	role_id     = COALESCE($7, role_id),
	password    = COALESCE($8, password),
	updated_at  = $9
WHERE id = $10
RETURNING id, email, first_name, last_name, middle_name, iin, phone, role_id, last_seen, created_at, updated_at;
`, r.usersTable())

	var u models.UserPublic
	err := r.DB.QueryRowContext(ctx, q,
		p.Email, p.FirstName, p.LastName, p.MiddleName,
		p.IIN, p.Phone, p.RoleID, p.PassHash, now, p.ID,
	).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
		&u.IIN, &u.Phone, &u.RoleID, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "unique") && strings.Contains(low, "email") {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &u, nil
}
