package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/warehouse_module/models"
)

var ErrPartIDExists = errors.New("part_id already exists")
var ErrPartNameExists = errors.New("part name already exists")

type CreatePartParams struct {
	PartID       string
	Name         string
	Quantity     int64
	Category     string
	Dimensions   *string
	Manufacturer *string
}

type UpdatePartParams struct {
	Name         string
	Quantity     int64
	Category     string
	Dimensions   *string
	Manufacturer *string
}

type ListPartsParams struct {
	PartID   string
	Name     string
	Category string
	Limit    int
	Offset   int
	SortBy   string
	Order    string
}

type PartRepository interface {
	Create(ctx context.Context, p CreatePartParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Part, error)
	List(ctx context.Context, p ListPartsParams) ([]models.Part, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdatePartParams) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
}

type partRepo struct {
	db *sql.DB
}

func NewPartRepository(db *sql.DB) PartRepository {
	return &partRepo{db: db}
}

func (r *partRepo) Create(ctx context.Context, p CreatePartParams) (int64, error) {
	const q = `
		INSERT INTO parts_catalog (part_id, name, quantity, category, dimensions, manufacturer, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.PartID, p.Name, p.Quantity, p.Category, p.Dimensions, p.Manufacturer).Scan(&id); err != nil {
		return 0, mapConstraintError(err)
	}
	return id, nil
}

func (r *partRepo) GetByID(ctx context.Context, id int64) (*models.Part, error) {
	const q = `
		SELECT id, part_id, name, quantity, category, dimensions, manufacturer, created_at, updated_at
		FROM parts_catalog
		WHERE id = $1;
	`

	var item models.Part
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&item.ID,
		&item.PartID,
		&item.Name,
		&item.Quantity,
		&item.Category,
		&item.Dimensions,
		&item.Manufacturer,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get part by id: %w", err)
	}
	return &item, nil
}

func (r *partRepo) List(ctx context.Context, p ListPartsParams) ([]models.Part, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := normalizeSortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	conds := make([]string, 0, 3)
	args := make([]any, 0, 8)
	argPos := 1

	if v := strings.TrimSpace(p.PartID); v != "" {
		conds = append(conds, fmt.Sprintf("part_id ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(p.Name); v != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(p.Category); v != "" {
		conds = append(conds, fmt.Sprintf("category ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}

	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `SELECT COUNT(*) FROM parts_catalog` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list parts count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT id, part_id, name, quantity, category, dimensions, manufacturer, created_at, updated_at
		FROM parts_catalog
		%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list parts: %w", err)
	}
	defer rows.Close()

	items := make([]models.Part, 0)
	for rows.Next() {
		var item models.Part
		if err := rows.Scan(
			&item.ID,
			&item.PartID,
			&item.Name,
			&item.Quantity,
			&item.Category,
			&item.Dimensions,
			&item.Manufacturer,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list parts scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list parts rows: %w", err)
	}

	return items, total, nil
}

func (r *partRepo) UpdateByID(ctx context.Context, id int64, p UpdatePartParams) (bool, error) {
	const q = `
		UPDATE parts_catalog
		SET name = $1,
			quantity = $2,
			category = $3,
			dimensions = $4,
			manufacturer = $5,
			updated_at = NOW()
		WHERE id = $6;
	`
	res, err := r.db.ExecContext(ctx, q, p.Name, p.Quantity, p.Category, p.Dimensions, p.Manufacturer, id)
	if err != nil {
		return false, mapConstraintError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update part rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *partRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	const q = `DELETE FROM parts_catalog WHERE id = $1;`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete part by id: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete part rows affected: %w", err)
	}
	return aff > 0, nil
}

func normalizeSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "part_id":
		return "part_id"
	case "name":
		return "name"
	case "quantity":
		return "quantity"
	case "category":
		return "category"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}

func normalizeOrder(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "ASC"
	}
	return "DESC"
}

func mapConstraintError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "parts_catalog_part_id_key"):
		return ErrPartIDExists
	case strings.Contains(msg, "parts_catalog_name_key"):
		return ErrPartNameExists
	default:
		return err
	}
}
