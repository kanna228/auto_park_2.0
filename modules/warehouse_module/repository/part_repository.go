package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/internal/apierrors"
	"auto_park/modules/warehouse_module/models"

	"github.com/lib/pq"
)

var ErrPartIDExists = errors.New("part_id already exists")
var ErrPartNameExists = errors.New("part name already exists")

type CreatePartParams struct {
	PartID           string
	Name             string
	Quantity         int64
	MinStockQuantity int64
	Unit             string
	Price            float64
	Category         string
	Dimensions       *string
	Manufacturer     *string
	IsConsumable     bool
}

type UpdatePartParams struct {
	Name             string
	Quantity         int64
	MinStockQuantity int64
	Unit             string
	Price            float64
	Category         string
	Dimensions       *string
	Manufacturer     *string
	IsConsumable     bool
}

type ListPartsParams struct {
	PartID          string
	Name            string
	Category        string
	Limit           int
	Offset          int
	SortBy          string
	Order           string
	IncludeArchived bool
}

type ListPartMovementsParams struct {
	PartID int64
	Limit  int
	Offset int
}

type PartStockMovementRow struct {
	Type           string
	Quantity       int64
	Vehicle        *string
	PartRequestID  *int64
	DocumentNumber *string
	CreatedAt      time.Time
	Actor          *string
}

type PartSummaryRow struct {
	Total           int64
	LowCount        int64
	CriticalCount   int64
	IssuedLastMonth int64
}

type PartArrivalCreateItemParams struct {
	PartID   int64
	Quantity int64
	Price    *float64
}

type CreatePartArrivalParams struct {
	DocumentNumber  string
	ArrivalDate     string
	Comment         *string
	CreatedByUserID int64
	Items           []PartArrivalCreateItemParams
}

type ListPartArrivalsParams struct {
	DateFrom string
	DateTo   string
	Status   string
	Limit    int
	Offset   int
}

type PartArrivalItemRow struct {
	ID          int64
	PartID      int64
	PartCode    string
	PartName    string
	Quantity    int64
	Price       *float64
	TotalAmount *float64
}

type PartArrivalRow struct {
	ID             int64
	DocumentNumber string
	ArrivalDate    time.Time
	Status         string
	Comment        *string
	CreatedBy      *string
	AcceptedBy     *string
	AcceptedAt     *time.Time
	Items          []PartArrivalItemRow
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PartArrivalSummaryRow struct {
	Total         int64
	DraftCount    int64
	AcceptedCount int64
	AcceptedItems int64
}

type PartRepository interface {
	Create(ctx context.Context, p CreatePartParams) (int64, error)
	GetByID(ctx context.Context, id int64, includeArchived ...bool) (*models.Part, error)
	List(ctx context.Context, p ListPartsParams) ([]models.Part, int64, error)
	UpdateByID(ctx context.Context, id int64, p UpdatePartParams) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	Summary(ctx context.Context) (PartSummaryRow, error)
	ListMovements(ctx context.Context, p ListPartMovementsParams) ([]PartStockMovementRow, int64, error)
	CreateArrival(ctx context.Context, p CreatePartArrivalParams) (int64, error)
	ListArrivals(ctx context.Context, p ListPartArrivalsParams) ([]PartArrivalRow, int64, error)
	AcceptArrival(ctx context.Context, id int64, actorUserID int64) (bool, error)
	ArrivalSummary(ctx context.Context) (PartArrivalSummaryRow, error)
}

type partRepo struct {
	db *sql.DB
}

func NewPartRepository(db *sql.DB) PartRepository {
	return &partRepo{db: db}
}

func (r *partRepo) Create(ctx context.Context, p CreatePartParams) (int64, error) {
	const q = `
		INSERT INTO parts_catalog (part_id, name, quantity, min_stock_quantity, unit, price, category, dimensions, manufacturer, is_consumable, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id;
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, p.PartID, p.Name, p.Quantity, p.MinStockQuantity, p.Unit, p.Price, p.Category, p.Dimensions, p.Manufacturer, p.IsConsumable).Scan(&id); err != nil {
		return 0, mapConstraintError(err)
	}
	return id, nil
}

func (r *partRepo) GetByID(ctx context.Context, id int64, includeArchived ...bool) (*models.Part, error) {
	archiveCond := "AND is_archived = FALSE"
	if len(includeArchived) > 0 && includeArchived[0] {
		archiveCond = ""
	}
	q := fmt.Sprintf(`
		SELECT id, part_id, name, quantity, min_stock_quantity, unit, price, category, dimensions, manufacturer, is_consumable, is_archived, deleted_at, created_at, updated_at
		FROM parts_catalog
		WHERE id = $1
		  %s;
	`, archiveCond)

	var item models.Part
	var deletedAt sql.NullTime
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&item.ID,
		&item.PartID,
		&item.Name,
		&item.Quantity,
		&item.MinStockQuantity,
		&item.Unit,
		&item.Price,
		&item.Category,
		&item.Dimensions,
		&item.Manufacturer,
		&item.IsConsumable,
		&item.IsArchived,
		&deletedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get part by id: %w", err)
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}

func (r *partRepo) List(ctx context.Context, p ListPartsParams) ([]models.Part, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	sortBy := normalizeWarehousePartSortBy(p.SortBy)
	order := normalizeOrder(p.Order)

	conds := make([]string, 0, 3)
	args := make([]any, 0, 8)
	argPos := 1

	if v := strings.TrimSpace(p.PartID); v != "" {
		conds = append(conds, fmt.Sprintf("part_id ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(p.Name); v != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if v := strings.TrimSpace(p.Category); v != "" {
		conds = append(conds, fmt.Sprintf("category ILIKE $%d", argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	if !p.IncludeArchived {
		conds = append(conds, "is_archived = FALSE")
	}

	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = " WHERE " + strings.Join(conds, " AND ")
	}

	countQ := `SELECT COUNT(*) FROM parts_catalog` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list parts count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT id, part_id, name, quantity, min_stock_quantity, unit, price, category, dimensions, manufacturer, is_consumable, is_archived, deleted_at, created_at, updated_at
		FROM parts_catalog
		%s
		ORDER BY %s %s, id ASC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, sortBy, order, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list parts: %w", err)
	}
	defer rows.Close()

	items := make([]models.Part, 0)
	for rows.Next() {
		var item models.Part
		var deletedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.PartID,
			&item.Name,
			&item.Quantity,
			&item.MinStockQuantity,
			&item.Unit,
			&item.Price,
			&item.Category,
			&item.Dimensions,
			&item.Manufacturer,
			&item.IsConsumable,
			&item.IsArchived,
			&deletedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("list parts scan: %w", err)
		}
		if deletedAt.Valid {
			item.DeletedAt = &deletedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list parts rows: %w", err)
	}

	return items, total, nil
}

func (r *partRepo) UpdateByID(ctx context.Context, id int64, p UpdatePartParams) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin update part tx: %w", err)
	}
	defer rollbackPartTx(tx)

	var currentQuantity int64
	const currentQ = `SELECT quantity FROM parts_catalog WHERE id = $1 AND is_archived = FALSE FOR UPDATE;`
	if err := tx.QueryRowContext(ctx, currentQ, id).Scan(&currentQuantity); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get part quantity before update: %w", err)
	}

	const q = `
		UPDATE parts_catalog
		SET name = $1,
			quantity = $2,
			min_stock_quantity = $3,
			unit = $4,
			price = $5,
			category = $6,
			dimensions = $7,
			manufacturer = $8,
			is_consumable = $9,
			updated_at = NOW()
		WHERE id = $10
		  AND is_archived = FALSE;
	`
	res, err := tx.ExecContext(ctx, q, p.Name, p.Quantity, p.MinStockQuantity, p.Unit, p.Price, p.Category, p.Dimensions, p.Manufacturer, p.IsConsumable, id)
	if err != nil {
		return false, mapConstraintError(err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update part rows affected: %w", err)
	}
	if aff == 0 {
		return false, nil
	}

	documentNumber := "manual-adjustment"
	switch {
	case p.Quantity > currentQuantity:
		if err := insertPartStockMovementTx(ctx, tx, id, "arrival", p.Quantity-currentQuantity, nil, nil, &documentNumber, 0); err != nil {
			return false, err
		}
	case p.Quantity < currentQuantity:
		if err := insertPartStockMovementTx(ctx, tx, id, "writeoff", currentQuantity-p.Quantity, nil, nil, &documentNumber, 0); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit update part tx: %w", err)
	}
	return true, nil
}

func (r *partRepo) DeleteByID(ctx context.Context, id int64) (bool, error) {
	const q = `
		UPDATE parts_catalog
		SET is_archived = TRUE,
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
		  AND is_archived = FALSE;
	`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return false, fmt.Errorf("delete part by id: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete part rows affected: %w", err)
	}
	return aff > 0, nil
}

func (r *partRepo) Summary(ctx context.Context) (PartSummaryRow, error) {
	const q = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE quantity > 0 AND quantity <= min_stock_quantity) AS low_count,
			COUNT(*) FILTER (WHERE quantity <= 0) AS critical_count,
			COALESCE((
				SELECT SUM(quantity)
				FROM part_stock_movements
				WHERE type = 'issue'
				  AND created_at >= NOW() - INTERVAL '1 month'
			), 0) AS issued_last_month
		FROM parts_catalog
		WHERE is_archived = FALSE;
	`
	var row PartSummaryRow
	if err := r.db.QueryRowContext(ctx, q).Scan(&row.Total, &row.LowCount, &row.CriticalCount, &row.IssuedLastMonth); err != nil {
		return PartSummaryRow{}, fmt.Errorf("parts summary: %w", err)
	}
	return row, nil
}

func (r *partRepo) ListMovements(ctx context.Context, p ListPartMovementsParams) ([]PartStockMovementRow, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	const countQ = `SELECT COUNT(*) FROM part_stock_movements WHERE part_id = $1;`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, p.PartID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list part movements count: %w", err)
	}

	const q = `
		SELECT
			m.type,
			m.quantity,
			NULLIF(TRIM(CONCAT_WS(' ', v.state_number, v.brand_model)), '') AS vehicle,
			m.part_request_id,
			m.document_number,
			m.created_at,
			NULLIF(TRIM(CONCAT_WS(' ', u.last_name, u.first_name, u.middle_name)), '') AS actor
		FROM part_stock_movements m
		LEFT JOIN vehicles v ON v.id = m.vehicle_id
		LEFT JOIN users u ON u.id = m.actor_user_id
		WHERE m.part_id = $1
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.db.QueryContext(ctx, q, p.PartID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list part movements: %w", err)
	}
	defer rows.Close()

	items := make([]PartStockMovementRow, 0)
	for rows.Next() {
		var item PartStockMovementRow
		var vehicle, documentNumber, actor sql.NullString
		var partRequestID sql.NullInt64
		if err := rows.Scan(&item.Type, &item.Quantity, &vehicle, &partRequestID, &documentNumber, &item.CreatedAt, &actor); err != nil {
			return nil, 0, fmt.Errorf("list part movements scan: %w", err)
		}
		item.Vehicle = nullableStringPtr(vehicle)
		item.PartRequestID = nullableInt64PtrPart(partRequestID)
		item.DocumentNumber = nullableStringPtr(documentNumber)
		item.Actor = nullableStringPtr(actor)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list part movements rows: %w", err)
	}
	return items, total, nil
}

func (r *partRepo) CreateArrival(ctx context.Context, p CreatePartArrivalParams) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin create part arrival: %w", err)
	}
	defer rollbackPartTx(tx)

	const insertArrivalQ = `
		INSERT INTO part_arrivals (document_number, arrival_date, comment, created_by_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id;
	`
	var id int64
	if err := tx.QueryRowContext(ctx, insertArrivalQ, p.DocumentNumber, p.ArrivalDate, p.Comment, p.CreatedByUserID).Scan(&id); err != nil {
		return 0, err
	}

	const insertItemQ = `
		INSERT INTO part_arrival_items (arrival_id, part_id, quantity, price, created_at)
		VALUES ($1, $2, $3, $4, NOW());
	`
	for _, item := range p.Items {
		if err := ensurePartExistsTx(ctx, tx, item.PartID); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, insertItemQ, id, item.PartID, item.Quantity, item.Price); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit create part arrival: %w", err)
	}
	return id, nil
}

func (r *partRepo) ListArrivals(ctx context.Context, p ListPartArrivalsParams) ([]PartArrivalRow, int64, error) {
	limit := p.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}

	conds := []string{"1=1"}
	args := make([]any, 0, 6)
	argPos := 1
	if v := strings.TrimSpace(p.DateFrom); v != "" {
		conds = append(conds, fmt.Sprintf("a.arrival_date >= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.DateTo); v != "" {
		conds = append(conds, fmt.Sprintf("a.arrival_date <= $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(p.Status); v != "" {
		conds = append(conds, fmt.Sprintf("a.status = $%d", argPos))
		args = append(args, strings.ToLower(v))
		argPos++
	}
	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `SELECT COUNT(*) FROM part_arrivals a` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list part arrivals count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT
			a.id,
			a.document_number,
			a.arrival_date,
			a.status,
			a.comment,
			NULLIF(TRIM(CONCAT_WS(' ', cu.last_name, cu.first_name, cu.middle_name)), '') AS created_by,
			NULLIF(TRIM(CONCAT_WS(' ', au.last_name, au.first_name, au.middle_name)), '') AS accepted_by,
			a.accepted_at,
			a.created_at,
			a.updated_at
		FROM part_arrivals a
		LEFT JOIN users cu ON cu.id = a.created_by_user_id
		LEFT JOIN users au ON au.id = a.accepted_by_user_id
		%s
		ORDER BY a.arrival_date DESC, a.id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list part arrivals: %w", err)
	}
	defer rows.Close()

	items := make([]PartArrivalRow, 0)
	ids := make([]int64, 0)
	indexByID := make(map[int64]int)
	for rows.Next() {
		var item PartArrivalRow
		var comment, createdBy, acceptedBy sql.NullString
		var acceptedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.DocumentNumber, &item.ArrivalDate, &item.Status, &comment, &createdBy, &acceptedBy, &acceptedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("list part arrivals scan: %w", err)
		}
		item.Comment = nullableStringPtr(comment)
		item.CreatedBy = nullableStringPtr(createdBy)
		item.AcceptedBy = nullableStringPtr(acceptedBy)
		item.AcceptedAt = nullableTimePtrPart(acceptedAt)
		item.Items = []PartArrivalItemRow{}
		indexByID[item.ID] = len(items)
		ids = append(ids, item.ID)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list part arrivals rows: %w", err)
	}

	if len(ids) > 0 {
		arrivalItems, err := r.listArrivalItems(ctx, ids)
		if err != nil {
			return nil, 0, err
		}
		for arrivalID, rows := range arrivalItems {
			if idx, ok := indexByID[arrivalID]; ok {
				items[idx].Items = rows
			}
		}
	}

	return items, total, nil
}

func (r *partRepo) AcceptArrival(ctx context.Context, id int64, actorUserID int64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin accept part arrival: %w", err)
	}
	defer rollbackPartTx(tx)

	var documentNumber string
	var status string
	const lockQ = `SELECT document_number, status FROM part_arrivals WHERE id = $1 FOR UPDATE;`
	if err := tx.QueryRowContext(ctx, lockQ, id).Scan(&documentNumber, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get part arrival for accept: %w", err)
	}
	if status == "accepted" {
		return true, nil
	}

	const itemsQ = `SELECT part_id, quantity FROM part_arrival_items WHERE arrival_id = $1 ORDER BY id ASC;`
	rows, err := tx.QueryContext(ctx, itemsQ, id)
	if err != nil {
		return false, fmt.Errorf("list part arrival items for accept: %w", err)
	}
	defer rows.Close()
	type arrivalQty struct {
		partID int64
		qty    int64
	}
	items := make([]arrivalQty, 0)
	for rows.Next() {
		var item arrivalQty
		if err := rows.Scan(&item.partID, &item.qty); err != nil {
			return false, fmt.Errorf("scan part arrival items for accept: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("rows part arrival items for accept: %w", err)
	}
	if len(items) == 0 {
		return false, fmt.Errorf("arrival has no items")
	}

	for _, item := range items {
		if err := increasePartQuantityPartTx(ctx, tx, item.partID, item.qty); err != nil {
			return false, err
		}
		if err := insertPartStockMovementTx(ctx, tx, item.partID, "arrival", item.qty, nil, nil, &documentNumber, actorUserID); err != nil {
			return false, err
		}
	}

	const updateQ = `
		UPDATE part_arrivals
		SET status = 'accepted',
			accepted_by_user_id = $1,
			accepted_at = NOW(),
			updated_at = NOW()
		WHERE id = $2;
	`
	if _, err := tx.ExecContext(ctx, updateQ, actorUserID, id); err != nil {
		return false, fmt.Errorf("accept part arrival update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit accept part arrival: %w", err)
	}
	return true, nil
}

func (r *partRepo) ArrivalSummary(ctx context.Context) (PartArrivalSummaryRow, error) {
	const q = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'draft') AS draft_count,
			COUNT(*) FILTER (WHERE status = 'accepted') AS accepted_count,
			COALESCE((
				SELECT SUM(i.quantity)
				FROM part_arrival_items i
				INNER JOIN part_arrivals a ON a.id = i.arrival_id
				WHERE a.status = 'accepted'
			), 0) AS accepted_items
		FROM part_arrivals;
	`
	var row PartArrivalSummaryRow
	if err := r.db.QueryRowContext(ctx, q).Scan(&row.Total, &row.DraftCount, &row.AcceptedCount, &row.AcceptedItems); err != nil {
		return PartArrivalSummaryRow{}, fmt.Errorf("part arrivals summary: %w", err)
	}
	return row, nil
}

func (r *partRepo) listArrivalItems(ctx context.Context, ids []int64) (map[int64][]PartArrivalItemRow, error) {
	const q = `
		SELECT
			i.arrival_id,
			i.id,
			i.part_id,
			p.part_id,
			p.name,
			i.quantity,
			i.price,
			CASE WHEN i.price IS NULL THEN NULL ELSE i.price * i.quantity END AS total_amount
		FROM part_arrival_items i
		INNER JOIN parts_catalog p ON p.id = i.part_id
		WHERE i.arrival_id = ANY($1)
		ORDER BY i.arrival_id ASC, i.id ASC;
	`
	rows, err := r.db.QueryContext(ctx, q, pq.Int64Array(ids))
	if err != nil {
		return nil, fmt.Errorf("list part arrival items: %w", err)
	}
	defer rows.Close()

	out := make(map[int64][]PartArrivalItemRow)
	for rows.Next() {
		var arrivalID int64
		var item PartArrivalItemRow
		var price, totalAmount sql.NullFloat64
		if err := rows.Scan(&arrivalID, &item.ID, &item.PartID, &item.PartCode, &item.PartName, &item.Quantity, &price, &totalAmount); err != nil {
			return nil, fmt.Errorf("list part arrival items scan: %w", err)
		}
		item.Price = nullableFloat64Ptr(price)
		item.TotalAmount = nullableFloat64Ptr(totalAmount)
		out[arrivalID] = append(out[arrivalID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list part arrival items rows: %w", err)
	}
	return out, nil
}

func normalizeWarehousePartSortBy(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "id":
		return "id"
	case "part_id":
		return "part_id"
	case "name":
		return "name"
	case "category":
		return "category"
	case "quantity":
		return "quantity"
	case "min_stock_quantity":
		return "min_stock_quantity"
	case "price":
		return "price"
	case "created_at":
		return "created_at"
	default:
		return "created_at"
	}
}

func normalizeSortBy(v string) string {
	return normalizeWarehousePartSortBy(v)
}

func ensurePartExistsTx(ctx context.Context, tx *sql.Tx, partID int64) error {
	const q = `SELECT EXISTS(SELECT 1 FROM parts_catalog WHERE id = $1 AND is_archived = FALSE);`
	var exists bool
	if err := tx.QueryRowContext(ctx, q, partID).Scan(&exists); err != nil {
		return fmt.Errorf("check part exists: %w", err)
	}
	if !exists {
		const archivedQ = `SELECT EXISTS(SELECT 1 FROM parts_catalog WHERE id = $1 AND is_archived = TRUE);`
		var archived bool
		if err := tx.QueryRowContext(ctx, archivedQ, partID).Scan(&archived); err != nil {
			return fmt.Errorf("check part archived: %w", err)
		}
		if archived {
			return apierrors.ErrEntityArchived
		}
		return ErrPartRequestPartNotFound
	}
	return nil
}

func increasePartQuantityPartTx(ctx context.Context, tx *sql.Tx, partID, qty int64) error {
	const q = `
		UPDATE parts_catalog
		SET quantity = quantity + $1,
			updated_at = NOW()
		WHERE id = $2
		  AND is_archived = FALSE;
	`
	_, err := tx.ExecContext(ctx, q, qty, partID)
	if err != nil {
		return fmt.Errorf("increase part quantity: %w", err)
	}
	return nil
}

func insertPartStockMovementTx(ctx context.Context, tx *sql.Tx, partID int64, movementType string, quantity int64, vehicleID *int64, partRequestID *int64, documentNumber *string, actorUserID int64) error {
	const q = `
		INSERT INTO part_stock_movements (
			part_id,
			type,
			quantity,
			vehicle_id,
			part_request_id,
			document_number,
			actor_user_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());
	`
	_, err := tx.ExecContext(ctx, q, partID, movementType, quantity, vehicleID, partRequestID, documentNumber, nullableUserID(actorUserID))
	if err != nil {
		return fmt.Errorf("insert part stock movement: %w", err)
	}
	return nil
}

func nullableUserID(id int64) any {
	if id <= 0 {
		return nil
	}
	return id
}

func rollbackPartTx(tx *sql.Tx) {
	_ = tx.Rollback()
}

func nullableInt64PtrPart(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
}

func nullableTimePtrPart(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	out := v.Time
	return &out
}

func nullableFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	out := v.Float64
	return &out
}

func normalizeOrder(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "ASC"
	}
	return "DESC"
}

func mapConstraintError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "parts_catalog_part_id_key"):
		return ErrPartIDExists
	case strings.Contains(msg, "parts_catalog_name_key"):
		return ErrPartNameExists
	default:
		return err
	}
}
