package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/maintenance_execution_module/dto"
)

var ErrMaintenanceScheduleNotFound = errors.New("maintenance schedule not found")

type CreateMaintenanceExecutionParams struct {
	ScheduleID          int64
	Board               string
	Action              string
	Comment             *string
	ResponsibleMechanic *string
}

type MaintenanceExecutionRepository interface {
	Create(ctx context.Context, p CreateMaintenanceExecutionParams) (*dto.MaintenanceExecutionResponse, error)
	BulkCreate(ctx context.Context, items []CreateMaintenanceExecutionParams) ([]dto.MaintenanceExecutionResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.MaintenanceExecutionResponse, error)
	ListBySchedule(ctx context.Context, scheduleID int64, limit int, offset int) ([]dto.MaintenanceExecutionResponse, int64, error)
	Update(ctx context.Context, id int64, p CreateMaintenanceExecutionParams) (*dto.MaintenanceExecutionResponse, error)
	Delete(ctx context.Context, id int64) (bool, error)
}

type maintenanceExecutionRepo struct {
	db *sql.DB
}

func NewMaintenanceExecutionRepository(db *sql.DB) MaintenanceExecutionRepository {
	return &maintenanceExecutionRepo{db: db}
}

func (r *maintenanceExecutionRepo) Create(ctx context.Context, p CreateMaintenanceExecutionParams) (*dto.MaintenanceExecutionResponse, error) {
	item, err := insertMaintenanceExecution(ctx, r.db, p)
	if err != nil {
		return nil, mapMaintenanceExecutionError(err)
	}
	return item, nil
}

func (r *maintenanceExecutionRepo) BulkCreate(ctx context.Context, items []CreateMaintenanceExecutionParams) ([]dto.MaintenanceExecutionResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin bulk create maintenance executions: %w", err)
	}
	defer rollbackTx(tx)

	out := make([]dto.MaintenanceExecutionResponse, 0, len(items))
	for _, p := range items {
		item, err := insertMaintenanceExecution(ctx, tx, p)
		if err != nil {
			return nil, mapMaintenanceExecutionError(err)
		}
		out = append(out, *item)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit bulk create maintenance executions: %w", err)
	}
	return out, nil
}

func (r *maintenanceExecutionRepo) GetByID(ctx context.Context, id int64) (*dto.MaintenanceExecutionResponse, error) {
	const q = `
		SELECT id, schedule_id, board, action, comment, responsible_mechanic, created_at, updated_at
		FROM maintenance_executions
		WHERE id = $1;
	`
	item, err := scanMaintenanceExecution(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get maintenance execution by id: %w", err)
	}
	return item, nil
}

func (r *maintenanceExecutionRepo) ListBySchedule(ctx context.Context, scheduleID int64, limit int, offset int) ([]dto.MaintenanceExecutionResponse, int64, error) {
	limit = normalizeLimit(limit)
	offset = normalizeOffset(offset)

	const countQ = `SELECT COUNT(*) FROM maintenance_executions WHERE schedule_id = $1;`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, scheduleID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list maintenance executions count: %w", err)
	}

	const q = `
		SELECT id, schedule_id, board, action, comment, responsible_mechanic, created_at, updated_at
		FROM maintenance_executions
		WHERE schedule_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.db.QueryContext(ctx, q, scheduleID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list maintenance executions: %w", err)
	}
	defer rows.Close()

	items := make([]dto.MaintenanceExecutionResponse, 0)
	for rows.Next() {
		item, err := scanMaintenanceExecution(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list maintenance executions scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list maintenance executions rows: %w", err)
	}
	return items, total, nil
}

func (r *maintenanceExecutionRepo) Update(ctx context.Context, id int64, p CreateMaintenanceExecutionParams) (*dto.MaintenanceExecutionResponse, error) {
	const q = `
		UPDATE maintenance_executions
		SET schedule_id = $1,
			board = $2,
			action = $3,
			comment = $4,
			responsible_mechanic = $5,
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, schedule_id, board, action, comment, responsible_mechanic, created_at, updated_at;
	`
	item, err := scanMaintenanceExecution(r.db.QueryRowContext(ctx, q, p.ScheduleID, p.Board, p.Action, p.Comment, p.ResponsibleMechanic, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, mapMaintenanceExecutionError(err)
	}
	return item, nil
}

func (r *maintenanceExecutionRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM maintenance_executions WHERE id = $1;`, id)
	if err != nil {
		return false, fmt.Errorf("delete maintenance execution: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete maintenance execution rows affected: %w", err)
	}
	return affected > 0, nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func insertMaintenanceExecution(ctx context.Context, db queryRower, p CreateMaintenanceExecutionParams) (*dto.MaintenanceExecutionResponse, error) {
	const q = `
		INSERT INTO maintenance_executions (schedule_id, board, action, comment, responsible_mechanic, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, schedule_id, board, action, comment, responsible_mechanic, created_at, updated_at;
	`
	return scanMaintenanceExecution(db.QueryRowContext(ctx, q, p.ScheduleID, p.Board, p.Action, p.Comment, p.ResponsibleMechanic))
}

type maintenanceExecutionScanner interface {
	Scan(dest ...any) error
}

func scanMaintenanceExecution(scanner maintenanceExecutionScanner) (*dto.MaintenanceExecutionResponse, error) {
	var item dto.MaintenanceExecutionResponse
	var comment, responsibleMechanic sql.NullString
	if err := scanner.Scan(
		&item.ID,
		&item.ScheduleID,
		&item.Board,
		&item.Action,
		&comment,
		&responsibleMechanic,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.Comment = nullableStringPtr(comment)
	item.ResponsibleMechanic = nullableStringPtr(responsibleMechanic)
	return &item, nil
}

func nullableStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func rollbackTx(tx *sql.Tx) {
	_ = tx.Rollback()
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 10000 {
		return 10000
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func mapMaintenanceExecutionError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "maintenance_executions_schedule_id_fkey") {
		return ErrMaintenanceScheduleNotFound
	}
	return err
}
