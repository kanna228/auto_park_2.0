package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"auto_park/modules/audit_log_module/dto"
)

type CreateAuditLogParams struct {
	Level      string
	Function   string
	FromStatus *string
	ToStatus   *string
	Actor      string
	Message    *string
}

type AuditLogRepository interface {
	Create(ctx context.Context, p CreateAuditLogParams) error
	List(ctx context.Context, q dto.AuditLogListQuery) ([]dto.AuditLogResponse, int64, error)
}

type auditLogRepo struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) AuditLogRepository {
	return &auditLogRepo{db: db}
}

func (r *auditLogRepo) Create(ctx context.Context, p CreateAuditLogParams) error {
	const q = `
		INSERT INTO audit_logs (level, function, from_status, to_status, actor, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW());
	`
	if _, err := r.db.ExecContext(ctx, q, p.Level, p.Function, p.FromStatus, p.ToStatus, p.Actor, p.Message); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *auditLogRepo) List(ctx context.Context, q dto.AuditLogListQuery) ([]dto.AuditLogResponse, int64, error) {
	limit := normalizeLimit(q.Limit)
	offset := normalizeOffset(q.Offset)

	conds := []string{"1=1"}
	args := make([]any, 0, 8)
	argPos := 1

	if v := strings.TrimSpace(q.Function); v != "" {
		conds = append(conds, fmt.Sprintf("function = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(q.Level); v != "" {
		conds = append(conds, fmt.Sprintf("level = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(q.Date); v != "" {
		conds = append(conds, fmt.Sprintf("created_at::date = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(q.Search); v != "" {
		conds = append(conds, fmt.Sprintf("(actor ILIKE $%d OR COALESCE(from_status, '') ILIKE $%d OR COALESCE(to_status, '') ILIKE $%d OR COALESCE(message, '') ILIKE $%d)", argPos, argPos, argPos, argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")
	countQ := `SELECT COUNT(*) FROM audit_logs` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list audit logs count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT id, level, function, from_status, to_status, actor, message, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	items := make([]dto.AuditLogResponse, 0)
	for rows.Next() {
		item, err := scanAuditLog(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list audit logs scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list audit logs rows: %w", err)
	}
	return items, total, nil
}

type auditLogScanner interface {
	Scan(dest ...any) error
}

func scanAuditLog(scanner auditLogScanner) (*dto.AuditLogResponse, error) {
	var item dto.AuditLogResponse
	var fromStatus, toStatus, message sql.NullString
	if err := scanner.Scan(
		&item.ID,
		&item.Level,
		&item.Function,
		&fromStatus,
		&toStatus,
		&item.Actor,
		&message,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.FromStatus = nullableStringPtr(fromStatus)
	item.ToStatus = nullableStringPtr(toStatus)
	item.Message = nullableStringPtr(message)
	return &item, nil
}

func nullableStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
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
