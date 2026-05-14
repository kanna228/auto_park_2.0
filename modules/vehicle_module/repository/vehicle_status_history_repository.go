package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"auto_park/modules/vehicle_module/models"
)

type ListVehicleStatusHistoryParams struct {
	VehicleID int64
	StatusID  int64
	StartFrom string
	StartTo   string
	EndFrom   string
	EndTo     string
	IsCurrent *bool
	Limit     int
	Offset    int
	SortBy    string
	Order     string
}

func (r *vehicleRepo) GetVehicleStatusHistoryByID(ctx context.Context, id int64) (*models.VehicleStatusHistory, error) {
	const q = `
		SELECT
			h.id,
			h.vehicle_id,
			v.state_number,
			v.brand_model,
			h.status_id,
			vs.name AS status_name,
			h.start_date,
			h.end_date,
			h.created_at,
			h.updated_at
		FROM vehicle_status_history h
		INNER JOIN vehicles v ON v.id = h.vehicle_id
		INNER JOIN vehicle_status vs ON vs.id = h.status_id
		WHERE h.id = $1;
	`

	item, err := scanVehicleStatusHistory(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get vehicle status history by id: %w", err)
	}

	return item, nil
}

func (r *vehicleRepo) ListVehicleStatusHistory(ctx context.Context, p ListVehicleStatusHistoryParams) ([]models.VehicleStatusHistory, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := make([]string, 0, 8)
	args := make([]any, 0, 10)
	argPos := 1

	if p.VehicleID > 0 {
		conds = append(conds, fmt.Sprintf("h.vehicle_id = $%d", argPos))
		args = append(args, p.VehicleID)
		argPos++
	}

	if p.StatusID > 0 {
		conds = append(conds, fmt.Sprintf("h.status_id = $%d", argPos))
		args = append(args, p.StatusID)
		argPos++
	}

	if v := strings.TrimSpace(p.StartFrom); v != "" {
		conds = append(conds, fmt.Sprintf("h.start_date::date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if v := strings.TrimSpace(p.StartTo); v != "" {
		conds = append(conds, fmt.Sprintf("h.start_date::date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if v := strings.TrimSpace(p.EndFrom); v != "" {
		conds = append(conds, fmt.Sprintf("h.end_date::date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if v := strings.TrimSpace(p.EndTo); v != "" {
		conds = append(conds, fmt.Sprintf("h.end_date::date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if p.IsCurrent != nil {
		if *p.IsCurrent {
			conds = append(conds, "h.end_date IS NULL")
		} else {
			conds = append(conds, "h.end_date IS NOT NULL")
		}
	}

	whereSQL := "TRUE"
	if len(conds) > 0 {
		whereSQL = strings.Join(conds, " AND ")
	}

	countQ := `
		SELECT COUNT(*)
		FROM vehicle_status_history h
		INNER JOIN vehicles v ON v.id = h.vehicle_id
		INNER JOIN vehicle_status vs ON vs.id = h.status_id
		WHERE ` + whereSQL + `;
	`

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list vehicle status history count: %w", err)
	}

	sortBy := normalizeVehicleStatusHistorySortBy(p.SortBy)
	order := normalizeVehicleStatusHistoryOrder(p.Order)

	listQ := fmt.Sprintf(`
		SELECT
			h.id,
			h.vehicle_id,
			v.state_number,
			v.brand_model,
			h.status_id,
			vs.name AS status_name,
			h.start_date,
			h.end_date,
			h.created_at,
			h.updated_at
		FROM vehicle_status_history h
		INNER JOIN vehicles v ON v.id = h.vehicle_id
		INNER JOIN vehicle_status vs ON vs.id = h.status_id
		WHERE %s
		ORDER BY %s %s, h.id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list vehicle status history: %w", err)
	}
	defer rows.Close()

	items := make([]models.VehicleStatusHistory, 0, limit)
	for rows.Next() {
		item, err := scanVehicleStatusHistoryRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan vehicle status history: %w", err)
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows vehicle status history: %w", err)
	}

	return items, total, nil
}

type vehicleStatusHistoryScanner interface {
	Scan(dest ...any) error
}

func scanVehicleStatusHistory(scanner vehicleStatusHistoryScanner) (*models.VehicleStatusHistory, error) {
	var item models.VehicleStatusHistory
	var endDate sql.NullTime

	if err := scanner.Scan(
		&item.ID,
		&item.VehicleID,
		&item.VehicleStateNumber,
		&item.VehicleBrandModel,
		&item.StatusID,
		&item.StatusName,
		&item.StartDate,
		&endDate,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if endDate.Valid {
		v := endDate.Time
		item.EndDate = &v
	}

	return &item, nil
}

func scanVehicleStatusHistoryRows(rows *sql.Rows) (*models.VehicleStatusHistory, error) {
	return scanVehicleStatusHistory(rows)
}

func normalizeVehicleStatusHistorySortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "id":
		return "h.id"
	case "vehicle_id":
		return "h.vehicle_id"
	case "status_id":
		return "h.status_id"
	case "status_name":
		return "vs.name"
	case "end_date":
		return "h.end_date"
	case "created_at":
		return "h.created_at"
	case "updated_at":
		return "h.updated_at"
	default:
		return "h.start_date"
	}
}

func normalizeVehicleStatusHistoryOrder(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "ASC"
	}
	return "DESC"
}
