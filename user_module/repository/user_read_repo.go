package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"auto_park/user_module/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UsersReadRepo interface {
	GetUserPublicByID(ctx context.Context, id int64) (*models.UserPublic, error)
	ListUsersForRole(ctx context.Context, requesterRole int64, requesterID int64) ([]models.UserPublic, error)
}

func (r *UserRepo) GetUserPublicByID(ctx context.Context, id int64) (*models.UserPublic, error) {
	q := fmt.Sprintf(`
		SELECT id, email, first_name, last_name, middle_name, iin, phone, role_id, last_seen, created_at, updated_at
		FROM %s
		WHERE id = $1
		LIMIT 1;
	`, r.usersTable())

	var u models.UserPublic
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
		&u.IIN, &u.Phone, &u.RoleID, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

// ListUsersForRole возвращает список пользователей с учётом роли запрашивающего:
// role=1 → все
// role=2 → только (2,3)
// role=3 → только сам себя
func (r *UserRepo) ListUsersForRole(ctx context.Context, requesterRole int64, requesterID int64) ([]models.UserPublic, error) {
	var (
		q    string
		args []any
	)

	baseSelect := fmt.Sprintf(`
		SELECT id, email, first_name, last_name, middle_name, iin, phone, role_id, last_seen, created_at, updated_at
		FROM %s
	`, r.usersTable())

	switch requesterRole {
	case 1:
		q = baseSelect + ` ORDER BY id ASC;`
	case 2:
		q = baseSelect + ` WHERE role_id IN (2,3) ORDER BY id ASC;`
	case 3:
		q = baseSelect + ` WHERE id = $1 ORDER BY id ASC;`
		args = append(args, requesterID)
	default:
		// неизвестная роль — ничего
		return []models.UserPublic{}, nil
	}

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []models.UserPublic
	for rows.Next() {
		var u models.UserPublic
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
			&u.IIN, &u.Phone, &u.RoleID, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return out, nil
}

// (опционально) если у тебя где-то есть rolesTable и т.п.
// убедись что usersTable возвращает schema.users
func (r *UserRepo) usersTable() string {
	return "users"
}
