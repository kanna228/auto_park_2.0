package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/maintenance_schedule_module/dto"
)

type MaintenanceScheduleRepository interface {
	Create(ctx context.Context, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.MaintenanceScheduleResponse, error)
	List(ctx context.Context, q dto.MaintenanceScheduleListQuery) ([]dto.MaintenanceScheduleResponse, int64, error)
	Update(ctx context.Context, id int64, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type maintenanceScheduleRepo struct {
	db *sql.DB
}

func NewMaintenanceScheduleRepository(db *sql.DB) MaintenanceScheduleRepository {
	return &maintenanceScheduleRepo{db: db}
}

func (r *maintenanceScheduleRepo) Create(ctx context.Context, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error) {
	categories, boards, mechanics, err := marshalScheduleArrays(req)
	if err != nil {
		return nil, err
	}

	const q = `
		INSERT INTO maintenance_schedules (
			is_draft,
			date_from,
			date_to,
			consecutive_count,
			consecutive_unit,
			every_count,
			every_unit,
			time_from,
			time_to,
			duration_value,
			duration_unit,
			limit_boards_by_time,
			categories,
			boards,
			mechanics,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb, $15::jsonb, NOW(), NOW())
		RETURNING id, is_draft, date_from, date_to, consecutive_count, consecutive_unit, every_count, every_unit, time_from, time_to, duration_value, duration_unit, limit_boards_by_time, categories, boards, mechanics, created_at, updated_at;
	`
	item, err := scanMaintenanceSchedule(r.db.QueryRowContext(
		ctx,
		q,
		req.IsDraft,
		req.DateFrom,
		req.DateTo,
		req.ConsecutiveCount,
		req.ConsecutiveUnit,
		req.EveryCount,
		req.EveryUnit,
		req.TimeFrom,
		req.TimeTo,
		req.DurationValue,
		req.DurationUnit,
		req.LimitBoardsByTime,
		string(categories),
		string(boards),
		string(mechanics),
	))
	if err != nil {
		return nil, fmt.Errorf("create maintenance schedule: %w", err)
	}
	return item, nil
}

func (r *maintenanceScheduleRepo) GetByID(ctx context.Context, id int64) (*dto.MaintenanceScheduleResponse, error) {
	const q = `
		SELECT id, is_draft, date_from, date_to, consecutive_count, consecutive_unit, every_count, every_unit, time_from, time_to, duration_value, duration_unit, limit_boards_by_time, categories, boards, mechanics, created_at, updated_at
		FROM maintenance_schedules
		WHERE id = $1;
	`
	item, err := scanMaintenanceSchedule(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get maintenance schedule by id: %w", err)
	}
	return item, nil
}

func (r *maintenanceScheduleRepo) List(ctx context.Context, q dto.MaintenanceScheduleListQuery) ([]dto.MaintenanceScheduleResponse, int64, error) {
	limit := normalizeLimit(q.Limit)
	offset := normalizeOffset(q.Offset)

	conds := []string{"1=1"}
	args := make([]any, 0, 4)
	argPos := 1
	if q.IsDraft != nil {
		conds = append(conds, fmt.Sprintf("is_draft = $%d", argPos))
		args = append(args, *q.IsDraft)
		argPos++
	}
	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `SELECT COUNT(*) FROM maintenance_schedules` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list maintenance schedules count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT id, is_draft, date_from, date_to, consecutive_count, consecutive_unit, every_count, every_unit, time_from, time_to, duration_value, duration_unit, limit_boards_by_time, categories, boards, mechanics, created_at, updated_at
		FROM maintenance_schedules
		%s
		ORDER BY date_from DESC, id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list maintenance schedules: %w", err)
	}
	defer rows.Close()

	items := make([]dto.MaintenanceScheduleResponse, 0)
	for rows.Next() {
		item, err := scanMaintenanceSchedule(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list maintenance schedules scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list maintenance schedules rows: %w", err)
	}
	return items, total, nil
}

func (r *maintenanceScheduleRepo) Update(ctx context.Context, id int64, req dto.MaintenanceScheduleRequest) (*dto.MaintenanceScheduleResponse, error) {
	categories, boards, mechanics, err := marshalScheduleArrays(req)
	if err != nil {
		return nil, err
	}

	const q = `
		UPDATE maintenance_schedules
		SET is_draft = $1,
			date_from = $2,
			date_to = $3,
			consecutive_count = $4,
			consecutive_unit = $5,
			every_count = $6,
			every_unit = $7,
			time_from = $8,
			time_to = $9,
			duration_value = $10,
			duration_unit = $11,
			limit_boards_by_time = $12,
			categories = $13::jsonb,
			boards = $14::jsonb,
			mechanics = $15::jsonb,
			updated_at = NOW()
		WHERE id = $16
		RETURNING id, is_draft, date_from, date_to, consecutive_count, consecutive_unit, every_count, every_unit, time_from, time_to, duration_value, duration_unit, limit_boards_by_time, categories, boards, mechanics, created_at, updated_at;
	`
	item, err := scanMaintenanceSchedule(r.db.QueryRowContext(
		ctx,
		q,
		req.IsDraft,
		req.DateFrom,
		req.DateTo,
		req.ConsecutiveCount,
		req.ConsecutiveUnit,
		req.EveryCount,
		req.EveryUnit,
		req.TimeFrom,
		req.TimeTo,
		req.DurationValue,
		req.DurationUnit,
		req.LimitBoardsByTime,
		string(categories),
		string(boards),
		string(mechanics),
		id,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update maintenance schedule: %w", err)
	}
	return item, nil
}

func (r *maintenanceScheduleRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM maintenance_schedules WHERE id = $1;`, id)
	if err != nil {
		return false, fmt.Errorf("delete maintenance schedule: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete maintenance schedule rows affected: %w", err)
	}
	return affected > 0, nil
}

type maintenanceScheduleScanner interface {
	Scan(dest ...any) error
}

func scanMaintenanceSchedule(scanner maintenanceScheduleScanner) (*dto.MaintenanceScheduleResponse, error) {
	var item dto.MaintenanceScheduleResponse
	var dateFrom, dateTo time.Time
	var categoriesRaw, boardsRaw, mechanicsRaw []byte
	if err := scanner.Scan(
		&item.ID,
		&item.IsDraft,
		&dateFrom,
		&dateTo,
		&item.ConsecutiveCount,
		&item.ConsecutiveUnit,
		&item.EveryCount,
		&item.EveryUnit,
		&item.TimeFrom,
		&item.TimeTo,
		&item.DurationValue,
		&item.DurationUnit,
		&item.LimitBoardsByTime,
		&categoriesRaw,
		&boardsRaw,
		&mechanicsRaw,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.DateFrom = dateFrom.Format("2006-01-02")
	item.DateTo = dateTo.Format("2006-01-02")
	item.Categories = unmarshalStringSlice(categoriesRaw)
	item.Boards = unmarshalStringSlice(boardsRaw)
	item.Mechanics = unmarshalStringSlice(mechanicsRaw)
	return &item, nil
}

func marshalScheduleArrays(req dto.MaintenanceScheduleRequest) ([]byte, []byte, []byte, error) {
	categories, err := json.Marshal(normalizeStringSlice(req.Categories))
	if err != nil {
		return nil, nil, nil, err
	}
	boards, err := json.Marshal(normalizeStringSlice(req.Boards))
	if err != nil {
		return nil, nil, nil, err
	}
	mechanics, err := json.Marshal(normalizeStringSlice(req.Mechanics))
	if err != nil {
		return nil, nil, nil, err
	}
	return categories, boards, mechanics, nil
}

func normalizeStringSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func unmarshalStringSlice(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	if out == nil {
		return []string{}
	}
	return out
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 200 {
		return 50
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
