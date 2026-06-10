package service

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"auto_park/internal/apierrors"
	"auto_park/internal/excelreport"
	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/models"
	"auto_park/modules/warehouse_module/repository"

	"github.com/xuri/excelize/v2"
)

var allowedTextRe = regexp.MustCompile(`^[A-Za-zА-Яа-яЁё0-9\s\-_/().,xX№]+$`)

type PartService interface {
	Create(ctx context.Context, req dto.PartCreateRequest) (int64, error)
	GetByID(ctx context.Context, id int64, includeArchived ...bool) (*dto.PartResponse, error)
	List(ctx context.Context, q dto.PartListQuery) (*dto.PartListResponse, error)
	UpdateByID(ctx context.Context, id int64, req dto.PartUpdateRequest) (bool, error)
	DeleteByID(ctx context.Context, id int64) (bool, error)
	Summary(ctx context.Context) (*dto.PartSummaryResponse, error)
	ListMovements(ctx context.Context, partID int64, limit int, offset int) (*dto.PartStockMovementListResponse, error)
	CreateArrival(ctx context.Context, userID int64, req dto.PartArrivalCreateRequest) (int64, error)
	ListArrivals(ctx context.Context, dateFrom string, dateTo string, status string, limit int, offset int) (*dto.PartArrivalListResponse, error)
	AcceptArrival(ctx context.Context, id int64, userID int64) (bool, error)
	ArrivalSummary(ctx context.Context) (*dto.PartArrivalSummaryResponse, error)
	Export(ctx context.Context, q dto.PartListQuery, format string) ([]byte, string, error)
}

type partService struct {
	repo repository.PartRepository
}

func NewPartService(repo repository.PartRepository) PartService {
	return &partService{repo: repo}
}

func (s *partService) Create(ctx context.Context, req dto.PartCreateRequest) (int64, error) {
	partID, err := normalizeRequiredField("part_id", req.PartID)
	if err != nil {
		return 0, err
	}
	name, err := normalizeRequiredField("name", req.Name)
	if err != nil {
		return 0, err
	}
	category, err := normalizeRequiredField("category", req.Category)
	if err != nil {
		return 0, err
	}
	if req.StartQuantity < 0 {
		return 0, fmt.Errorf("start_quantity must be greater than or equal to 0")
	}
	if req.MinStockQuantity < 0 {
		return 0, fmt.Errorf("min_stock_quantity must be greater than or equal to 0")
	}
	unit, err := normalizeUnit(req.Unit)
	if err != nil {
		return 0, err
	}
	if req.Price < 0 {
		return 0, fmt.Errorf("price must be greater than or equal to 0")
	}
	dimensions, err := normalizeOptionalField("dimensions", req.Dimensions)
	if err != nil {
		return 0, err
	}
	manufacturer, err := normalizeOptionalField("manufacturer", req.Manufacturer)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, repository.CreatePartParams{
		PartID:           partID,
		Name:             name,
		Quantity:         req.StartQuantity,
		MinStockQuantity: req.MinStockQuantity,
		Unit:             unit,
		Price:            req.Price,
		Category:         category,
		Dimensions:       dimensions,
		Manufacturer:     manufacturer,
		IsConsumable:     req.IsConsumable,
	})
}

func (s *partService) GetByID(ctx context.Context, id int64, includeArchived ...bool) (*dto.PartResponse, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	item, err := s.repo.GetByID(ctx, id, includeArchived...)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}
	resp := mapPartToDTO(*item)
	return &resp, nil
}

func (s *partService) List(ctx context.Context, q dto.PartListQuery) (*dto.PartListResponse, error) {
	items, total, err := s.repo.List(ctx, repository.ListPartsParams{
		PartID:          strings.TrimSpace(q.PartID),
		Name:            strings.TrimSpace(q.Name),
		Category:        strings.TrimSpace(q.Category),
		Limit:           q.Limit,
		Offset:          q.Offset,
		SortBy:          q.SortBy,
		Order:           q.Order,
		IncludeArchived: q.IncludeArchived,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.PartResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartToDTO(item))
	}

	limit := q.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	return &dto.PartListResponse{Items: out, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *partService) UpdateByID(ctx context.Context, id int64, req dto.PartUpdateRequest) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	current, err := s.repo.GetByID(ctx, id, true)
	if err != nil {
		return false, err
	}
	if current == nil {
		return false, nil
	}
	if current.IsArchived {
		return false, apierrors.ErrEntityArchived
	}
	name, err := normalizeRequiredField("name", req.Name)
	if err != nil {
		return false, err
	}
	category, err := normalizeRequiredField("category", req.Category)
	if err != nil {
		return false, err
	}
	if req.Quantity < 0 {
		return false, fmt.Errorf("quantity must be greater than or equal to 0")
	}
	if req.MinStockQuantity < 0 {
		return false, fmt.Errorf("min_stock_quantity must be greater than or equal to 0")
	}
	unit, err := normalizeUnit(req.Unit)
	if err != nil {
		return false, err
	}
	if req.Price < 0 {
		return false, fmt.Errorf("price must be greater than or equal to 0")
	}
	dimensions, err := normalizeOptionalField("dimensions", req.Dimensions)
	if err != nil {
		return false, err
	}
	manufacturer, err := normalizeOptionalField("manufacturer", req.Manufacturer)
	if err != nil {
		return false, err
	}

	return s.repo.UpdateByID(ctx, id, repository.UpdatePartParams{
		Name:             name,
		Quantity:         req.Quantity,
		MinStockQuantity: req.MinStockQuantity,
		Unit:             unit,
		Price:            req.Price,
		Category:         category,
		Dimensions:       dimensions,
		Manufacturer:     manufacturer,
		IsConsumable:     req.IsConsumable,
	})
}

func (s *partService) DeleteByID(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *partService) Summary(ctx context.Context) (*dto.PartSummaryResponse, error) {
	row, err := s.repo.Summary(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.PartSummaryResponse{Total: row.Total, LowCount: row.LowCount, CriticalCount: row.CriticalCount, IssuedLastMonth: row.IssuedLastMonth}, nil
}

func (s *partService) ListMovements(ctx context.Context, partID int64, limit int, offset int) (*dto.PartStockMovementListResponse, error) {
	if partID <= 0 {
		return nil, fmt.Errorf("invalid part id")
	}
	items, total, err := s.repo.ListMovements(ctx, repository.ListPartMovementsParams{PartID: partID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	out := make([]dto.PartStockMovementResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.PartStockMovementResponse{
			Type:           item.Type,
			Quantity:       item.Quantity,
			Vehicle:        item.Vehicle,
			PartRequestID:  item.PartRequestID,
			DocumentNumber: item.DocumentNumber,
			CreatedAt:      item.CreatedAt,
			Actor:          item.Actor,
		})
	}
	return &dto.PartStockMovementListResponse{Items: out, Total: total, Limit: normalizeLimit(limit), Offset: normalizeOffset(offset)}, nil
}

func (s *partService) CreateArrival(ctx context.Context, userID int64, req dto.PartArrivalCreateRequest) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user id")
	}
	documentNumber, err := normalizeRequiredField("document_number", req.DocumentNumber)
	if err != nil {
		return 0, err
	}
	arrivalDate, err := normalizeWarehouseDate(req.ArrivalDate, "arrival_date")
	if err != nil {
		return 0, err
	}
	if len(req.Items) == 0 {
		return 0, fmt.Errorf("items is required")
	}
	comment, err := normalizeOptionalField("comment", req.Comment)
	if err != nil {
		return 0, err
	}
	items := make([]repository.PartArrivalCreateItemParams, 0, len(req.Items))
	for _, item := range req.Items {
		if item.PartID <= 0 {
			return 0, fmt.Errorf("part_id is required")
		}
		if item.Quantity <= 0 {
			return 0, fmt.Errorf("quantity must be greater than 0")
		}
		if item.Price != nil && *item.Price < 0 {
			return 0, fmt.Errorf("price must be greater than or equal to 0")
		}
		items = append(items, repository.PartArrivalCreateItemParams{PartID: item.PartID, Quantity: item.Quantity, Price: item.Price})
	}
	return s.repo.CreateArrival(ctx, repository.CreatePartArrivalParams{DocumentNumber: documentNumber, ArrivalDate: arrivalDate, Comment: comment, CreatedByUserID: userID, Items: items})
}

func (s *partService) ListArrivals(ctx context.Context, dateFrom string, dateTo string, status string, limit int, offset int) (*dto.PartArrivalListResponse, error) {
	normDateFrom, err := normalizeOptionalWarehouseDate(dateFrom, "date_from")
	if err != nil {
		return nil, err
	}
	normDateTo, err := normalizeOptionalWarehouseDate(dateTo, "date_to")
	if err != nil {
		return nil, err
	}
	normStatus := strings.ToLower(strings.TrimSpace(status))
	if normStatus != "" && normStatus != "draft" && normStatus != "accepted" {
		return nil, fmt.Errorf("invalid status")
	}
	items, total, err := s.repo.ListArrivals(ctx, repository.ListPartArrivalsParams{DateFrom: normDateFrom, DateTo: normDateTo, Status: normStatus, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	out := make([]dto.PartArrivalResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapPartArrivalToDTO(item))
	}
	return &dto.PartArrivalListResponse{Items: out, Total: total, Limit: normalizeLimit(limit), Offset: normalizeOffset(offset)}, nil
}

func (s *partService) AcceptArrival(ctx context.Context, id int64, userID int64) (bool, error) {
	if id <= 0 {
		return false, fmt.Errorf("invalid id")
	}
	if userID <= 0 {
		return false, fmt.Errorf("invalid user id")
	}
	return s.repo.AcceptArrival(ctx, id, userID)
}

func (s *partService) ArrivalSummary(ctx context.Context) (*dto.PartArrivalSummaryResponse, error) {
	row, err := s.repo.ArrivalSummary(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.PartArrivalSummaryResponse{Total: row.Total, DraftCount: row.DraftCount, AcceptedCount: row.AcceptedCount, AcceptedItems: row.AcceptedItems}, nil
}

func (s *partService) Export(ctx context.Context, q dto.PartListQuery, format string) ([]byte, string, error) {
	if strings.ToLower(strings.TrimSpace(format)) != "xlsx" {
		return nil, "", fmt.Errorf("unsupported export format")
	}
	q.Limit = 10000
	q.Offset = 0
	resp, err := s.List(ctx, q)
	if err != nil {
		return nil, "", err
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	styles := excelreport.NewStyles(f)
	headers := []string{"ID", "Артикул", "Наименование", "Количество", "Мин. остаток", "Ед.", "Цена", "Стоимость", "Категория", "Производитель"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}
	for rowIdx, item := range resp.Items {
		values := []any{item.ID, item.PartID, item.Name, item.Quantity, item.MinStockQuantity, item.Unit, item.Price, item.TotalValue, item.Category, derefPartString(item.Manufacturer)}
		for colIdx, value := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
	rowCount := len(resp.Items) + 1
	excelreport.ApplyTable(f, sheet, 1, rowCount, 10, styles, true)
	excelreport.FreezeBelow(f, sheet, 1)
	excelreport.SetWidths(f, sheet, []float64{10, 20, 38, 14, 14, 12, 16, 18, 24, 24})
	if rowCount > 1 {
		_ = f.SetCellStyle(sheet, "A2", fmt.Sprintf("A%d", rowCount), styles.Integer)
		_ = f.SetCellStyle(sheet, "D2", fmt.Sprintf("E%d", rowCount), styles.Integer)
		_ = f.SetCellStyle(sheet, "G2", fmt.Sprintf("H%d", rowCount), styles.Money)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "warehouse_parts.xlsx", nil
}

func normalizeRequiredField(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if !allowedTextRe.MatchString(trimmed) {
		return "", fmt.Errorf("%s contains invalid characters", field)
	}
	return trimmed, nil
}

func normalizeOptionalField(field string, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, nil
	}
	if !allowedTextRe.MatchString(trimmed) {
		return nil, fmt.Errorf("%s contains invalid characters", field)
	}
	return &trimmed, nil
}

func normalizeUnit(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "шт", nil
	}
	if !allowedTextRe.MatchString(trimmed) {
		return "", fmt.Errorf("unit contains invalid characters")
	}
	return trimmed, nil
}

func normalizeWarehouseDate(value, field string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return trimmed, nil
}

func normalizeOptionalWarehouseDate(value, field string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return normalizeWarehouseDate(value, field)
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

func mapPartToDTO(item models.Part) dto.PartResponse {
	return dto.PartResponse{
		ID:               item.ID,
		PartID:           item.PartID,
		Name:             item.Name,
		Quantity:         item.Quantity,
		MinStockQuantity: item.MinStockQuantity,
		Unit:             item.Unit,
		Price:            item.Price,
		TotalValue:       float64(item.Quantity) * item.Price,
		Category:         item.Category,
		Dimensions:       item.Dimensions,
		Manufacturer:     item.Manufacturer,
		IsConsumable:     item.IsConsumable,
		IsArchived:       item.IsArchived,
		DeletedAt:        item.DeletedAt,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func mapPartArrivalToDTO(item repository.PartArrivalRow) dto.PartArrivalResponse {
	items := make([]dto.PartArrivalItemResponse, 0, len(item.Items))
	for _, row := range item.Items {
		items = append(items, dto.PartArrivalItemResponse{
			ID:          row.ID,
			PartID:      row.PartID,
			PartCode:    row.PartCode,
			PartName:    row.PartName,
			Quantity:    row.Quantity,
			Price:       row.Price,
			TotalAmount: row.TotalAmount,
		})
	}
	return dto.PartArrivalResponse{
		ID:             item.ID,
		DocumentNumber: item.DocumentNumber,
		ArrivalDate:    item.ArrivalDate.Format("2006-01-02"),
		Status:         item.Status,
		Comment:        item.Comment,
		CreatedBy:      item.CreatedBy,
		AcceptedBy:     item.AcceptedBy,
		AcceptedAt:     item.AcceptedAt,
		Items:          items,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func derefPartString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
