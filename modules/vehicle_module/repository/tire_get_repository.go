package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"auto_park/modules/vehicle_module/models"
)

func (r *tireRepoImpl) GetByID(ctx context.Context, id int64) (*models.Tire, error) {
	const q = `
		SELECT
			t.id,
			t.place_id,
			tp.name,
			t.vehicle_id,
			t.tire,
			t.mileage,
			t.max_usage,
			t.created_at,
			t.updated_at
		FROM tires t
		JOIN tire_places tp ON tp.id = t.place_id
		WHERE t.id = $1;
	`

	var item models.Tire
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&item.ID,
		&item.PlaceID,
		&item.PlaceName,
		&item.VehicleID,
		&item.Tire,
		&item.Mileage,
		&item.MaxUsage,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tire by id: %w", err)
	}

	return &item, nil
}

func (r *tireRepoImpl) List(ctx context.Context, q ListTiresParams) ([]models.Tire, int64, error) {
	where := make([]string, 0, 4)
	args := make([]any, 0, 8)

	add := func(cond string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}

	if q.VehicleID != nil {
		add("t.vehicle_id = $%d", *q.VehicleID)
	}
	if q.PlaceID != nil {
		add("t.place_id = $%d", *q.PlaceID)
	}
	if strings.TrimSpace(q.Tire) != "" {
		add("t.tire ILIKE '%%' || $%d || '%%'", strings.TrimSpace(q.Tire))
	}

	whereSQL := "TRUE"
	if len(where) > 0 {
		whereSQL = strings.Join(where, " AND ")
	}

	sortCol := "t.id"
	switch q.SortBy {
	case "id":
		sortCol = "t.id"
	case "vehicle_id":
		sortCol = "t.vehicle_id"
	case "place_id":
		sortCol = "t.place_id"
	case "mileage":
		sortCol = "t.mileage"
	case "max_usage":
		sortCol = "t.max_usage"
	case "created_at":
		sortCol = "t.created_at"
	}

	order := "DESC"
	if strings.EqualFold(q.Order, "asc") {
		order = "ASC"
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	countSQL := `SELECT COUNT(*) FROM tires t WHERE ` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list tires count: %w", err)
	}

	argsData := append([]any{}, args...)
	argsData = append(argsData, limit, offset)

	dataSQL := fmt.Sprintf(`
		SELECT
			t.id,
			t.place_id,
			tp.name,
			t.vehicle_id,
			t.tire,
			t.mileage,
			t.max_usage,
			t.created_at,
			t.updated_at
		FROM tires t
		JOIN tire_places tp ON tp.id = t.place_id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortCol, order, len(args)+1, len(args)+2)

	rows, err := r.db.QueryContext(ctx, dataSQL, argsData...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tires query: %w", err)
	}
	defer rows.Close()

	items := make([]models.Tire, 0, limit)
	for rows.Next() {
		var item models.Tire
		if err := rows.Scan(
			&item.ID,
			&item.PlaceID,
			&item.PlaceName,
			&item.VehicleID,
			&item.Tire,
			&item.Mileage,
			&item.MaxUsage,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list tires scan: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list tires rows: %w", err)
	}

	return items, total, nil
}

func (r *tireRepoImpl) GetByVehicleID(ctx context.Context, vehicleID int64) ([]models.Tire, error) {
	const q = `
		SELECT
			t.id,
			t.place_id,
			tp.name,
			t.vehicle_id,
			t.tire,
			t.mileage,
			t.max_usage,
			t.created_at,
			t.updated_at
		FROM tires t
		JOIN tire_places tp ON tp.id = t.place_id
		WHERE t.vehicle_id = $1
		ORDER BY t.place_id ASC, t.id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("get tires by vehicle id: %w", err)
	}
	defer rows.Close()

	items := make([]models.Tire, 0)
	for rows.Next() {
		var item models.Tire
		if err := rows.Scan(
			&item.ID,
			&item.PlaceID,
			&item.PlaceName,
			&item.VehicleID,
			&item.Tire,
			&item.Mileage,
			&item.MaxUsage,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("get tires by vehicle id scan: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get tires by vehicle id rows: %w", err)
	}

	return items, nil
}
