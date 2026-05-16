package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"auto_park/modules/invoice_module/dto"
	"auto_park/modules/invoice_module/repository"
)

type InvoiceService interface {
	Create(ctx context.Context, req dto.InvoiceCreateRequest) (*dto.InvoiceResponse, error)
	CreateForPartRequest(ctx context.Context, partRequestID int64, requestNumber string) (*dto.InvoiceResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.InvoiceResponse, error)
	List(ctx context.Context, q dto.InvoiceListQuery) (*dto.InvoiceListResponse, error)
	ExportCSV(ctx context.Context, q dto.InvoiceListQuery) ([]byte, error)
}

type invoiceService struct {
	repo repository.InvoiceRepository
}

func NewInvoiceService(repo repository.InvoiceRepository) InvoiceService {
	return &invoiceService{repo: repo}
}

func (s *invoiceService) Create(ctx context.Context, req dto.InvoiceCreateRequest) (*dto.InvoiceResponse, error) {
	p, err := validateInvoiceCreate(req)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, p)
}

func (s *invoiceService) CreateForPartRequest(ctx context.Context, partRequestID int64, requestNumber string) (*dto.InvoiceResponse, error) {
	if partRequestID <= 0 {
		return nil, fmt.Errorf("part_request_id is required")
	}
	if existing, err := s.repo.GetByPartRequestID(ctx, partRequestID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	requestNumber = strings.TrimSpace(requestNumber)
	if requestNumber == "" {
		requestNumber = strconv.FormatInt(partRequestID, 10)
	}
	partRequestIDPtr := partRequestID
	for i := 0; i < 3; i++ {
		invoiceNumber, err := s.repo.NextInvoiceNumber(ctx)
		if err != nil {
			return nil, err
		}
		item, err := s.repo.Create(ctx, repository.CreateInvoiceParams{
			InvoiceNumber: invoiceNumber,
			InvoiceDate:   time.Now().Format("2006-01-02"),
			PartRequestID: &partRequestIDPtr,
			RequestNumber: requestNumber,
		})
		if err == nil {
			return item, nil
		}
		if !errors.Is(err, repository.ErrInvoiceNumberExists) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("failed to generate unique invoice number")
}

func (s *invoiceService) GetByID(ctx context.Context, id int64) (*dto.InvoiceResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *invoiceService) List(ctx context.Context, q dto.InvoiceListQuery) (*dto.InvoiceListResponse, error) {
	if err := validateInvoiceQuery(q); err != nil {
		return nil, err
	}
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}
	return &dto.InvoiceListResponse{Items: items, Total: total, Limit: normalizeLimit(q.Limit), Offset: normalizeOffset(q.Offset)}, nil
}

func (s *invoiceService) ExportCSV(ctx context.Context, q dto.InvoiceListQuery) ([]byte, error) {
	if err := validateInvoiceQuery(q); err != nil {
		return nil, err
	}
	q.Limit = 10000
	q.Offset = 0
	items, _, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"id", "invoice_number", "invoice_date", "part_request_id", "request_number", "created_at"}); err != nil {
		return nil, err
	}
	for _, item := range items {
		if err := w.Write([]string{
			fmt.Sprint(item.ID),
			item.InvoiceNumber,
			item.InvoiceDate,
			formatOptionalInt64(item.PartRequestID),
			item.RequestNumber,
			item.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func validateInvoiceCreate(req dto.InvoiceCreateRequest) (repository.CreateInvoiceParams, error) {
	invoiceNumber := strings.TrimSpace(req.InvoiceNumber)
	if invoiceNumber == "" {
		return repository.CreateInvoiceParams{}, fmt.Errorf("invoice_number is required")
	}
	if len([]rune(invoiceNumber)) > 32 {
		return repository.CreateInvoiceParams{}, fmt.Errorf("invoice_number must be less than or equal to 32 characters")
	}
	invoiceDate := strings.TrimSpace(req.InvoiceDate)
	if _, err := time.Parse("2006-01-02", invoiceDate); err != nil {
		return repository.CreateInvoiceParams{}, fmt.Errorf("invoice_date must be in YYYY-MM-DD format")
	}
	if req.PartRequestID != nil && *req.PartRequestID <= 0 {
		return repository.CreateInvoiceParams{}, fmt.Errorf("part_request_id must be greater than 0")
	}
	requestNumber := strings.TrimSpace(req.RequestNumber)
	if requestNumber == "" {
		return repository.CreateInvoiceParams{}, fmt.Errorf("request_number is required")
	}
	if len([]rune(requestNumber)) > 32 {
		return repository.CreateInvoiceParams{}, fmt.Errorf("request_number must be less than or equal to 32 characters")
	}
	return repository.CreateInvoiceParams{
		InvoiceNumber: invoiceNumber,
		InvoiceDate:   invoiceDate,
		PartRequestID: req.PartRequestID,
		RequestNumber: requestNumber,
	}, nil
}

func validateInvoiceQuery(q dto.InvoiceListQuery) error {
	if v := strings.TrimSpace(q.Date); v != "" {
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return fmt.Errorf("date must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func formatOptionalInt64(value *int64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(*value, 10)
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
