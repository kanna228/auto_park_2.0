package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"auto_park/modules/warehouse_module/models"
)

var ErrPurchaseRequestLocked = errors.New("purchase request cannot be changed after confirmation or cancellation")
var ErrPurchaseRequestSourcePartRequestNotFound = errors.New("source part request not found")

type CreatePurchaseRequestParams struct {
	PartID              int64
	Quantity            int64
	SourcePartRequestID *int64
	Comment             *string
	CreatedByUserID     int64
}

type ListPurchaseRequestsParams struct {
	PartID              int64
	SourcePartRequestID int64
	Status              string
	Limit               int
	Offset              int
	SortBy              string
	Order               string
}

type PurchaseRequestRepository interface {
	Create(ctx context.Context, p CreatePurchaseRequestParams) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.PurchaseRequest, error)
	List(ctx context.Context, p ListPurchaseRequestsParams) ([]models.PurchaseRequest, int64, error)
	Confirm(ctx context.Context, id int64, actorUserID int64) (bool, error)
	Cancel(ctx context.Context, id int64, actorUserID int64, comment *string) (bool, error)
}

type purchaseRequestRepo struct {
	db *sql.DB
}

func NewPurchaseRequestRepository(db *sql.DB) PurchaseRequestRepository {
	return &purchaseRequestRepo{db: db}
}

func (r *purchaseRequestRepo) Create(ctx context.Context, p CreatePurchaseRequestParams) (int64, error) {
	const q = `
		INSERT INTO part_purchase_requests (
			part_id,
			quantity,
			source_part_request_id,
			comment,
			created_by_user_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.PartID, p.Quantity, p.SourcePartRequestID, p.Comment, p.CreatedByUserID).Scan(&id); err != nil {
		return 0, mapPurchaseRequestError(err)
	}
	return id, nil
}

func (r *purchaseRequestRepo) GetByID(ctx context.Context, id int64) (*models.PurchaseRequest, error) {
	const q = purchaseRequestSelect + `
		WHERE pr.id = $1;
	`

	item, err := scanPurchaseRequest(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get purchase request by id: %w", err)
	}
	return item, nil
}

func (r *purchaseRequestRepo) List(ctx context.Context, p ListPurchaseRequestsParams) ([]models.PurchaseRequest, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := []string{"1=1"}
	args := make([]any, 0, 8)
	argPos := 1

	if p.PartID > 0 {
		conds = append(conds, fmt.Sprintf("pr.part_id = $%d", argPos))
		args = append(args, p.PartID)
		argPos++
	}
	if p.SourcePartRequestID > 0 {
		conds = append(conds, fmt.Sprintf("pr.source_part_request_id = $%d", argPos))
		args = append(args, p.SourcePartRequestID)
		argPos++
	}
	if v := strings.TrimSpace(p.Status); v != "" {
		conds = append(conds, fmt.Sprintf("pr.status = $%d", argPos))
		args = append(args, strings.ToLower(v))
		argPos++
	}

	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `SELECT COUNT(*) FROM part_purchase_requests pr` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list purchase requests count: %w", err)
	}

	sortBy := normalizePurchaseRequestSortBy(p.SortBy)
	order := normalizeOrder(p.Order)
	listQ := fmt.Sprintf(`
		%s
		%s
		ORDER BY %s %s, pr.id DESC
		LIMIT $%d OFFSET $%d;
	`, purchaseRequestSelect, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list purchase requests: %w", err)
	}
	defer rows.Close()

	items := make([]models.PurchaseRequest, 0)
	for rows.Next() {
		item, err := scanPurchaseRequestRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list purchase requests scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list purchase requests rows: %w", err)
	}

	return items, total, nil
}

func (r *purchaseRequestRepo) Confirm(ctx context.Context, id int64, actorUserID int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin confirm purchase request: %w", err)
	}
	defer rollbackPartTx(tx)

	var partID int64
	var quantity int64
	var status string
	var sourcePartRequestID sql.NullInt64
	const lockQ = `
		SELECT part_id, quantity, status, source_part_request_id
		FROM part_purchase_requests
		WHERE id = $1
		FOR UPDATE;
	`
	if err := tx.QueryRowContext(ctx, lockQ, id).Scan(&partID, &quantity, &status, &sourcePartRequestID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock purchase request: %w", err)
	}
	if status == "confirmed" {
		return true, nil
	}
	if status != "new" {
		return false, ErrPurchaseRequestLocked
	}

	if err := increasePartQuantityPartTx(ctx, tx, partID, quantity); err != nil {
		return false, err
	}

	documentNumber := fmt.Sprintf("purchase-request-%d", id)
	var sourceID *int64
	if sourcePartRequestID.Valid {
		v := sourcePartRequestID.Int64
		sourceID = &v
	}
	if err := insertPartStockMovementTx(ctx, tx, partID, "arrival", quantity, nil, sourceID, &documentNumber, actorUserID); err != nil {
		return false, err
	}

	const updateQ = `
		UPDATE part_purchase_requests
		SET status = 'confirmed',
			confirmed_by_user_id = $1,
			confirmed_at = NOW(),
			updated_at = NOW()
		WHERE id = $2;
	`
	if _, err := tx.ExecContext(ctx, updateQ, actorUserID, id); err != nil {
		return false, fmt.Errorf("confirm purchase request update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit confirm purchase request: %w", err)
	}
	return true, nil
}

func (r *purchaseRequestRepo) Cancel(ctx context.Context, id int64, actorUserID int64, comment *string) (bool, error) {
	const q = `
		UPDATE part_purchase_requests
		SET status = 'cancelled',
			comment = COALESCE($1, comment),
			confirmed_by_user_id = $2,
			confirmed_at = NOW(),
			updated_at = NOW()
		WHERE id = $3
		  AND status = 'new';
	`
	res, err := r.db.ExecContext(ctx, q, comment, actorUserID, id)
	if err != nil {
		return false, fmt.Errorf("cancel purchase request: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("cancel purchase request rows affected: %w", err)
	}
	return affected > 0, nil
}

const purchaseRequestSelect = `
	SELECT
		pr.id,
		pr.part_id,
		p.part_id,
		p.name,
		p.category,
		pr.quantity,
		pr.status,
		pr.source_part_request_id,
		pr.comment,
		pr.created_by_user_id,
		cu.email,
		NULLIF(TRIM(CONCAT_WS(' ', cu.last_name, cu.first_name, cu.middle_name)), '') AS created_by_full_name,
		pr.confirmed_by_user_id,
		au.email,
		NULLIF(TRIM(CONCAT_WS(' ', au.last_name, au.first_name, au.middle_name)), '') AS confirmed_by_full_name,
		pr.confirmed_at,
		pr.created_at,
		pr.updated_at
	FROM part_purchase_requests pr
	INNER JOIN parts_catalog p ON p.id = pr.part_id
	LEFT JOIN users cu ON cu.id = pr.created_by_user_id
	LEFT JOIN users au ON au.id = pr.confirmed_by_user_id
`

type purchaseRequestScanner interface {
	Scan(dest ...any) error
}

func scanPurchaseRequest(scanner purchaseRequestScanner) (*models.PurchaseRequest, error) {
	var item models.PurchaseRequest
	var sourcePartRequestID sql.NullInt64
	var comment sql.NullString
	var createdByEmail sql.NullString
	var createdByFullName sql.NullString
	var confirmedByUserID sql.NullInt64
	var confirmedByEmail sql.NullString
	var confirmedByFullName sql.NullString
	var confirmedAt sql.NullTime

	if err := scanner.Scan(
		&item.ID,
		&item.PartID,
		&item.PartCatalogCode,
		&item.PartName,
		&item.PartCategory,
		&item.Quantity,
		&item.Status,
		&sourcePartRequestID,
		&comment,
		&item.CreatedByUserID,
		&createdByEmail,
		&createdByFullName,
		&confirmedByUserID,
		&confirmedByEmail,
		&confirmedByFullName,
		&confirmedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.SourcePartRequestID = nullableInt64PtrPart(sourcePartRequestID)
	item.Comment = nullableStringPtr(comment)
	item.CreatedByEmail = nullableStringPtr(createdByEmail)
	item.CreatedByFullName = nullableStringPtr(createdByFullName)
	item.ConfirmedByUserID = nullableInt64PtrPart(confirmedByUserID)
	item.ConfirmedByEmail = nullableStringPtr(confirmedByEmail)
	item.ConfirmedByFullName = nullableStringPtr(confirmedByFullName)
	item.ConfirmedAt = nullableTimePtrPart(confirmedAt)

	return &item, nil
}

func scanPurchaseRequestRows(rows *sql.Rows) (*models.PurchaseRequest, error) {
	return scanPurchaseRequest(rows)
}

func normalizePurchaseRequestSortBy(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "part_id":
		return "pr.part_id"
	case "quantity":
		return "pr.quantity"
	case "status":
		return "pr.status"
	case "updated_at":
		return "pr.updated_at"
	default:
		return "pr.created_at"
	}
}

func mapPurchaseRequestError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "part_purchase_requests_part_id_fkey"):
		return ErrPartRequestPartNotFound
	case strings.Contains(msg, "part_purchase_requests_source_part_request_id_fkey"):
		return ErrPurchaseRequestSourcePartRequestNotFound
	case strings.Contains(msg, "part_purchase_requests_created_by_user_id_fkey"):
		return ErrPartRequestUserNotFound
	default:
		return err
	}
}
