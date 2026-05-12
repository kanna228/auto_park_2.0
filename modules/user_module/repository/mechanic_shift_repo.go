package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/user_module/models"
)

var (
	ErrMechanicShiftNotFound        = errors.New("mechanic shift not found")
	ErrMechanicShiftUserNotFound    = errors.New("mechanic shift user not found")
	ErrMechanicShiftUserNotMechanic = errors.New("user is not duty mechanic")
)

type MechanicShiftRepo struct {
	db *sql.DB
}

func NewMechanicShiftRepo(db *sql.DB) *MechanicShiftRepo {
	return &MechanicShiftRepo{db: db}
}

type CreateMechanicShiftParams struct {
	UserID    int64
	ShiftDate string
	TimeFrom  string
	TimeTo    *string
	Comment   *string
	IsActive  bool
}

type UpdateMechanicShiftParams struct {
	UserID    *int64
	ShiftDate *string
	TimeFrom  *string
	TimeTo    *string
	Comment   *string
	IsActive  *bool
}

type ListMechanicShiftsParams struct {
	UserID   int64
	DateFrom string
	DateTo   string
	IsActive *bool
	Limit    int
	Offset   int
	SortBy   string
	Order    string
}

type MechanicShiftRepository interface {
	RefreshActivity(ctx context.Context) error
	Create(ctx context.Context, p CreateMechanicShiftParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.MechanicShift, error)
	List(ctx context.Context, p ListMechanicShiftsParams) ([]models.MechanicShift, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateMechanicShiftParams) (bool, error)
	UpdateActivityByID(ctx context.Context, id int64, isActive bool) (bool, error)
	SoftDeleteByID(ctx context.Context, id int64, deletedByUserID int64) (bool, error)
	UserIsDutyMechanic(ctx context.Context, userID int64) (bool, error)
}

func (r *MechanicShiftRepo) RefreshActivity(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `SELECT refresh_mechanic_shifts_activity();`)
	if err != nil {
		return fmt.Errorf("refresh mechanic shifts activity: %w", err)
	}

	return nil
}

func (r *MechanicShiftRepo) Create(ctx context.Context, p CreateMechanicShiftParams) (int64, error) {
	const q = `
		INSERT INTO mechanic_shifts (user_id, shift_date, time_from, time_to, comment, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.UserID, p.ShiftDate, p.TimeFrom, p.TimeTo, p.Comment, p.IsActive).Scan(&id); err != nil {
		return 0, mapMechanicShiftError(err)
	}

	return id, nil
}

func (r *MechanicShiftRepo) GetByID(ctx context.Context, id int64) (*models.MechanicShift, error) {
	const q = `
		SELECT
			ms.id,
			ms.user_id,
			u.email,
			u.first_name,
			u.last_name,
			u.middle_name,
			u.role_id,
			roles.name AS role_name,
			ms.shift_date,
			ms.time_from::text,
			ms.time_to::text,
			ms.comment,
			ms.is_active,
			ms.created_at,
			ms.updated_at
		FROM mechanic_shifts ms
		INNER JOIN users u ON u.id = ms.user_id
		INNER JOIN roles ON roles.id = u.role_id
		WHERE ms.id = $1
		  AND ms.is_deleted = FALSE;
	`

	item, err := scanMechanicShift(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMechanicShiftNotFound
		}
		return nil, fmt.Errorf("get mechanic shift by id: %w", err)
	}

	return item, nil
}

func (r *MechanicShiftRepo) List(ctx context.Context, p ListMechanicShiftsParams) ([]models.MechanicShift, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := []string{"ms.is_deleted = FALSE"}
	args := make([]any, 0, 8)
	argPos := 1

	if p.UserID > 0 {
		conds = append(conds, fmt.Sprintf("ms.user_id = $%d", argPos))
		args = append(args, p.UserID)
		argPos++
	}

	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("ms.shift_date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("ms.shift_date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if p.IsActive != nil {
		conds = append(conds, fmt.Sprintf("ms.is_active = $%d", argPos))
		args = append(args, *p.IsActive)
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `
		SELECT COUNT(*)
		FROM mechanic_shifts ms
		INNER JOIN users u ON u.id = ms.user_id
		INNER JOIN roles ON roles.id = u.role_id
	` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list mechanic shifts count: %w", err)
	}

	sortBy := normalizeMechanicShiftSortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	listQ := fmt.Sprintf(`
		SELECT
			ms.id,
			ms.user_id,
			u.email,
			u.first_name,
			u.last_name,
			u.middle_name,
			u.role_id,
			roles.name AS role_name,
			ms.shift_date,
			ms.time_from::text,
			ms.time_to::text,
			ms.comment,
			ms.is_active,
			ms.created_at,
			ms.updated_at
		FROM mechanic_shifts ms
		INNER JOIN users u ON u.id = ms.user_id
		INNER JOIN roles ON roles.id = u.role_id
		%s
		ORDER BY %s %s, ms.id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list mechanic shifts: %w", err)
	}
	defer rows.Close()

	items := make([]models.MechanicShift, 0)
	for rows.Next() {
		item, err := scanMechanicShiftRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan mechanic shift: %w", err)
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows mechanic shifts: %w", err)
	}

	return items, total, nil
}

func (r *MechanicShiftRepo) UpdateByID(ctx context.Context, id int64, p UpdateMechanicShiftParams) (bool, error) {
	setParts := make([]string, 0, 8)
	args := make([]any, 0, 10)
	argPos := 1

	if p.UserID != nil {
		setParts = append(setParts, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, *p.UserID)
		argPos++
	}
	if p.ShiftDate != nil {
		setParts = append(setParts, fmt.Sprintf("shift_date = $%d", argPos))
		args = append(args, *p.ShiftDate)
		argPos++
	}
	if p.TimeFrom != nil {
		setParts = append(setParts, fmt.Sprintf("time_from = $%d", argPos))
		args = append(args, *p.TimeFrom)
		argPos++
	}
	if p.TimeTo != nil {
		setParts = append(setParts, fmt.Sprintf("time_to = $%d", argPos))
		args = append(args, nullableStringValue(*p.TimeTo))
		argPos++
	}
	if p.Comment != nil {
		setParts = append(setParts, fmt.Sprintf("comment = $%d", argPos))
		args = append(args, nullableStringValue(*p.Comment))
		argPos++
	}
	if p.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *p.IsActive)
		argPos++
	}

	if len(setParts) == 0 {
		return true, nil
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)

	q := fmt.Sprintf(`
		UPDATE mechanic_shifts
		SET %s
		WHERE id = $%d
		  AND is_deleted = FALSE;
	`, strings.Join(setParts, ", "), argPos)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return false, mapMechanicShiftError(err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update mechanic shift rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *MechanicShiftRepo) UpdateActivityByID(ctx context.Context, id int64, isActive bool) (bool, error) {
	const q = `
		UPDATE mechanic_shifts
		SET is_active = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND is_deleted = FALSE;
	`

	res, err := r.db.ExecContext(ctx, q, isActive, id)
	if err != nil {
		return false, fmt.Errorf("update mechanic shift activity: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update mechanic shift activity rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *MechanicShiftRepo) SoftDeleteByID(ctx context.Context, id int64, deletedByUserID int64) (bool, error) {
	const q = `
		UPDATE mechanic_shifts
		SET is_deleted = TRUE,
			is_active = FALSE,
			deleted_at = NOW(),
			deleted_by_user_id = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND is_deleted = FALSE;
	`

	res, err := r.db.ExecContext(ctx, q, deletedByUserID, id)
	if err != nil {
		return false, mapMechanicShiftError(err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("soft delete mechanic shift rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *MechanicShiftRepo) UserIsDutyMechanic(ctx context.Context, userID int64) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1
			FROM users u
			INNER JOIN roles r ON r.id = u.role_id
			WHERE u.id = $1
			  AND r.name = 'duty_mechanic'
		);
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check duty mechanic user: %w", err)
	}

	return exists, nil
}

type mechanicShiftScanner interface {
	Scan(dest ...any) error
}

func scanMechanicShift(scanner mechanicShiftScanner) (*models.MechanicShift, error) {
	var item models.MechanicShift
	var middleName sql.NullString
	var timeTo sql.NullString
	var comment sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.UserEmail,
		&item.UserFirstName,
		&item.UserLastName,
		&middleName,
		&item.UserRoleID,
		&item.UserRoleName,
		&item.ShiftDate,
		&item.TimeFrom,
		&timeTo,
		&comment,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.UserMiddleName = nullableStringPtrUser(middleName)
	item.TimeTo = nullableStringPtrUser(timeTo)
	item.Comment = nullableStringPtrUser(comment)

	return &item, nil
}

func scanMechanicShiftRows(rows *sql.Rows) (*models.MechanicShift, error) {
	return scanMechanicShift(rows)
}

func nullableStringPtrUser(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func nullableStringValue(v string) any {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func normalizeMechanicShiftSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "user_id":
		return "ms.user_id"
	case "shift_date", "date":
		return "ms.shift_date"
	case "time_from":
		return "ms.time_from"
	case "time_to":
		return "ms.time_to"
	case "is_active":
		return "ms.is_active"
	case "updated_at":
		return "ms.updated_at"
	case "created_at":
		return "ms.created_at"
	default:
		return "ms.shift_date"
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

func mapMechanicShiftError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "mechanic_shifts_user_id_fkey"):
		return ErrMechanicShiftUserNotFound
	default:
		return err
	}
}
