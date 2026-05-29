package handlers

import (
	"errors"
	"net/http"

	"auto_park/internal/apierrors"
	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/repository"
	"auto_park/modules/warehouse_module/service"

	"github.com/gin-gonic/gin"
)

type PurchaseRequestHandler struct {
	svc service.PurchaseRequestService
}

func NewPurchaseRequestHandler(svc service.PurchaseRequestService) *PurchaseRequestHandler {
	return &PurchaseRequestHandler{svc: svc}
}

func (h *PurchaseRequestHandler) CreatePurchaseRequest(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	var req dto.PurchaseRequestCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	id, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		writePurchaseRequestError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

func (h *PurchaseRequestHandler) GetPurchaseRequestByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		writePurchaseRequestError(c, err)
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "purchase request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

func (h *PurchaseRequestHandler) ListPurchaseRequests(c *gin.Context) {
	q := dto.PurchaseRequestListQuery{
		Status: c.Query("status"),
		SortBy: c.Query("sort_by"),
		Order:  c.Query("order"),
	}

	var ok bool
	if q.PartID, ok = parseInt64Query(c, "part_id", false); !ok {
		return
	}
	if q.SourcePartRequestID, ok = parseInt64Query(c, "source_part_request_id", false); !ok {
		return
	}
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return
	}

	resp, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		writePurchaseRequestError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *PurchaseRequestHandler) ConfirmPurchaseRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	confirmed, err := h.svc.Confirm(c.Request.Context(), id, userID)
	if err != nil {
		writePurchaseRequestError(c, err)
		return
	}
	if !confirmed {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "purchase request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func (h *PurchaseRequestHandler) CancelPurchaseRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	var req struct {
		Comment *string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	cancelled, err := h.svc.Cancel(c.Request.Context(), id, userID, req.Comment)
	if err != nil {
		writePurchaseRequestError(c, err)
		return
	}
	if !cancelled {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "purchase request not found or already closed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

func writePurchaseRequestError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrPartRequestPartNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part not found"})
	case errors.Is(err, apierrors.ErrEntityArchived):
		c.JSON(http.StatusConflict, gin.H{"success": false, "code": apierrors.CodeEntityArchived, "error": "Нельзя изменить архивный объект"})
	case errors.Is(err, repository.ErrPartRequestUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "user not found"})
	case errors.Is(err, repository.ErrPurchaseRequestLocked):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "purchase request cannot be changed after confirmation or cancellation"})
	case errors.Is(err, repository.ErrPurchaseRequestAlreadyConfirmed):
		c.JSON(http.StatusConflict, gin.H{"success": false, "code": apierrors.CodePurchaseAlreadyConfirmed, "error": "purchase request already confirmed"})
	case errors.Is(err, repository.ErrPurchaseRequestSourcePartRequestNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "source part request not found"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
