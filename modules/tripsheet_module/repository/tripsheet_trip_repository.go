package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/tripsheet_module/dto"
	"auto_park/modules/tripsheet_module/models"
)

type TripsheetTripRepo interface {
	Create(ctx context.Context, input models.CreateTripsheetTripInput) (*models.TripsheetTrip, error)
	Update(ctx context.Context, input models.UpdateTripsheetTripInput) (*models.TripsheetTrip, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*dto.TripsheetTripResponse, error)
	GetAll(ctx context.Context, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error)
	GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error)
}

type tripsheetTripRepo struct {
	db     *sql.DB
	schema string
}

func NewTripsheetTripRepo(db *sql.DB, schema string) TripsheetTripRepo {
	return &tripsheetTripRepo{db: db, schema: schema}
}

func (r *tripsheetTripRepo) table() string { return "tripsheet_trips" }

func (r *tripsheetTripRepo) Create(ctx context.Context, input models.CreateTripsheetTripInput) (*models.TripsheetTrip, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			tripsheet_id,
			route_description,
			start_time,
			end_time,
			distance_passed,
			status_id
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			tripsheet_id,
			route_description,
			start_time,
			end_time,
			distance_passed,
			status_id,
			created_at,
			updated_at
	`, r.table())

	var item models.TripsheetTrip

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.TripsheetID,
		input.RouteDescription,
		input.StartTime,
		input.EndTime,
		input.DistancePassed,
		input.StatusID,
	).Scan(
		&item.ID,
		&item.TripsheetID,
		&item.RouteDescription,
		&item.StartTime,
		&item.EndTime,
		&item.DistancePassed,
		&item.StatusID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create tripsheet trip: %w", err)
	}

	return &item, nil
}

func (r *tripsheetTripRepo) Update(ctx context.Context, input models.UpdateTripsheetTripInput) (*models.TripsheetTrip, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET
			tripsheet_id = $1,
			route_description = $2,
			start_time = $3,
			end_time = $4,
			distance_passed = $5,
			status_id = $6,
			updated_at = NOW()
		WHERE id = $7
		RETURNING
			id,
			tripsheet_id,
			route_description,
			start_time,
			end_time,
			distance_passed,
			status_id,
			created_at,
			updated_at
	`, r.table())

	var item models.TripsheetTrip

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.TripsheetID,
		input.RouteDescription,
		input.StartTime,
		input.EndTime,
		input.DistancePassed,
		input.StatusID,
		input.ID,
	).Scan(
		&item.ID,
		&item.TripsheetID,
		&item.RouteDescription,
		&item.StartTime,
		&item.EndTime,
		&item.DistancePassed,
		&item.StatusID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("update tripsheet trip: %w", err)
	}

	return &item, nil
}

func (r *tripsheetTripRepo) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id = $1
	`, r.table())

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete tripsheet trip: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected after delete tripsheet trip: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *tripsheetTripRepo) GetByID(ctx context.Context, id int64) (*dto.TripsheetTripResponse, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			tripsheet_id,
			route_description,
			start_time,
			end_time,
			distance_passed,
			status_id,
			created_at,
			updated_at
		FROM %s
		WHERE id = $1
	`, r.table())

	return r.scanOneTrip(ctx, query, id)
}

func (r *tripsheetTripRepo) GetAll(ctx context.Context, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	whereSQL, args := r.buildTripsheetTripWhere(filter, 1, false)
	return r.listTripsheetTrips(ctx, whereSQL, args, filter)
}

func (r *tripsheetTripRepo) GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	filter.TripsheetID = &tripsheetID
	whereSQL, args := r.buildTripsheetTripWhere(filter, 1, true)
	return r.listTripsheetTrips(ctx, whereSQL, args, filter)
}

func (r *tripsheetTripRepo) listTripsheetTrips(ctx context.Context, whereSQL string, args []any, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s tt
		LEFT JOIN tripsheets t ON t.id = tt.tripsheet_id
		WHERE %s;
	`, r.table(), whereSQL)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tripsheet trips: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := normalizeTripsheetTripSortBy(filter.SortBy)
	order := normalizeTripsheetTripOrder(filter.Order)
	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT
			tt.id,
			tt.tripsheet_id,
			tt.route_description,
			tt.start_time,
			tt.end_time,
			tt.distance_passed,
			tt.status_id,
			tt.created_at,
			tt.updated_at
		FROM %s tt
		LEFT JOIN tripsheets t ON t.id = tt.tripsheet_id
		WHERE %s
		ORDER BY %s %s, tt.id DESC
		LIMIT $%d OFFSET $%d;
	`, r.table(), whereSQL, sortBy, order, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get all tripsheet trips: %w", err)
	}
	defer rows.Close()

	items, err := scanTripRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *tripsheetTripRepo) buildTripsheetTripWhere(filter dto.TripsheetTripFilter, startIndex int, forceTripsheet bool) (string, []any) {
	where := []string{"1=1"}
	args := make([]any, 0, 8)
	i := startIndex

	if filter.TripsheetID != nil && (forceTripsheet || startIndex == 1) {
		where = append(where, fmt.Sprintf("tt.tripsheet_id = $%d", i))
		args = append(args, *filter.TripsheetID)
		i++
	}
	if filter.VehicleID != nil {
		where = append(where, fmt.Sprintf("t.vehicle_id = $%d", i))
		args = append(args, *filter.VehicleID)
		i++
	}
	if filter.DriverID != nil {
		where = append(where, fmt.Sprintf("t.driver_id = $%d", i))
		args = append(args, *filter.DriverID)
		i++
	}
	if filter.StatusID != nil {
		where = append(where, fmt.Sprintf("tt.status_id = $%d", i))
		args = append(args, *filter.StatusID)
		i++
	}
	if filter.DateFrom != nil && strings.TrimSpace(*filter.DateFrom) != "" {
		where = append(where, fmt.Sprintf("tt.created_at::date >= $%d", i))
		args = append(args, strings.TrimSpace(*filter.DateFrom))
		i++
	}
	if filter.DateTo != nil && strings.TrimSpace(*filter.DateTo) != "" {
		where = append(where, fmt.Sprintf("tt.created_at::date <= $%d", i))
		args = append(args, strings.TrimSpace(*filter.DateTo))
		i++
	}

	return strings.Join(where, " AND "), args
}

func (r *tripsheetTripRepo) scanOneTrip(ctx context.Context, query string, args ...any) (*dto.TripsheetTripResponse, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := scanTripRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return &items[0], nil
}

func scanTripRows(rows *sql.Rows) ([]dto.TripsheetTripResponse, error) {
	items := make([]dto.TripsheetTripResponse, 0)
	for rows.Next() {
		var (
			item      dto.TripsheetTripResponse
			startTime sql.NullTime
			endTime   sql.NullTime
			createdAt time.Time
			updatedAt time.Time
		)

		if err := rows.Scan(
			&item.ID,
			&item.TripsheetID,
			&item.RouteDescription,
			&startTime,
			&endTime,
			&item.DistancePassed,
			&item.StatusID,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tripsheet trip row: %w", err)
		}

		if startTime.Valid {
			v := startTime.Time.Format(time.RFC3339)
			item.StartTime = &v
		}
		if endTime.Valid {
			v := endTime.Time.Format(time.RFC3339)
			item.EndTime = &v
		}

		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tripsheet trip rows: %w", err)
	}
	return items, nil
}

func normalizeTripsheetTripSortBy(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "tripsheet_id":
		return "tt.tripsheet_id"
	case "vehicle_id":
		return "t.vehicle_id"
	case "driver_id":
		return "t.driver_id"
	case "status_id":
		return "tt.status_id"
	case "start_time":
		return "tt.start_time"
	case "end_time":
		return "tt.end_time"
	case "created_at":
		return "tt.created_at"
	case "updated_at":
		return "tt.updated_at"
	default:
		return "tt.id"
	}
}

func normalizeTripsheetTripOrder(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "asc") {
		return "ASC"
	}
	return "DESC"
}
