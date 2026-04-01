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

var ErrNotFound = errors.New("not found")

type DriverRepo struct {
	db     *sql.DB
	schema string
}

func NewDriverRepo(db *sql.DB, schema string) *DriverRepo {
	return &DriverRepo{db: db, schema: schema}
}

func (r *DriverRepo) table() string {
	return "drivers"
}

func (r *DriverRepo) Create(ctx context.Context, d *models.Driver) (*models.Driver, error) {
	q := fmt.Sprintf(`
		INSERT INTO %s (iin, name, surname, middlename, phone, mail)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, iin, name, surname, middlename, phone, mail, created_at, updated_at
	`, r.table())

	row := r.db.QueryRowContext(ctx, q,
		d.IIN, d.Name, d.Surname, nullIfEmpty(d.Middlename), nullIfEmpty(d.Phone), nullIfEmpty(d.Mail),
	)

	out := &models.Driver{}
	var middlename, phone, mail sql.NullString
	if err := row.Scan(
		&out.ID, &out.IIN, &out.Name, &out.Surname,
		&middlename, &phone, &mail,
		&out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	out.Middlename = middlename.String
	out.Phone = phone.String
	out.Mail = mail.String
	return out, nil
}

func (r *DriverRepo) GetByID(ctx context.Context, id int64) (*models.Driver, error) {
	q := fmt.Sprintf(`
		SELECT id, iin, name, surname, middlename, phone, mail, created_at, updated_at
		FROM %s
		WHERE id=$1
	`, r.table())

	out := &models.Driver{}
	var middlename, phone, mail sql.NullString
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&out.ID, &out.IIN, &out.Name, &out.Surname,
		&middlename, &phone, &mail,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	out.Middlename = middlename.String
	out.Phone = phone.String
	out.Mail = mail.String
	return out, nil
}

func (r *DriverRepo) List(ctx context.Context, limit, offset int) ([]models.Driver, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	q := fmt.Sprintf(`
		SELECT id, iin, name, surname, middlename, phone, mail, created_at, updated_at
		FROM %s
		ORDER BY surname ASC, name ASC, id ASC
		LIMIT $1 OFFSET $2
	`, r.table())

	rows, err := r.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Driver
	for rows.Next() {
		var d models.Driver
		var middlename, phone, mail sql.NullString
		if err := rows.Scan(
			&d.ID, &d.IIN, &d.Name, &d.Surname,
			&middlename, &phone, &mail,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		d.Middlename = middlename.String
		d.Phone = phone.String
		d.Mail = mail.String
		res = append(res, d)
	}
	return res, rows.Err()
}

func (r *DriverRepo) Update(ctx context.Context, id int64, upd map[string]any) (*models.Driver, error) {
	// Собираем динамический UPDATE
	if len(upd) == 0 {
		return r.GetByID(ctx, id)
	}

	allowed := map[string]bool{
		"iin": true, "name": true, "surname": true,
		"middlename": true, "phone": true, "mail": true,
	}

	setParts := make([]string, 0, len(upd)+1)
	args := make([]any, 0, len(upd)+1)
	i := 1

	for k, v := range upd {
		if !allowed[k] {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s=$%d", k, i))
		args = append(args, v)
		i++
	}

	// updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at=$%d", i))
	args = append(args, time.Now())
	i++

	if len(setParts) == 1 { // только updated_at
		return r.GetByID(ctx, id)
	}

	args = append(args, id)

	q := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE id=$%d
		RETURNING id, iin, name, surname, middlename, phone, mail, created_at, updated_at
	`, r.table(), strings.Join(setParts, ", "), i)

	row := r.db.QueryRowContext(ctx, q, args...)

	out := &models.Driver{}
	var middlename, phone, mail sql.NullString
	if err := row.Scan(
		&out.ID, &out.IIN, &out.Name, &out.Surname,
		&middlename, &phone, &mail,
		&out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	out.Middlename = middlename.String
	out.Phone = phone.String
	out.Mail = mail.String
	return out, nil
}

func (r *DriverRepo) Delete(ctx context.Context, id int64) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE id=$1`, r.table())
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

// helpers

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
