package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/user_module/models"
)

var (
	ErrDriverShiftNotFound       = errors.New("driver shift not found")
	ErrDriverShiftDriverNotFound = errors.New("driver not found")
)

type DriverShiftRepo struct {
	db *sql.DB
}

func NewDriverShiftRepo(db *sql.DB) *DriverShiftRepo {
	return &DriverShiftRepo{db: db}
}

type CreateDriverShiftParams struct {
	DriverID  int64
	ShiftDate string
	TimeFrom  string
	TimeTo    *string
	Comment   *string
	IsActive  bool
}

type UpdateDriverShiftParams struct {
	DriverID  *int64
	ShiftDate *string
	TimeFrom  *string
	TimeTo    *string
	Comment   *string
	IsActive  *bool
}

type ListDriverShiftsParams struct {
	DriverID int64
	DateFrom string
	DateTo   string
	IsActive *bool
	Limit    int
	Offset   int
	SortBy   string
	Order    string
}

type DriverShiftRepository interface {
	RefreshActivity(ctx context.Context) error
	Create(ctx context.Context, p CreateDriverShiftParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.DriverShift, error)
	List(ctx context.Context, p ListDriverShiftsParams) ([]models.DriverShift, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdateDriverShiftParams) (bool, error)
	UpdateActivityByID(ctx context.Context, id int64, isActive bool) (bool, error)
	SoftDeleteByID(ctx context.Context, id int64, deletedByUserID int64) (bool, error)
	DriverExists(ctx context.Context, driverID int64) (bool, error)
}

func (r *DriverShiftRepo) RefreshActivity(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `SELECT refresh_driver_shifts_activity();`)
	if err != nil {
		return fmt.Errorf("refresh driver shifts activity: %w", err)
	}

	return nil
}

func (r *DriverShiftRepo) Create(ctx context.Context, p CreateDriverShiftParams) (int64, error) {
	const q = `
		INSERT INTO driver_shifts (driver_id, shift_date, time_from, time_to, comment, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.DriverID, p.ShiftDate, p.TimeFrom, p.TimeTo, p.Comment, p.IsActive).Scan(&id); err != nil {
		return 0, mapDriverShiftError(err)
	}

	return id, nil
}

func (r *DriverShiftRepo) GetByID(ctx context.Context, id int64) (*models.DriverShift, error) {
	const q = `
		SELECT
			ds.id,
			ds.driver_id,
			d.iin,
			d.name,
			d.surname,
			d.middlename,
			d.phone,
			d.mail,
			d.status_id AS driver_status_id,
			st.code AS driver_status_code,
			st.name AS driver_status_name,
			ds.shift_date,
			ds.time_from::text,
			ds.time_to::text,
			ds.comment,
			ds.is_active,
			COUNT(t.id) AS tripsheets_count,
			ds.created_at,
			ds.updated_at
		FROM driver_shifts ds
		INNER JOIN drivers d ON d.id = ds.driver_id
		INNER JOIN driver_statuses st ON st.id = d.status_id
		LEFT JOIN tripsheets t ON t.driver_shift_id = ds.id
		WHERE ds.id = $1
		  AND ds.is_deleted = FALSE
		GROUP BY ds.id, d.id, st.id;
	`

	item, err := scanDriverShift(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDriverShiftNotFound
		}
		return nil, fmt.Errorf("get driver shift by id: %w", err)
	}

	tripsheets, err := r.ListTripsheetsByShiftID(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Tripsheets = tripsheets

	return item, nil
}

func (r *DriverShiftRepo) List(ctx context.Context, p ListDriverShiftsParams) ([]models.DriverShift, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := []string{"ds.is_deleted = FALSE"}
	args := make([]any, 0, 8)
	argPos := 1

	if p.DriverID > 0 {
		conds = append(conds, fmt.Sprintf("ds.driver_id = $%d", argPos))
		args = append(args, p.DriverID)
		argPos++
	}

	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("ds.shift_date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("ds.shift_date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if p.IsActive != nil {
		conds = append(conds, fmt.Sprintf("ds.is_active = $%d", argPos))
		args = append(args, *p.IsActive)
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `
		SELECT COUNT(*)
		FROM driver_shifts ds
		INNER JOIN drivers d ON d.id = ds.driver_id
		INNER JOIN driver_statuses st ON st.id = d.status_id
	` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list driver shifts count: %w", err)
	}

	sortBy := normalizeDriverShiftSortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	listQ := fmt.Sprintf(`
		SELECT
			ds.id,
			ds.driver_id,
			d.iin,
			d.name,
			d.surname,
			d.middlename,
			d.phone,
			d.mail,
			d.status_id AS driver_status_id,
			st.code AS driver_status_code,
			st.name AS driver_status_name,
			ds.shift_date,
			ds.time_from::text,
			ds.time_to::text,
			ds.comment,
			ds.is_active,
			COUNT(t.id) AS tripsheets_count,
			ds.created_at,
			ds.updated_at
		FROM driver_shifts ds
		INNER JOIN drivers d ON d.id = ds.driver_id
		INNER JOIN driver_statuses st ON st.id = d.status_id
		LEFT JOIN tripsheets t ON t.driver_shift_id = ds.id
		%s
		GROUP BY ds.id, d.id, st.id
		ORDER BY %s %s, ds.id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list driver shifts: %w", err)
	}
	defer rows.Close()

	items := make([]models.DriverShift, 0)
	for rows.Next() {
		item, err := scanDriverShiftRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan driver shift: %w", err)
		}
		item.Tripsheets = []models.DriverShiftTripsheet{}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows driver shifts: %w", err)
	}

	return items, total, nil
}

func (r *DriverShiftRepo) UpdateByID(ctx context.Context, id int64, p UpdateDriverShiftParams) (bool, error) {
	setParts := make([]string, 0, 8)
	args := make([]any, 0, 10)
	argPos := 1

	if p.DriverID != nil {
		setParts = append(setParts, fmt.Sprintf("driver_id = $%d", argPos))
		args = append(args, *p.DriverID)
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
		UPDATE driver_shifts
		SET %s
		WHERE id = $%d
		  AND is_deleted = FALSE;
	`, strings.Join(setParts, ", "), argPos)

	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return false, mapDriverShiftError(err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update driver shift rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *DriverShiftRepo) UpdateActivityByID(ctx context.Context, id int64, isActive bool) (bool, error) {
	const q = `
		UPDATE driver_shifts
		SET is_active = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND is_deleted = FALSE;
	`

	res, err := r.db.ExecContext(ctx, q, isActive, id)
	if err != nil {
		return false, fmt.Errorf("update driver shift activity: %w", err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update driver shift activity rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *DriverShiftRepo) SoftDeleteByID(ctx context.Context, id int64, deletedByUserID int64) (bool, error) {
	const q = `
		UPDATE driver_shifts
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
		return false, mapDriverShiftError(err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("soft delete driver shift rows affected: %w", err)
	}

	return aff > 0, nil
}

func (r *DriverShiftRepo) DriverExists(ctx context.Context, driverID int64) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM drivers WHERE id = $1);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, driverID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check driver exists: %w", err)
	}
	return exists, nil
}

func (r *DriverShiftRepo) ListTripsheetsByShiftID(ctx context.Context, driverShiftID int64) ([]models.DriverShiftTripsheet, error) {
	const q = `
		SELECT
			t.id,
			t.tripsheet_number,
			t.tripsheet_date,
			t.vehicle_id,
			t.vehicle_brand,
			t.vehicle_plate_number,
			t.start_time,
			t.end_time,
			t.status_id,
			ts.name AS status_name,
			t.created_at,
			t.updated_at
		FROM tripsheets t
		LEFT JOIN tripsheet_statuses ts ON ts.id = t.status_id
		WHERE t.driver_shift_id = $1
		ORDER BY t.tripsheet_date DESC, t.id DESC;
	`

	rows, err := r.db.QueryContext(ctx, q, driverShiftID)
	if err != nil {
		return nil, fmt.Errorf("list tripsheets by driver shift id: %w", err)
	}
	defer rows.Close()

	items := make([]models.DriverShiftTripsheet, 0)
	for rows.Next() {
		var item models.DriverShiftTripsheet
		var vehicleID sql.NullInt64
		var vehicleBrand sql.NullString
		var startTime sql.NullTime
		var endTime sql.NullTime
		var statusName sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.TripsheetNumber,
			&item.TripsheetDate,
			&vehicleID,
			&vehicleBrand,
			&item.VehiclePlateNumber,
			&startTime,
			&endTime,
			&item.StatusID,
			&statusName,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan tripsheet by driver shift id: %w", err)
		}

		item.VehicleID = nullableInt64PtrDriverShift(vehicleID)
		item.VehicleBrand = nullableStringPtrDriverShift(vehicleBrand)
		item.StartTime = nullableTimePtrDriverShift(startTime)
		item.EndTime = nullableTimePtrDriverShift(endTime)
		item.StatusName = nullableStringPtrDriverShift(statusName)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows tripsheets by driver shift id: %w", err)
	}

	return items, nil
}

type driverShiftScanner interface {
	Scan(dest ...any) error
}

func scanDriverShift(scanner driverShiftScanner) (*models.DriverShift, error) {
	var item models.DriverShift
	var middlename sql.NullString
	var phone sql.NullString
	var mail sql.NullString
	var timeTo sql.NullString
	var comment sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.DriverID,
		&item.DriverIIN,
		&item.DriverName,
		&item.DriverSurname,
		&middlename,
		&phone,
		&mail,
		&item.DriverStatusID,
		&item.DriverStatusCode,
		&item.DriverStatusName,
		&item.ShiftDate,
		&item.TimeFrom,
		&timeTo,
		&comment,
		&item.IsActive,
		&item.TripsheetsCount,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.DriverMiddlename = nullableStringValueFromSQL(middlename)
	item.DriverPhone = nullableStringValueFromSQL(phone)
	item.DriverMail = nullableStringValueFromSQL(mail)
	item.TimeTo = nullableStringPtrDriverShift(timeTo)
	item.Comment = nullableStringPtrDriverShift(comment)

	return &item, nil
}

func scanDriverShiftRows(rows *sql.Rows) (*models.DriverShift, error) {
	return scanDriverShift(rows)
}

func nullableStringValueFromSQL(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func nullableStringPtrDriverShift(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func nullableInt64PtrDriverShift(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullableTimePtrDriverShift(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	return &v.Time
}

func normalizeDriverShiftSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "driver_id":
		return "ds.driver_id"
	case "shift_date", "date":
		return "ds.shift_date"
	case "time_from":
		return "ds.time_from"
	case "time_to":
		return "ds.time_to"
	case "is_active":
		return "ds.is_active"
	case "created_at":
		return "ds.created_at"
	case "updated_at":
		return "ds.updated_at"
	case "tripsheets_count":
		return "tripsheets_count"
	default:
		return "ds.shift_date"
	}
}

func mapDriverShiftError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "driver_shifts_driver_id_fkey"):
		return ErrDriverShiftDriverNotFound
	case strings.Contains(msg, "driver_shifts_deleted_by_user_id_fkey"):
		return ErrMechanicShiftUserNotFound
	default:
		return err
	}
}
