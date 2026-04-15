package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"auto_park/modules/vehicle_module/models"
)

type TirePlaceRepository interface {
	Create(ctx context.Context, name string) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.TirePlace, error)
	List(ctx context.Context) ([]models.TirePlace, int64, error)
	UpdateByID(ctx context.Context, id int64, name string) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type tirePlaceRepo struct {
	db *sql.DB
}

func NewTirePlaceRepository(db *sql.DB) TirePlaceRepository {
	return &tirePlaceRepo{db: db}
}

func (r *tirePlaceRepo) Create(ctx context.Context, name string) (int64, error) {
	const q = `
		INSERT INTO tire_places (name)
		VALUES ($1)
		RETURNING id;
	`

	var id int64
	err := r.db.QueryRowContext(ctx, q, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create tire place: %w", err)
	}

	return id, nil
}

func (r *tirePlaceRepo) GetByID(ctx context.Context, id int64) (*models.TirePlace, error) {
	const q = `
		SELECT id, name
		FROM tire_places
		WHERE id = $1;
	`

	var item models.TirePlace
	err := r.db.QueryRowContext(ctx, q, id).Scan(&item.ID, &item.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tire place by id: %w", err)
	}

	return &item, nil
}

func (r *tirePlaceRepo) List(ctx context.Context) ([]models.TirePlace, int64, error) {
	const countQ = `SELECT COUNT(*) FROM tire_places;`

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list tire places count: %w", err)
	}

	const q = `
		SELECT id, name
		FROM tire_places
		ORDER BY id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, 0, fmt.Errorf("list tire places: %w", err)
	}
	defer rows.Close()

	items := make([]models.TirePlace, 0)
	for rows.Next() {
		var item models.TirePlace
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, 0, fmt.Errorf("list tire places scan: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list tire places rows: %w", err)
	}

	return items, total, nil
}

func (r *tirePlaceRepo) UpdateByID(ctx context.Context, id int64, name string) (bool, error) {
	const q = `
		UPDATE tire_places
		SET name = $1
		WHERE id = $2;
	`

	res, err := r.db.ExecContext(ctx, q, strings.TrimSpace(name), id)
	if err != nil {
		return false, fmt.Errorf("update tire place by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update tire place rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *tirePlaceRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	const q = `DELETE FROM tire_places WHERE id = $1;`

	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete tire place by id: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete tire place rows affected: %w", err)
	}

	return aff > 0, nil
}
