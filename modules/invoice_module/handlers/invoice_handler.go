package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/modules/invoice_module/dto"
	"auto_park/modules/invoice_module/repository"
	"auto_park/modules/invoice_module/service"

	"github.com/gin-gonic/gin"
)

type InvoiceHandler struct {
	svc service.InvoiceService
}

func NewInvoiceHandler(svc service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// ListInvoices godoc
// @Summary      List invoices
// @Description  Returns paginated warehouse invoices with date and search filters.
// @Tags         Warehouse Invoices
// @Produce      json
// @Security     BearerAuth
// @Param        date query string false "Invoice date, YYYY-MM-DD"
// @Param        search query string false "Substring by invoice_number/request_number"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} InvoiceListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/invoices [get]
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	q, ok := parseInvoiceQuery(c)
	if !ok {
		return
	}
	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ExportInvoices godoc
// @Summary      Export invoices
// @Description  Exports filtered warehouse invoices as UTF-8 CSV with BOM.
// @Tags         Warehouse Invoices
// @Produce      text/csv
// @Security     BearerAuth
// @Param        date query string false "Invoice date, YYYY-MM-DD"
// @Param        search query string false "Substring by invoice_number/request_number"
// @Success      200 {file} file
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/invoices/export [get]
func (h *InvoiceHandler) ExportInvoices(c *gin.Context) {
	q, ok := parseInvoiceQuery(c)
	if !ok {
		return
	}
	data, err := h.svc.ExportCSV(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="invoices.csv"`)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// GetInvoiceByID godoc
// @Summary      Get invoice by ID
// @Description  Returns warehouse invoice by ID.
// @Tags         Warehouse Invoices
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Invoice ID"
// @Success      200 {object} InvoiceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoiceByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "invoice not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// CreateInvoice godoc
// @Summary      Create invoice
// @Description  Creates warehouse invoice history record.
// @Tags         Warehouse Invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.InvoiceCreateRequest true "Invoice payload"
// @Success      201 {object} InvoiceResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/invoices [post]
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req dto.InvoiceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		writeInvoiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": item})
}

func parseInvoiceQuery(c *gin.Context) (dto.InvoiceListQuery, bool) {
	q := dto.InvoiceListQuery{
		Date:   c.Query("date"),
		Search: c.Query("search"),
	}
	var ok bool
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return q, false
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return q, false
	}
	return q, true
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return 0, false
	}
	return id, true
}

func parseIntQuery(c *gin.Context, key string, allowZero bool) (int, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}
	return n, true
}

func writeInvoiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrInvoiceNumberExists):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "invoice_number already exists"})
	case errors.Is(err, repository.ErrInvoicePartRequestNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part request not found"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
