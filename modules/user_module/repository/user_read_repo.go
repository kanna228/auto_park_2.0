package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/user_module/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UsersReadRepo interface {
	GetUserPublicByID(ctx context.Context, id int64, includeArchived bool) (*models.UserPublic, error)
	GetDriverPublicByID(ctx context.Context, id int64) (*models.UserPublic, error)
	ListUsersForRole(ctx context.Context, requesterRole int64, requesterID int64, p ListUsersParams) ([]models.UserPublic, int64, error)
}

type ListUsersParams struct {
	Limit           int
	Offset          int
	SortBy          string
	Order           string
	IncludeArchived bool
}

func (r *UserRepo) GetUserPublicByID(ctx context.Context, id int64, includeArchived bool) (*models.UserPublic, error) {
	archiveCond := "AND u.is_archived = FALSE"
	if includeArchived {
		archiveCond = ""
	}
	q := fmt.Sprintf(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.middle_name, u.iin, u.phone, u.role_id,
		       COALESCE(r.name, '') AS role_name,
		       u.is_active, u.is_archived, u.deleted_at, u.last_seen, u.created_at, u.updated_at
		FROM %s u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
		  %s
		LIMIT 1;
	`, r.usersTable(), archiveCond)

	var u models.UserPublic
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
		&u.IIN, &u.Phone, &u.RoleID, &u.RoleName, &u.IsActive, &u.IsArchived,
		&u.DeletedAt, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) GetDriverPublicByID(ctx context.Context, id int64) (*models.UserPublic, error) {
	const q = `
		SELECT d.id, COALESCE(d.mail, '') AS email, d.name, d.surname, NULLIF(d.middlename, ''),
		       d.iin, NULLIF(d.phone, ''), COALESCE(d.role_id, 0), COALESCE(r.name, '') AS role_name,
		       TRUE AS is_active, d.is_archived, d.deleted_at, d.last_seen, d.created_at, d.updated_at
		FROM drivers d
		LEFT JOIN roles r ON r.id = d.role_id
		WHERE d.id = $1
		  AND d.is_archived = FALSE
		LIMIT 1;
	`
	var u models.UserPublic
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
		&u.IIN, &u.Phone, &u.RoleID, &u.RoleName, &u.IsActive, &u.IsArchived,
		&u.DeletedAt, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get driver by id: %w", err)
	}
	return &u, nil
}

// ListUsersForRole возвращает список пользователей с учётом роли запрашивающего:
// role=1 → все
// role=2 → только (2,3)
// role=3 → только сам себя
func (r *UserRepo) ListUsersForRole(ctx context.Context, requesterRole int64, requesterID int64, p ListUsersParams) ([]models.UserPublic, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if !p.IncludeArchived {
		conds = append(conds, "u.is_archived = FALSE")
	}
	switch requesterRole {
	case 1:
	case 2:
		conds = append(conds, "u.role_id IN (2,3)")
	case 3:
		args = append(args, requesterID)
		conds = append(conds, "u.id = $1")
	default:
		return []models.UserPublic{}, 0, nil
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM %s u%s;`, r.usersTable(), where)
	var total int64
	if err := r.DB.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	args = append(args, limit, offset)
	sortBy := normalizeUserSortBy(p.SortBy)
	order := normalizeUserOrder(p.Order)
	q := fmt.Sprintf(`
		SELECT u.id, u.email, u.first_name, u.last_name, u.middle_name, u.iin, u.phone, u.role_id,
		       COALESCE(r.name, '') AS role_name,
		       u.is_active, u.is_archived, u.deleted_at, u.last_seen, u.created_at, u.updated_at
		FROM %s u
		LEFT JOIN roles r ON r.id = u.role_id
		%s
		ORDER BY %s %s, u.id ASC
		LIMIT $%d OFFSET $%d;
	`, r.usersTable(), where, sortBy, order, len(args)-1, len(args))

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []models.UserPublic
	for rows.Next() {
		var u models.UserPublic
		if err := rows.Scan(
			&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
			&u.IIN, &u.Phone, &u.RoleID, &u.RoleName, &u.IsActive, &u.IsArchived,
			&u.DeletedAt, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return out, total, nil
}

// (опционально) если у тебя где-то есть rolesTable и т.п.
// убедись что usersTable возвращает schema.users
func (r *UserRepo) usersTable() string {
	return "users"
}

func normalizeUserSortBy(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "id":
		return "u.id"
	case "email":
		return "u.email"
	case "first_name":
		return "u.first_name"
	case "last_name":
		return "u.last_name"
	case "role_id":
		return "u.role_id"
	case "is_active":
		return "u.is_active"
	case "created_at":
		return "u.created_at"
	default:
		return "u.id"
	}
}

func normalizeUserOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "desc") {
		return "DESC"
	}
	return "ASC"
}
