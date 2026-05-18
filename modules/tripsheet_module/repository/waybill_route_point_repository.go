package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"auto_park/modules/tripsheet_module/models"
)

type WaybillRoutePointRepository interface {
	List(ctx context.Context, waybillID int64) ([]models.WaybillRoutePoint, error)
	Create(ctx context.Context, waybillID int64, p CreateWaybillRoutePointParams) (int64, error)
	Update(ctx context.Context, waybillID int64, routeID int64, p UpdateWaybillRoutePointParams) (bool, error)
	Delete(ctx context.Context, waybillID int64, routeID int64) (bool, error)
}

type CreateWaybillRoutePointParams struct {
	SeqNumber           int
	Destination         string
	ArrivalTime         *string
	HospitalizationTime *string
	LPUArrivalTime      *string
	ReleaseTime         *string
}

type UpdateWaybillRoutePointParams struct {
	SeqNumber           *int
	Destination         *string
	ArrivalTime         *string
	HospitalizationTime *string
	LPUArrivalTime      *string
	ReleaseTime         *string
}

type waybillRoutePointRepo struct {
	db *sql.DB
}

func NewWaybillRoutePointRepository(db *sql.DB) WaybillRoutePointRepository {
	return &waybillRoutePointRepo{db: db}
}

func (r *waybillRoutePointRepo) List(ctx context.Context, waybillID int64) ([]models.WaybillRoutePoint, error) {
	const q = `
		SELECT
			id,
			waybill_id,
			seq_number,
			destination,
			arrival_time::text,
			hospitalization_time::text,
			lpu_arrival_time::text,
			release_time::text,
			created_at
		FROM waybill_route_points
		WHERE waybill_id = $1
		ORDER BY seq_number ASC, id ASC;
	`

	rows, err := r.db.QueryContext(ctx, q, waybillID)
	if err != nil {
		return nil, fmt.Errorf("list waybill route points: %w", err)
	}
	defer rows.Close()

	items := make([]models.WaybillRoutePoint, 0)
	for rows.Next() {
		item, err := scanWaybillRoutePoint(rows)
		if err != nil {
			return nil, fmt.Errorf("list waybill route points scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list waybill route points rows: %w", err)
	}
	return items, nil
}

func (r *waybillRoutePointRepo) Create(ctx context.Context, waybillID int64, p CreateWaybillRoutePointParams) (int64, error) {
	const q = `
		INSERT INTO waybill_route_points (
			waybill_id,
			seq_number,
			destination,
			arrival_time,
			hospitalization_time,
			lpu_arrival_time,
			release_time,
			created_at
		) VALUES ($1, $2, $3, $4::time, $5::time, $6::time, $7::time, NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(
		ctx,
		q,
		waybillID,
		p.SeqNumber,
		p.Destination,
		nullableTimeString(p.ArrivalTime),
		nullableTimeString(p.HospitalizationTime),
		nullableTimeString(p.LPUArrivalTime),
		nullableTimeString(p.ReleaseTime),
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("create waybill route point: %w", err)
	}
	return id, nil
}

func (r *waybillRoutePointRepo) Update(ctx context.Context, waybillID int64, routeID int64, p UpdateWaybillRoutePointParams) (bool, error) {
	setParts := make([]string, 0, 6)
	args := make([]any, 0, 8)
	argPos := 1

	add := func(expr string, value any) {
		setParts = append(setParts, fmt.Sprintf(expr, argPos))
		args = append(args, value)
		argPos++
	}

	if p.SeqNumber != nil {
		add("seq_number = $%d", *p.SeqNumber)
	}
	if p.Destination != nil {
		add("destination = $%d", strings.TrimSpace(*p.Destination))
	}
	if p.ArrivalTime != nil {
		add("arrival_time = $%d::time", nullableTimeString(p.ArrivalTime))
	}
	if p.HospitalizationTime != nil {
		add("hospitalization_time = $%d::time", nullableTimeString(p.HospitalizationTime))
	}
	if p.LPUArrivalTime != nil {
		add("lpu_arrival_time = $%d::time", nullableTimeString(p.LPUArrivalTime))
	}
	if p.ReleaseTime != nil {
		add("release_time = $%d::time", nullableTimeString(p.ReleaseTime))
	}

	if len(setParts) == 0 {
		return r.exists(ctx, waybillID, routeID)
	}

	args = append(args, routeID, waybillID)
	q := fmt.Sprintf(`
		UPDATE waybill_route_points
		SET %s
		WHERE id = $%d
		  AND waybill_id = $%d;
	`, strings.Join(setParts, ", "), argPos, argPos+1)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return false, fmt.Errorf("update waybill route point: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update waybill route point rows affected: %w", err)
	}
	return affected > 0, nil
}

func (r *waybillRoutePointRepo) Delete(ctx context.Context, waybillID int64, routeID int64) (bool, error) {
	const q = `
		DELETE FROM waybill_route_points
		WHERE id = $1
		  AND waybill_id = $2;
	`

	res, err := r.db.ExecContext(ctx, q, routeID, waybillID)
	if err != nil {
		return false, fmt.Errorf("delete waybill route point: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete waybill route point rows affected: %w", err)
	}
	return affected > 0, nil
}

func (r *waybillRoutePointRepo) exists(ctx context.Context, waybillID int64, routeID int64) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM waybill_route_points WHERE id = $1 AND waybill_id = $2);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, routeID, waybillID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check waybill route point exists: %w", err)
	}
	return exists, nil
}

type routePointScanner interface {
	Scan(dest ...any) error
}

func scanWaybillRoutePoint(scanner routePointScanner) (*models.WaybillRoutePoint, error) {
	var item models.WaybillRoutePoint
	var arrival, hospitalization, lpuArrival, release sql.NullString
	if err := scanner.Scan(
		&item.ID,
		&item.WaybillID,
		&item.SeqNumber,
		&item.Destination,
		&arrival,
		&hospitalization,
		&lpuArrival,
		&release,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.ArrivalTime = nullableStringPtr(arrival)
	item.HospitalizationTime = nullableStringPtr(hospitalization)
	item.LPUArrivalTime = nullableStringPtr(lpuArrival)
	item.ReleaseTime = nullableStringPtr(release)
	return &item, nil
}

func nullableTimeString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
