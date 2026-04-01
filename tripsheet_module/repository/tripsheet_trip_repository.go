package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"auto_park/tripsheet_module/dto"
	"auto_park/tripsheet_module/models"
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

	var (
		item      dto.TripsheetTripResponse
		startTime sql.NullTime
		endTime   sql.NullTime
		createdAt time.Time
		updatedAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.TripsheetID,
		&item.RouteDescription,
		&startTime,
		&endTime,
		&item.DistancePassed,
		&item.StatusID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
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

	return &item, nil
}

func (r *tripsheetTripRepo) GetAll(ctx context.Context, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	baseQuery := fmt.Sprintf(`
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
		WHERE 1=1
	`, r.table())

	query, args := r.applyTripsheetTripFilters(baseQuery, filter)
	query += " ORDER BY id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get all tripsheet trips: %w", err)
	}
	defer rows.Close()

	items := make([]dto.TripsheetTripResponse, 0)

	for rows.Next() {
		var (
			item      dto.TripsheetTripResponse
			startTime sql.NullTime
			endTime   sql.NullTime
			createdAt time.Time
			updatedAt time.Time
		)

		err := rows.Scan(
			&item.ID,
			&item.TripsheetID,
			&item.RouteDescription,
			&startTime,
			&endTime,
			&item.DistancePassed,
			&item.StatusID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan tripsheet trip row: %w", err)
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

	return items, len(items), nil
}

func (r *tripsheetTripRepo) GetAllByTripsheetID(ctx context.Context, tripsheetID int64, filter dto.TripsheetTripFilter) ([]dto.TripsheetTripResponse, int, error) {
	baseQuery := fmt.Sprintf(`
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
		WHERE tripsheet_id = $1
	`, r.table())

	args := []any{tripsheetID}
	query, extraArgs := r.applyTripsheetTripFiltersWithOffset(baseQuery, filter, 2)
	args = append(args, extraArgs...)
	query += " ORDER BY id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get all tripsheet trips by tripsheet_id: %w", err)
	}
	defer rows.Close()

	items := make([]dto.TripsheetTripResponse, 0)

	for rows.Next() {
		var (
			item      dto.TripsheetTripResponse
			startTime sql.NullTime
			endTime   sql.NullTime
			createdAt time.Time
			updatedAt time.Time
		)

		err := rows.Scan(
			&item.ID,
			&item.TripsheetID,
			&item.RouteDescription,
			&startTime,
			&endTime,
			&item.DistancePassed,
			&item.StatusID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan tripsheet trip row by tripsheet_id: %w", err)
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

	return items, len(items), nil
}

func (r *tripsheetTripRepo) applyTripsheetTripFilters(baseQuery string, filter dto.TripsheetTripFilter) (string, []any) {
	return r.applyTripsheetTripFiltersWithOffset(baseQuery, filter, 1)
}

func (r *tripsheetTripRepo) applyTripsheetTripFiltersWithOffset(baseQuery string, filter dto.TripsheetTripFilter, startIndex int) (string, []any) {
	query := baseQuery
	args := make([]any, 0)
	i := startIndex

	if filter.TripsheetID != nil && startIndex == 1 {
		query += fmt.Sprintf(" AND tripsheet_id = $%d", i)
		args = append(args, *filter.TripsheetID)
		i++
	}

	if filter.StatusID != nil {
		query += fmt.Sprintf(" AND status_id = $%d", i)
		args = append(args, *filter.StatusID)
		i++
	}

	if filter.DateFrom != nil && strings.TrimSpace(*filter.DateFrom) != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", i)
		args = append(args, strings.TrimSpace(*filter.DateFrom))
		i++
	}

	if filter.DateTo != nil && strings.TrimSpace(*filter.DateTo) != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", i)
		args = append(args, strings.TrimSpace(*filter.DateTo))
		i++
	}

	return query, args
}
