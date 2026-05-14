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
	List(ctx context.Context, p ListTirePlacesParams) ([]models.TirePlace, int64, error)
	UpdateByID(ctx context.Context, id int64, name string) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type ListTirePlacesParams struct {
	Name   string
	Limit  int
	Offset int
	SortBy string
	Order  string
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

func (r *tirePlaceRepo) List(ctx context.Context, p ListTirePlacesParams) ([]models.TirePlace, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := make([]string, 0, 1)
	args := make([]any, 0, 4)
	argPos := 1

	if v := strings.TrimSpace(p.Name); v != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE '%%' || $%d || '%%'", argPos))
		args = append(args, v)
		argPos++
	}

	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `SELECT COUNT(*) FROM tire_places` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list tire places count: %w", err)
	}

	sortBy := normalizeTirePlaceSortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	q := fmt.Sprintf(`
		SELECT id, name
		FROM tire_places
		%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
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

func normalizeTirePlaceSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "name":
		return "name"
	case "id":
		return "id"
	default:
		return "id"
	}
}

func normalizeOrder(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "asc":
		return "ASC"
	default:
		return "DESC"
	}
}
