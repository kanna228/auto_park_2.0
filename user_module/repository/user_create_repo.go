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

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrRoleNotFound = errors.New("role not found")
)

func (r *UserRepo) rolesTable() string { return "roles" }

func (r *UserRepo) EnsureRoleExists(ctx context.Context, roleID int64) error {
	q := fmt.Sprintf(`SELECT 1 FROM %s WHERE id=$1`, r.rolesTable())
	var one int
	if err := r.DB.QueryRowContext(ctx, q, roleID).Scan(&one); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRoleNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) EmailInUse(ctx context.Context, email string) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s WHERE LOWER(email)=LOWER($1) LIMIT 1`, r.usersTable())
	var one int
	err := r.DB.QueryRowContext(ctx, q, strings.TrimSpace(email)).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *UserRepo) GetRoleNameByID(ctx context.Context, id int64) (string, error) {
	q := fmt.Sprintf(`SELECT name FROM %s WHERE id=$1`, r.rolesTable())
	var name string
	if err := r.DB.QueryRowContext(ctx, q, id).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRoleNotFound
		}
		return "", err
	}
	return name, nil
}

type CreateUserParams struct {
	Email      string
	FirstName  string
	LastName   string
	MiddleName *string
	IIN        string
	Phone      *string
	RoleID     int64
	PassHash   string
}

func (r *UserRepo) CreateUser(ctx context.Context, p CreateUserParams) (*models.User, error) {
	now := time.Now().UTC()

	q := fmt.Sprintf(`
INSERT INTO %s
  (email, first_name, last_name, middle_name, iin, phone, password, role_id, created_at, updated_at)
VALUES
  ($1,    $2,         $3,        $4,          $5,  $6,    $7,       $8,      $9,        $10)
RETURNING id, email, first_name, last_name, middle_name, iin, phone, password, role_id, session_token, last_seen, created_at, updated_at;
`, r.usersTable())

	var u models.User
	err := r.DB.QueryRowContext(ctx, q,
		p.Email, p.FirstName, p.LastName, p.MiddleName,
		p.IIN, p.Phone, p.PassHash, p.RoleID, now, now,
	).Scan(
		&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.MiddleName,
		&u.IIN, &u.Phone, &u.PasswordHash, &u.RoleID, &u.SessionToken, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "unique") && strings.Contains(low, "email") {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return &u, nil
}
