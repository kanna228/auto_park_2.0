package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"auto_park/modules/invoice_module/dto"
)

var ErrInvoiceNumberExists = errors.New("invoice_number already exists")
var ErrInvoicePartRequestNotFound = errors.New("part request not found")

type CreateInvoiceParams struct {
	InvoiceNumber string
	InvoiceDate   string
	PartRequestID *int64
	RequestNumber string
}

type InvoiceRepository interface {
	Create(ctx context.Context, p CreateInvoiceParams) (*dto.InvoiceResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.InvoiceResponse, error)
	GetByPartRequestID(ctx context.Context, partRequestID int64) (*dto.InvoiceResponse, error)
	List(ctx context.Context, q dto.InvoiceListQuery) ([]dto.InvoiceResponse, int64, error)
	NextInvoiceNumber(ctx context.Context) (string, error)
}

type invoiceRepo struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepo{db: db}
}

func (r *invoiceRepo) Create(ctx context.Context, p CreateInvoiceParams) (*dto.InvoiceResponse, error) {
	const q = `
		INSERT INTO invoices (invoice_number, invoice_date, part_request_id, request_number, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, invoice_number, invoice_date, part_request_id, request_number, created_at;
	`
	item, err := scanInvoice(r.db.QueryRowContext(ctx, q, p.InvoiceNumber, p.InvoiceDate, p.PartRequestID, p.RequestNumber))
	if err != nil {
		return nil, mapInvoiceError(err)
	}
	return item, nil
}

func (r *invoiceRepo) GetByID(ctx context.Context, id int64) (*dto.InvoiceResponse, error) {
	const q = `
		SELECT id, invoice_number, invoice_date, part_request_id, request_number, created_at
		FROM invoices
		WHERE id = $1;
	`
	item, err := scanInvoice(r.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invoice by id: %w", err)
	}
	return item, nil
}

func (r *invoiceRepo) GetByPartRequestID(ctx context.Context, partRequestID int64) (*dto.InvoiceResponse, error) {
	const q = `
		SELECT id, invoice_number, invoice_date, part_request_id, request_number, created_at
		FROM invoices
		WHERE part_request_id = $1
		ORDER BY id DESC
		LIMIT 1;
	`
	item, err := scanInvoice(r.db.QueryRowContext(ctx, q, partRequestID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invoice by part request id: %w", err)
	}
	return item, nil
}

func (r *invoiceRepo) List(ctx context.Context, q dto.InvoiceListQuery) ([]dto.InvoiceResponse, int64, error) {
	limit := normalizeLimit(q.Limit)
	offset := normalizeOffset(q.Offset)

	conds := []string{"1=1"}
	args := make([]any, 0, 6)
	argPos := 1
	if v := strings.TrimSpace(q.Date); v != "" {
		conds = append(conds, fmt.Sprintf("invoice_date = $%d", argPos))
		args = append(args, v)
		argPos++
	}
	if v := strings.TrimSpace(q.Search); v != "" {
		conds = append(conds, fmt.Sprintf("(invoice_number ILIKE $%d OR request_number ILIKE $%d)", argPos, argPos))
		args = append(args, "%"+v+"%")
		argPos++
	}
	whereSQL := " WHERE " + strings.Join(conds, " AND ")

	countQ := `SELECT COUNT(*) FROM invoices` + whereSQL
	var total int64
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("list invoices count: %w", err)
	}

	listQ := fmt.Sprintf(`
		SELECT id, invoice_number, invoice_date, part_request_id, request_number, created_at
		FROM invoices
		%s
		ORDER BY invoice_date DESC, id DESC
		LIMIT $%d OFFSET $%d;
	`, whereSQL, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list invoices: %w", err)
	}
	defer rows.Close()

	items := make([]dto.InvoiceResponse, 0)
	for rows.Next() {
		item, err := scanInvoice(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list invoices scan: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list invoices rows: %w", err)
	}
	return items, total, nil
}

func (r *invoiceRepo) NextInvoiceNumber(ctx context.Context) (string, error) {
	var n int64
	if err := r.db.QueryRowContext(ctx, `SELECT nextval('invoice_number_seq');`).Scan(&n); err != nil {
		return "", fmt.Errorf("next invoice number: %w", err)
	}
	return fmt.Sprintf("A%06d", n), nil
}

type invoiceScanner interface {
	Scan(dest ...any) error
}

func scanInvoice(scanner invoiceScanner) (*dto.InvoiceResponse, error) {
	var item dto.InvoiceResponse
	var invoiceDate time.Time
	var partRequestID sql.NullInt64
	if err := scanner.Scan(
		&item.ID,
		&item.InvoiceNumber,
		&invoiceDate,
		&partRequestID,
		&item.RequestNumber,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	item.InvoiceDate = invoiceDate.Format("2006-01-02")
	item.PartRequestID = nullableInt64Ptr(partRequestID)
	return &item, nil
}

func nullableInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	out := v.Int64
	return &out
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

func mapInvoiceError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "invoices_invoice_number_key"):
		return ErrInvoiceNumberExists
	case strings.Contains(msg, "invoices_part_request_id_fkey"):
		return ErrInvoicePartRequestNotFound
	default:
		return err
	}
}
