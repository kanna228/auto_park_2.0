package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/warehouse_module/models"
)

var ErrPartRequestLocked = errors.New("part request cannot be changed after approval or rejection")
var ErrPartRequestPartNotFound = errors.New("part not found")
var ErrPartRequestStatusNotFound = errors.New("part request status not found")
var ErrPartRequestUserNotFound = errors.New("user not found")
var ErrPartRequestInsufficientStock = errors.New("not enough part quantity in stock")
var ErrPartRequestRepairContextRequired = errors.New("vehicle_id, mechanic_shift_id and planned_replacement_at are required to complete repair")
var ErrPartRequestRepairCompletionForbidden = errors.New("part request must be approved before repair can be completed")

type CreatePartRequestParams struct {
	PartID               int64
	Quantity             int64
	MechanicComment      string
	VehicleID            *int64
	MechanicShiftID      *int64
	PlannedReplacementAt string
	StatusID             int64
	AuthorUserID         int64
	HistoryComment       string
}

type UpdatePartRequestParams struct {
	PartID               int64
	Quantity             int64
	MechanicComment      string
	VehicleID            *int64
	MechanicShiftID      *int64
	PlannedReplacementAt string
	StatusID             int64
	TargetStatusCode     string
	RejectionComment     string
	ChangedByUserID      int64
	HistoryComment       string
}

type UpdatePartRequestStatusParams struct {
	StatusID         int64
	TargetStatusCode string
	RejectionComment string
	ChangedByUserID  int64
	HistoryComment   string
}

type DeletePartRequestParams struct {
	ChangedByUserID int64
	HistoryComment  string
}

type CompletePartRequestRepairParams struct {
	VehicleID            *int64
	MechanicShiftID      *int64
	InstalledAt          string
	PlannedReplacementAt string
	CompletedByUserID    int64
	HistoryComment       string
}

type ListPartRequestsParams struct {
	PartID       int64
	StatusID     int64
	StatusCode   string
	AuthorUserID int64
	DateFrom     string
	DateTo       string
	Limit        int
	Offset       int
	SortBy       string
	Order        string
}

type ListPartRequestHistoryParams struct {
	PartRequestID   int64
	StatusID        int64
	StatusCode      string
	ChangedByUserID int64
	DateFrom        string
	DateTo          string
	Limit           int
	Offset          int
	SortBy          string
	Order           string
}

type PartRequestRepository interface {
	Create(ctx context.Context, p CreatePartRequestParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.PartRequest, error)
	List(ctx context.Context, p ListPartRequestsParams) ([]models.PartRequest, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdatePartRequestParams) (bool, error)
	UpdateStatusByID(ctx context.Context, id int64, p UpdatePartRequestStatusParams) (bool, error)
	CompleteRepairByID(ctx context.Context, id int64, p CompletePartRequestRepairParams) (bool, error)
	DeleteByID(ctx context.Context, id int64, p DeletePartRequestParams) (bool, error)
	PartExists(ctx context.Context, partID int64) (bool, error)
	StatusExists(ctx context.Context, statusID int64) (bool, error)
	GetStatusByID(ctx context.Context, statusID int64) (*models.PartRequestStatus, error)
	ListStatuses(ctx context.Context) ([]models.PartRequestStatus, error)
	ListHistoryByRequestID(ctx context.Context, partRequestID int64, limit int, offset int) ([]models.PartRequestHistory, int64, error)
	ListHistory(ctx context.Context, p ListPartRequestHistoryParams) ([]models.PartRequestHistory, int64, error)
}

type partRequestRepo struct {
	db *sql.DB
}

func NewPartRequestRepository(db *sql.DB) PartRequestRepository {
	return &partRequestRepo{db: db}
}

func (r *partRequestRepo) Create(ctx context.Context, p CreatePartRequestParams) (int64, error) {
	statusID := p.StatusID
	if statusID <= 0 {
		statusID = 1
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin create part request tx: %w", err)
	}
	defer rollbackPartRequestTx(tx)

	const insertRequestQ = `
		INSERT INTO part_requests (
			part_id,
			quantity,
			mechanic_comment,
			vehicle_id,
			mechanic_shift_id,
			planned_replacement_at,
			status_id,
			author_user_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::DATE, $7, $8, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := tx.QueryRowContext(
		ctx,
		insertRequestQ,
		p.PartID,
		p.Quantity,
		p.MechanicComment,
		p.VehicleID,
		p.MechanicShiftID,
		p.PlannedReplacementAt,
		statusID,
		p.AuthorUserID,
	).Scan(&id); err != nil {
		return 0, mapPartRequestError(err)
	}

	if err := insertPartRequestHistory(ctx, tx, id, statusID, p.AuthorUserID, p.HistoryComment); err != nil {
		return 0, mapPartRequestError(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit create part request tx: %w", err)
	}

	return id, nil
}

func (r *partRequestRepo) GetByID(ctx context.Context, id int64) (*models.PartRequest, error) {
	const q = `
		SELECT
			pr.id,
			pr.part_id,
			p.part_id AS part_catalog_code,
			p.name AS part_name,
			p.category AS part_category,
			p.quantity AS part_quantity,
			pr.quantity,
			pr.mechanic_comment,
			pr.rejection_comment,
			pr.vehicle_id,
			pr.mechanic_shift_id,
			pr.planned_replacement_at,
			pr.repair_status,
			pr.completed_at,
			pr.completed_by_user_id,
			pr.vehicle_part_installation_id,
			pr.status_id,
			s.code AS status_code,
			s.name AS status_name,
			pr.author_user_id,
			u.email AS author_email,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS author_full_name,
			pr.created_at,
			pr.updated_at
		FROM part_requests pr
		INNER JOIN parts_catalog p ON p.id = pr.part_id
		INNER JOIN part_request_statuses s ON s.id = pr.status_id
		LEFT JOIN users u ON u.id = pr.author_user_id
		WHERE pr.id = $1
		  AND pr.is_deleted = FALSE;
	`

	item, err := scanPartRequest(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get part request by id: %w", err)
	}
	return item, nil
}

func (r *partRequestRepo) List(ctx context.Context, p ListPartRequestsParams) ([]models.PartRequest, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := normalizePartRequestSortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	conds := []string{"pr.is_deleted = FALSE"}
	args := make([]any, 0, 8)
	argPos := 1

	if p.PartID > 0 {
		conds = append(conds, fmt.Sprintf("pr.part_id = $%d", argPos))
		args = append(args, p.PartID)
		argPos++
	}
	if p.StatusID > 0 {
		conds = append(conds, fmt.Sprintf("pr.status_id = $%d", argPos))
		args = append(args, p.StatusID)
		argPos++
	}
	if v := strings.TrimSpace(p.StatusCode); v != "" {
		conds = append(conds, fmt.Sprintf("s.code = $%d", argPos))
		args = append(args, strings.ToLower(v))
		argPos++
	}
	if p.AuthorUserID > 0 {
		conds = append(conds, fmt.Sprintf("pr.author_user_id = $%d", argPos))
		args = append(args, p.AuthorUserID)
		argPos++
	}

	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("pr.created_at::date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("pr.created_at::date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `
		SELECT COUNT(*)
		FROM part_requests pr
		INNER JOIN part_request_statuses s ON s.id = pr.status_id
	` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list part requests count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT
			pr.id,
			pr.part_id,
			p.part_id AS part_catalog_code,
			p.name AS part_name,
			p.category AS part_category,
			p.quantity AS part_quantity,
			pr.quantity,
			pr.mechanic_comment,
			pr.rejection_comment,
			pr.vehicle_id,
			pr.mechanic_shift_id,
			pr.planned_replacement_at,
			pr.repair_status,
			pr.completed_at,
			pr.completed_by_user_id,
			pr.vehicle_part_installation_id,
			pr.status_id,
			s.code AS status_code,
			s.name AS status_name,
			pr.author_user_id,
			u.email AS author_email,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS author_full_name,
			pr.created_at,
			pr.updated_at
		FROM part_requests pr
		INNER JOIN parts_catalog p ON p.id = pr.part_id
		INNER JOIN part_request_statuses s ON s.id = pr.status_id
		LEFT JOIN users u ON u.id = pr.author_user_id
		%s
		ORDER BY %s %s, pr.id ASC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list part requests: %w", err)
	}
	defer rows.Close()

	items := make([]models.PartRequest, 0)
	for rows.Next() {
		item, err := scanPartRequestRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list part requests scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list part requests rows: %w", err)
	}

	return items, total, nil
}

func (r *partRequestRepo) UpdateByID(ctx context.Context, id int64, p UpdatePartRequestParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin update part request tx: %w", err)
	}
	defer rollbackPartRequestTx(tx)

	var currentStatusCode string
	const lockQ = `
		SELECT s.code
		FROM part_requests pr
		INNER JOIN part_request_statuses s ON s.id = pr.status_id
		WHERE pr.id = $1
		  AND pr.is_deleted = FALSE
		FOR UPDATE OF pr;
	`
	if err := tx.QueryRowContext(ctx, lockQ, id).Scan(&currentStatusCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock part request before update: %w", err)
	}

	const q = `
		UPDATE part_requests
		SET part_id = $1,
			quantity = $2,
			mechanic_comment = $3,
			status_id = $4,
			rejection_comment = $5,
			vehicle_id = $6,
			mechanic_shift_id = $7,
			planned_replacement_at = NULLIF($8, '')::DATE,
			repair_status = CASE WHEN $9 = 'approved' AND $10 <> 'approved' THEN 'in_progress' ELSE repair_status END,
			updated_at = NOW()
		WHERE id = $11
		  AND is_deleted = FALSE;
	`
	res, err := tx.ExecContext(
		ctx,
		q,
		p.PartID,
		p.Quantity,
		p.MechanicComment,
		p.StatusID,
		nullableTextParam(p.RejectionComment),
		p.VehicleID,
		p.MechanicShiftID,
		p.PlannedReplacementAt,
		p.TargetStatusCode,
		currentStatusCode,
		id,
	)
	if err != nil {
		return false, mapPartRequestError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update part request rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	if shouldIssuePartRequestStock(currentStatusCode, p.TargetStatusCode) {
		if err := issuePartRequestStockTx(ctx, tx, p.PartID, p.Quantity, id, p.ChangedByUserID); err != nil {
			return false, err
		}
	}

	if err := insertPartRequestHistory(ctx, tx, id, p.StatusID, p.ChangedByUserID, p.HistoryComment); err != nil {
		return false, mapPartRequestError(err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit update part request tx: %w", err)
	}

	return true, nil
}

func (r *partRequestRepo) UpdateStatusByID(ctx context.Context, id int64, p UpdatePartRequestStatusParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin update part request status tx: %w", err)
	}
	defer rollbackPartRequestTx(tx)

	var partID int64
	var quantity int64
	var currentStatusCode string
	const lockQ = `
		SELECT pr.part_id, pr.quantity, s.code
		FROM part_requests pr
		INNER JOIN part_request_statuses s ON s.id = pr.status_id
		WHERE pr.id = $1
		  AND pr.is_deleted = FALSE
		FOR UPDATE OF pr;
	`
	if err := tx.QueryRowContext(ctx, lockQ, id).Scan(&partID, &quantity, &currentStatusCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock part request before status update: %w", err)
	}

	const q = `
		UPDATE part_requests
		SET status_id = $1,
			rejection_comment = $2,
			repair_status = CASE WHEN $3 = 'approved' AND $4 <> 'approved' THEN 'in_progress' ELSE repair_status END,
			updated_at = NOW()
		WHERE id = $5
		  AND is_deleted = FALSE;
	`
	res, err := tx.ExecContext(ctx, q, p.StatusID, nullableTextParam(p.RejectionComment), p.TargetStatusCode, currentStatusCode, id)
	if err != nil {
		return false, mapPartRequestError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update part request status rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	if shouldIssuePartRequestStock(currentStatusCode, p.TargetStatusCode) {
		if err := issuePartRequestStockTx(ctx, tx, partID, quantity, id, p.ChangedByUserID); err != nil {
			return false, err
		}
	}

	if err := insertPartRequestHistory(ctx, tx, id, p.StatusID, p.ChangedByUserID, p.HistoryComment); err != nil {
		return false, mapPartRequestError(err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit update part request status tx: %w", err)
	}

	return true, nil
}

func (r *partRequestRepo) CompleteRepairByID(ctx context.Context, id int64, p CompletePartRequestRepairParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin complete part request repair tx: %w", err)
	}
	defer rollbackPartRequestTx(tx)

	var partID int64
	var quantity int64
	var statusID int64
	var statusCode string
	var storedVehicleID sql.NullInt64
	var storedMechanicShiftID sql.NullInt64
	var storedPlannedReplacementAt string
	var repairStatus string
	var vehiclePartInstallationID sql.NullInt64

	const lockQ = `
		SELECT
			pr.part_id,
			pr.quantity,
			pr.status_id,
			s.code,
			pr.vehicle_id,
			pr.mechanic_shift_id,
			COALESCE(pr.planned_replacement_at::TEXT, ''),
			pr.repair_status,
			pr.vehicle_part_installation_id
		FROM part_requests pr
		INNER JOIN part_request_statuses s ON s.id = pr.status_id
		WHERE pr.id = $1
		  AND pr.is_deleted = FALSE
		FOR UPDATE OF pr;
	`
	if err := tx.QueryRowContext(ctx, lockQ, id).Scan(
		&partID,
		&quantity,
		&statusID,
		&statusCode,
		&storedVehicleID,
		&storedMechanicShiftID,
		&storedPlannedReplacementAt,
		&repairStatus,
		&vehiclePartInstallationID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock part request before repair completion: %w", err)
	}

	if strings.ToLower(strings.TrimSpace(statusCode)) != "approved" {
		return false, ErrPartRequestRepairCompletionForbidden
	}
	if strings.ToLower(strings.TrimSpace(repairStatus)) == "completed" && vehiclePartInstallationID.Valid {
		return true, nil
	}

	vehicleID := resolveInt64Param(p.VehicleID, storedVehicleID)
	mechanicShiftID := resolveInt64Param(p.MechanicShiftID, storedMechanicShiftID)
	plannedReplacementAt := strings.TrimSpace(p.PlannedReplacementAt)
	if plannedReplacementAt == "" {
		plannedReplacementAt = strings.TrimSpace(storedPlannedReplacementAt)
	}
	if vehicleID <= 0 || mechanicShiftID <= 0 || plannedReplacementAt == "" {
		return false, ErrPartRequestRepairContextRequired
	}

	part, err := getPartForUpdate(ctx, tx, partID)
	if err != nil {
		return false, err
	}
	if err := ensureVehicleExistsTx(ctx, tx, vehicleID); err != nil {
		return false, err
	}
	if err := ensureInstallerExistsTx(ctx, tx, p.CompletedByUserID); err != nil {
		return false, err
	}
	if err := ensureMechanicShiftExistsTx(ctx, tx, mechanicShiftID); err != nil {
		return false, err
	}
	if !part.IsConsumable {
		exists, err := activeInstallationExistsTx(ctx, tx, partID, vehicleID, 0)
		if err != nil {
			return false, err
		}
		if exists {
			return false, ErrVehiclePartInstallationActiveDuplicate
		}
	}

	const insertInstallationQ = `
		INSERT INTO vehicle_part_installations (
			part_id,
			vehicle_id,
			mechanic_shift_id,
			installed_at,
			planned_replacement_at,
			quantity,
			unit_price,
			total_price,
			installed_by_user_id,
			is_active,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE, NOW(), NOW())
		RETURNING id;
	`

	var installationID int64
	if err := tx.QueryRowContext(
		ctx,
		insertInstallationQ,
		partID,
		vehicleID,
		mechanicShiftID,
		p.InstalledAt,
		plannedReplacementAt,
		quantity,
		part.Price,
		part.Price*float64(quantity),
		p.CompletedByUserID,
	).Scan(&installationID); err != nil {
		return false, mapVehiclePartInstallationError(err)
	}

	const updateRequestQ = `
		UPDATE part_requests
		SET vehicle_id = $1,
			mechanic_shift_id = $2,
			planned_replacement_at = $3,
			repair_status = 'completed',
			completed_at = NOW(),
			completed_by_user_id = $4,
			vehicle_part_installation_id = $5,
			updated_at = NOW()
		WHERE id = $6
		  AND is_deleted = FALSE;
	`
	if _, err := tx.ExecContext(ctx, updateRequestQ, vehicleID, mechanicShiftID, plannedReplacementAt, p.CompletedByUserID, installationID, id); err != nil {
		return false, mapPartRequestError(err)
	}

	if err := insertPartRequestHistory(ctx, tx, id, statusID, p.CompletedByUserID, p.HistoryComment); err != nil {
		return false, mapPartRequestError(err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit complete part request repair tx: %w", err)
	}

	return true, nil
}

func (r *partRequestRepo) DeleteByID(ctx context.Context, id int64, p DeletePartRequestParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin delete part request tx: %w", err)
	}
	defer rollbackPartRequestTx(tx)

	var statusID int64
	const getStatusQ = `
		SELECT status_id
		FROM part_requests
		WHERE id = $1
		  AND is_deleted = FALSE
		FOR UPDATE;
	`
	if err := tx.QueryRowContext(ctx, getStatusQ, id).Scan(&statusID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get part request status before delete: %w", err)
	}

	const q = `
		UPDATE part_requests
		SET is_deleted = TRUE,
			deleted_at = NOW(),
			deleted_by_user_id = $1,
			updated_at = NOW()
		WHERE id = $2
		  AND is_deleted = FALSE;
	`
	res, err := tx.ExecContext(ctx, q, p.ChangedByUserID, id)
	if err != nil {
		return false, mapPartRequestError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete part request rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	if err := insertPartRequestHistory(ctx, tx, id, statusID, p.ChangedByUserID, p.HistoryComment); err != nil {
		return false, mapPartRequestError(err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit delete part request tx: %w", err)
	}

	return true, nil
}

func (r *partRequestRepo) PartExists(ctx context.Context, partID int64) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM parts_catalog WHERE id = $1);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, partID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check part exists: %w", err)
	}
	return exists, nil
}

func (r *partRequestRepo) StatusExists(ctx context.Context, statusID int64) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM part_request_statuses WHERE id = $1);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, statusID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check part request status exists: %w", err)
	}
	return exists, nil
}

func (r *partRequestRepo) GetStatusByID(ctx context.Context, statusID int64) (*models.PartRequestStatus, error) {
	const q = `
		SELECT id, code, name
		FROM part_request_statuses
		WHERE id = $1;
	`

	var item models.PartRequestStatus
	if err := r.db.QueryRowContext(ctx, q, statusID).Scan(&item.ID, &item.Code, &item.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get part request status by id: %w", err)
	}

	return &item, nil
}

func (r *partRequestRepo) ListStatuses(ctx context.Context) ([]models.PartRequestStatus, error) {
	const q = `
		SELECT id, code, name
		FROM part_request_statuses
		ORDER BY id ASC;
	`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list part request statuses: %w", err)
	}
	defer rows.Close()

	items := make([]models.PartRequestStatus, 0)
	for rows.Next() {
		var item models.PartRequestStatus
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, fmt.Errorf("list part request statuses scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list part request statuses rows: %w", err)
	}
	return items, nil
}

func (r *partRequestRepo) ListHistoryByRequestID(ctx context.Context, partRequestID int64, limit int, offset int) ([]models.PartRequestHistory, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	const countQ = `
		SELECT COUNT(*)
		FROM part_request_history
		WHERE part_request_id = $1;
	`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, partRequestID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list part request history count: %w", err)
	}

	const q = `
		SELECT
			h.id,
			h.part_request_id,
			h.status_id,
			s.code AS status_code,
			s.name AS status_name,
			h.changed_by_user_id,
			u.email AS changed_by_email,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS changed_by_full_name,
			h.comment,
			h.changed_at
		FROM part_request_history h
		INNER JOIN part_request_statuses s ON s.id = h.status_id
		LEFT JOIN users u ON u.id = h.changed_by_user_id
		WHERE h.part_request_id = $1
		ORDER BY h.changed_at ASC, h.id ASC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.db.QueryContext(ctx, q, partRequestID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list part request history: %w", err)
	}
	defer rows.Close()

	items := make([]models.PartRequestHistory, 0)
	for rows.Next() {
		item, err := scanPartRequestHistoryRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list part request history scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list part request history rows: %w", err)
	}

	return items, total, nil
}

func (r *partRequestRepo) ListHistory(ctx context.Context, p ListPartRequestHistoryParams) ([]models.PartRequestHistory, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := normalizePartRequestHistorySortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	conds := []string{"1=1"}
	args := make([]any, 0, 10)
	argPos := 1

	if p.PartRequestID > 0 {
		conds = append(conds, fmt.Sprintf("h.part_request_id = $%d", argPos))
		args = append(args, p.PartRequestID)
		argPos++
	}
	if p.StatusID > 0 {
		conds = append(conds, fmt.Sprintf("h.status_id = $%d", argPos))
		args = append(args, p.StatusID)
		argPos++
	}
	if v := strings.TrimSpace(p.StatusCode); v != "" {
		conds = append(conds, fmt.Sprintf("s.code = $%d", argPos))
		args = append(args, strings.ToLower(v))
		argPos++
	}
	if p.ChangedByUserID > 0 {
		conds = append(conds, fmt.Sprintf("h.changed_by_user_id = $%d", argPos))
		args = append(args, p.ChangedByUserID)
		argPos++
	}
	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("h.changed_at::date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("h.changed_at::date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `
		SELECT COUNT(*)
		FROM part_request_history h
		INNER JOIN part_request_statuses s ON s.id = h.status_id
	` + whereSQL

	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list part request history count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT
			h.id,
			h.part_request_id,
			h.status_id,
			s.code AS status_code,
			s.name AS status_name,
			h.changed_by_user_id,
			u.email AS changed_by_email,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS changed_by_full_name,
			h.comment,
			h.changed_at
		FROM part_request_history h
		INNER JOIN part_request_statuses s ON s.id = h.status_id
		LEFT JOIN users u ON u.id = h.changed_by_user_id
		%s
		ORDER BY %s %s, h.id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list part request history: %w", err)
	}
	defer rows.Close()

	items := make([]models.PartRequestHistory, 0)
	for rows.Next() {
		item, err := scanPartRequestHistoryRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list part request history scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list part request history rows: %w", err)
	}

	return items, total, nil
}

func insertPartRequestHistory(ctx context.Context, tx *sql.Tx, partRequestID int64, statusID int64, changedByUserID int64, comment string) error {
	const q = `
		INSERT INTO part_request_history (part_request_id, status_id, changed_by_user_id, comment, changed_at)
		VALUES ($1, $2, $3, $4, NOW());
	`
	_, err := tx.ExecContext(ctx, q, partRequestID, statusID, changedByUserID, comment)
	return err
}

func shouldIssuePartRequestStock(currentStatusCode string, targetStatusCode string) bool {
	targetStatusCode = strings.ToLower(strings.TrimSpace(targetStatusCode))
	currentStatusCode = strings.ToLower(strings.TrimSpace(currentStatusCode))
	return targetStatusCode == "approved" && currentStatusCode != "approved"
}

func issuePartRequestStockTx(ctx context.Context, tx *sql.Tx, partID int64, quantity int64, partRequestID int64, actorUserID int64) error {
	alreadyIssued, err := partRequestStockAlreadyIssuedTx(ctx, tx, partRequestID)
	if err != nil {
		return err
	}
	if alreadyIssued {
		return nil
	}

	var currentQuantity int64
	const lockPartQ = `
		SELECT quantity
		FROM parts_catalog
		WHERE id = $1
		FOR UPDATE;
	`
	if err := tx.QueryRowContext(ctx, lockPartQ, partID).Scan(&currentQuantity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPartRequestPartNotFound
		}
		return fmt.Errorf("lock part before issue: %w", err)
	}
	if currentQuantity < quantity {
		return ErrPartRequestInsufficientStock
	}

	const updatePartQ = `
		UPDATE parts_catalog
		SET quantity = quantity - $1,
			updated_at = NOW()
		WHERE id = $2;
	`
	if _, err := tx.ExecContext(ctx, updatePartQ, quantity, partID); err != nil {
		return fmt.Errorf("decrease part quantity: %w", err)
	}

	documentNumber := fmt.Sprintf("part-request-%d", partRequestID)
	partRequestIDPtr := partRequestID
	if err := insertPartStockMovementTx(ctx, tx, partID, "issue", quantity, nil, &partRequestIDPtr, &documentNumber, actorUserID); err != nil {
		return err
	}

	return nil
}

func partRequestStockAlreadyIssuedTx(ctx context.Context, tx *sql.Tx, partRequestID int64) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1
			FROM part_stock_movements
			WHERE part_request_id = $1
			  AND type = 'issue'
		);
	`
	var exists bool
	if err := tx.QueryRowContext(ctx, q, partRequestID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check part request stock issue: %w", err)
	}
	return exists, nil
}

func rollbackPartRequestTx(tx *sql.Tx) {
	_ = tx.Rollback()
}

type partRequestScanner interface {
	Scan(dest ...any) error
}

func scanPartRequest(scanner partRequestScanner) (*models.PartRequest, error) {
	var item models.PartRequest
	var rejectionComment sql.NullString
	var vehicleID sql.NullInt64
	var mechanicShiftID sql.NullInt64
	var plannedReplacementAt sql.NullTime
	var completedAt sql.NullTime
	var completedByUserID sql.NullInt64
	var vehiclePartInstallationID sql.NullInt64
	var authorEmail sql.NullString
	var authorFullName sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.PartID,
		&item.PartCatalogCode,
		&item.PartName,
		&item.PartCategory,
		&item.PartQuantity,
		&item.Quantity,
		&item.MechanicComment,
		&rejectionComment,
		&vehicleID,
		&mechanicShiftID,
		&plannedReplacementAt,
		&item.RepairStatus,
		&completedAt,
		&completedByUserID,
		&vehiclePartInstallationID,
		&item.StatusID,
		&item.StatusCode,
		&item.StatusName,
		&item.AuthorUserID,
		&authorEmail,
		&authorFullName,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.RejectionComment = nullableStringPtr(rejectionComment)
	item.VehicleID = nullableInt64Ptr(vehicleID)
	item.MechanicShiftID = nullableInt64Ptr(mechanicShiftID)
	item.PlannedReplacementAt = nullableTimePtr(plannedReplacementAt)
	item.CompletedAt = nullableTimePtr(completedAt)
	item.CompletedByUserID = nullableInt64Ptr(completedByUserID)
	item.VehiclePartInstallationID = nullableInt64Ptr(vehiclePartInstallationID)
	item.AuthorEmail = nullableStringPtr(authorEmail)
	item.AuthorFullName = nullableStringPtr(authorFullName)
	return &item, nil
}

func scanPartRequestRows(rows *sql.Rows) (*models.PartRequest, error) {
	return scanPartRequest(rows)
}

func scanPartRequestHistory(scanner partRequestScanner) (*models.PartRequestHistory, error) {
	var item models.PartRequestHistory
	var changedByEmail sql.NullString
	var changedByFullName sql.NullString

	if err := scanner.Scan(
		&item.ID,
		&item.PartRequestID,
		&item.StatusID,
		&item.StatusCode,
		&item.StatusName,
		&item.ChangedByUserID,
		&changedByEmail,
		&changedByFullName,
		&item.Comment,
		&item.ChangedAt,
	); err != nil {
		return nil, err
	}

	item.ChangedByEmail = nullableStringPtr(changedByEmail)
	item.ChangedByFullName = nullableStringPtr(changedByFullName)
	return &item, nil
}

func scanPartRequestHistoryRows(rows *sql.Rows) (*models.PartRequestHistory, error) {
	return scanPartRequestHistory(rows)
}

func nullableStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func resolveInt64Param(value *int64, fallback sql.NullInt64) int64 {
	if value != nil {
		return *value
	}
	if fallback.Valid {
		return fallback.Int64
	}
	return 0
}

func nullableTextParam(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func normalizePartRequestSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "part_id":
		return "pr.part_id"
	case "quantity":
		return "pr.quantity"
	case "status_id":
		return "pr.status_id"
	case "status_code":
		return "s.code"
	case "author_user_id":
		return "pr.author_user_id"
	case "updated_at":
		return "pr.updated_at"
	default:
		return "pr.created_at"
	}
}

func normalizePartRequestHistorySortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "part_request_id":
		return "h.part_request_id"
	case "status_id":
		return "h.status_id"
	case "status_code":
		return "s.code"
	case "changed_by_user_id":
		return "h.changed_by_user_id"
	case "changed_at":
		return "h.changed_at"
	default:
		return "h.changed_at"
	}
}

func mapPartRequestError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "part_requests_part_id_fkey"):
		return ErrPartRequestPartNotFound
	case strings.Contains(msg, "part_requests_status_id_fkey"):
		return ErrPartRequestStatusNotFound
	case strings.Contains(msg, "part_requests_author_user_id_fkey"):
		return ErrPartRequestUserNotFound
	case strings.Contains(msg, "part_requests_vehicle_id_fkey"):
		return ErrVehiclePartInstallationVehicleNotFound
	case strings.Contains(msg, "part_requests_mechanic_shift_id_fkey"):
		return ErrVehiclePartInstallationMechanicShiftNotFound
	case strings.Contains(msg, "part_requests_completed_by_user_id_fkey"):
		return ErrPartRequestUserNotFound
	case strings.Contains(msg, "part_request_history_status_id_fkey"):
		return ErrPartRequestStatusNotFound
	case strings.Contains(msg, "part_request_history_changed_by_user_id_fkey"):
		return ErrPartRequestUserNotFound
	default:
		return err
	}
}
