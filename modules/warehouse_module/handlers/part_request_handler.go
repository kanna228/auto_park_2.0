package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	invoiceservice "auto_park/modules/invoice_module/service"
	"auto_park/modules/warehouse_module/dto"
	"auto_park/modules/warehouse_module/repository"
	"auto_park/modules/warehouse_module/service"

	"github.com/gin-gonic/gin"
)

type PartRequestHandler struct {
	svc        service.PartRequestService
	auditSvc   *auditlogservice.Service
	invoiceSvc invoiceservice.InvoiceService
}

func NewPartRequestHandler(svc service.PartRequestService, auditSvc *auditlogservice.Service, invoiceSvc invoiceservice.InvoiceService) *PartRequestHandler {
	return &PartRequestHandler{svc: svc, auditSvc: auditSvc, invoiceSvc: invoiceSvc}
}

// CreatePartRequest godoc
// @Summary      Create part order request
// @Description  Creates a new online request for one warehouse part. Only duty mechanic can create it. One request contains only one part type. The first history record is created automatically.
// @Tags         Warehouse Part Requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.PartRequestCreateRequest true "Part request create payload"
// @Success      201 {object} PartRequestCreateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests [post]
func (h *PartRequestHandler) CreatePartRequest(c *gin.Context) {
	authorUserID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	var req dto.PartRequestCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	id, err := h.svc.Create(c.Request.Context(), authorUserID, req)
	if err != nil {
		writePartRequestError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": id}})
}

// GetPartRequestByID godoc
// @Summary      Get part request by ID
// @Description  Returns a warehouse part order request by ID with full request history.
// @Tags         Warehouse Part Requests
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part request ID"
// @Success      200 {object} PartRequestResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests/{id} [get]
func (h *PartRequestHandler) GetPartRequestByID(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": item})
}

// ListPartRequests godoc
// @Summary      List part requests
// @Description  Returns warehouse part requests with access control: admin and warehouse manager see all requests, duty mechanic sees only own requests. Supports status/date filters and created_at sorting.
// @Tags         Warehouse Part Requests
// @Produce      json
// @Security     BearerAuth
// @Param        part_id query int false "Filter by warehouse part internal ID"
// @Param        status_id query int false "Filter by request status ID: 1=new, 2=rejected, 3=approved"
// @Param        status_code query string false "Filter by request status code: new, rejected, approved"
// @Param        author_user_id query int false "Filter by mechanic user ID. Works only for admin and warehouse manager"
// @Param        date_from query string false "Created date from, YYYY-MM-DD"
// @Param        date_to query string false "Created date to, YYYY-MM-DD"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by field: created_at, updated_at, part_id, quantity, status_id, status_code, author_user_id" default(created_at)
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} PartRequestListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests [get]
func (h *PartRequestHandler) ListPartRequests(c *gin.Context) {
	currentUserID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	roleID, ok := getRoleIDFromContext(c)
	if !ok {
		return
	}

	q := dto.PartRequestListQuery{
		StatusCode: c.Query("status_code"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		SortBy:     c.Query("sort_by"),
		Order:      c.Query("order"),
	}

	if q.SortBy == "" {
		q.SortBy = "created_at"
	}

	var okQuery bool

	if q.PartID, okQuery = parseInt64Query(c, "part_id", false); !okQuery {
		return
	}

	if q.StatusID, okQuery = parseInt64Query(c, "status_id", false); !okQuery {
		return
	}

	if q.AuthorUserID, okQuery = parseInt64Query(c, "author_user_id", false); !okQuery {
		return
	}

	if q.Limit, okQuery = parseIntQuery(c, "limit", false); !okQuery {
		return
	}

	if q.Offset, okQuery = parseIntQuery(c, "offset", true); !okQuery {
		return
	}

	resp, err := h.svc.List(c.Request.Context(), currentUserID, roleID, q)
	if err != nil {
		writePartRequestError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ListAllPartRequestHistory godoc
// @Summary      List all part request history
// @Description  Returns full immutable history journal for warehouse part requests with filters by request, status, changer and dates.
// @Tags         Warehouse Part Requests
// @Produce      json
// @Security     BearerAuth
// @Param        part_request_id query int false "Filter by part request ID"
// @Param        status_id query int false "Filter by status ID"
// @Param        status_code query string false "Filter by status code: new, rejected, approved"
// @Param        changed_by_user_id query int false "Filter by user who changed request"
// @Param        date_from query string false "Changed date from, YYYY-MM-DD"
// @Param        date_to query string false "Changed date to, YYYY-MM-DD"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        sort_by query string false "Sort by field: changed_at, part_request_id, status_id, status_code, changed_by_user_id"
// @Param        order query string false "Sort order: asc or desc"
// @Success      200 {object} PartRequestHistoryListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-request-history [get]
func (h *PartRequestHandler) ListAllPartRequestHistory(c *gin.Context) {
	q := dto.PartRequestHistoryListQuery{
		StatusCode: c.Query("status_code"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
		SortBy:     c.Query("sort_by"),
		Order:      c.Query("order"),
	}

	var ok bool
	if q.PartRequestID, ok = parseInt64Query(c, "part_request_id", false); !ok {
		return
	}
	if q.StatusID, ok = parseInt64Query(c, "status_id", false); !ok {
		return
	}
	if q.ChangedByUserID, ok = parseInt64Query(c, "changed_by_user_id", false); !ok {
		return
	}
	if q.Limit, ok = parseIntQuery(c, "limit", false); !ok {
		return
	}
	if q.Offset, ok = parseIntQuery(c, "offset", true); !ok {
		return
	}

	resp, err := h.svc.ListHistory(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// ListPartRequestHistory godoc
// @Summary      List part request history
// @Description  Returns immutable history records for a part request with pagination.
// @Tags         Warehouse Part Requests
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part request ID"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} PartRequestHistoryListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests/{id}/history [get]
func (h *PartRequestHandler) ListPartRequestHistory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	limit, ok := parseIntQuery(c, "limit", false)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}

	resp, err := h.svc.ListHistoryByRequestID(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// UpdatePartRequest godoc
// @Summary      Full update part request
// @Description  Fully updates part request fields including status and creates a history record. If status is rejected, rejection_comment is required and only warehouse manager can perform it. This route is blocked after request status becomes approved or rejected.
// @Tags         Warehouse Part Requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part request ID"
// @Param        payload body dto.PartRequestUpdateRequest true "Part request update payload"
// @Success      200 {object} PartRequestUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests/{id} [put]
func (h *PartRequestHandler) UpdatePartRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	changedByUserID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	roleID, ok := getRoleIDFromContext(c)
	if !ok {
		return
	}

	var req dto.PartRequestUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	updated, err := h.svc.UpdateByID(c.Request.Context(), id, changedByUserID, roleID, req)
	if err != nil {
		writePartRequestError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// UpdatePartRequestStatus godoc
// @Summary      Change part request status
// @Description  Changes only the request status and creates a history record. If status is rejected, rejection_comment is required and only warehouse manager can perform it. This route can change status even if the request was already approved or rejected.
// @Tags         Warehouse Part Requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part request ID"
// @Param        payload body dto.PartRequestStatusUpdateRequest true "Part request status update payload"
// @Success      200 {object} PartRequestStatusUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests/{id}/status [patch]
func (h *PartRequestHandler) UpdatePartRequestStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	changedByUserID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	roleID, ok := getRoleIDFromContext(c)
	if !ok {
		return
	}

	var req dto.PartRequestStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	current, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if current == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	updated, err := h.svc.UpdateStatusByID(c.Request.Context(), id, changedByUserID, roleID, req)
	if err != nil {
		writePartRequestError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	updatedItem, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if updatedItem != nil {
		actor := actorFromContext(c, changedByUserID)
		if h.auditSvc != nil {
			if err := h.auditSvc.Write(c.Request.Context(), "info", "request", statusDisplay(current.Status), statusDisplay(updatedItem.Status), actor, ""); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
				return
			}
		}
		if h.invoiceSvc != nil && isIssuedPartRequestStatus(updatedItem.Status) {
			if _, err := h.invoiceSvc.CreateForPartRequest(c.Request.Context(), id, strconv.FormatInt(id, 10)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// UpdatePartRequestRepairStatus godoc
// @Summary      Complete repair for approved part request
// @Description  Marks the repair as completed and creates vehicle_part_installations on backend without a second warehouse stock write-off. Stock is already issued when the part request is approved.
// @Tags         Warehouse Part Requests
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part request ID"
// @Param        payload body dto.PartRequestRepairStatusUpdateRequest true "Repair status update payload"
// @Success      200 {object} PartRequestStatusUpdateResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests/{id}/repair-status [patch]
func (h *PartRequestHandler) UpdatePartRequestRepairStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	changedByUserID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	roleID, ok := getRoleIDFromContext(c)
	if !ok {
		return
	}

	var req dto.PartRequestRepairStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	updated, err := h.svc.UpdateRepairStatusByID(c.Request.Context(), id, changedByUserID, roleID, req)
	if err != nil {
		writePartRequestError(c, err)
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// DeletePartRequest godoc
// @Summary      Delete part request
// @Description  Soft-deletes a part request and keeps immutable history. Deletion is blocked after request status becomes approved or rejected.
// @Tags         Warehouse Part Requests
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Part request ID"
// @Success      200 {object} PartRequestDeleteResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      409 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-requests/{id} [delete]
func (h *PartRequestHandler) DeletePartRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	changedByUserID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	deleted, err := h.svc.DeleteByID(c.Request.Context(), id, changedByUserID)
	if err != nil {
		writePartRequestError(c, err)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "part request not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// ListPartRequestStatuses godoc
// @Summary      List part request statuses
// @Description  Returns available statuses for warehouse part requests with pagination.
// @Tags         Warehouse Part Requests
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} PartRequestStatusListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/warehouse/part-request-statuses [get]
func (h *PartRequestHandler) ListPartRequestStatuses(c *gin.Context) {
	limit, ok := parseIntQuery(c, "limit", false)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}

	resp, err := h.svc.ListStatuses(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func getUserIDFromContext(c *gin.Context) (int64, bool) {
	return middleware.CurrentUserIDOrAbort(c)
}

func actorFromContext(c *gin.Context, userID int64) string {
	actor := strings.TrimSpace(middleware.CurrentEmail(c))
	if actor == "" {
		actor = strconv.FormatInt(userID, 10)
	}
	return actor
}

func statusDisplay(status dto.PartRequestStatusResponse) string {
	if value := strings.TrimSpace(status.Name); value != "" {
		return value
	}
	return strings.TrimSpace(status.Code)
}

func isIssuedPartRequestStatus(status dto.PartRequestStatusResponse) bool {
	code := strings.ToLower(strings.TrimSpace(status.Code))
	name := strings.ToLower(strings.TrimSpace(status.Name))
	return code == "issued" || strings.Contains(name, "\u0432\u044b\u0434\u0430\u043d")
}

func parseInt64Query(c *gin.Context, key string, allowZero bool) (int64, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, true
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}
	return n, true
}

func parseIntQuery(c *gin.Context, key string, allowZero bool) (int, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return 0, true
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 || (!allowZero && n == 0) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid " + key})
		return 0, false
	}
	return n, true
}

func writePartRequestError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrPartRequestPartNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part not found"})
	case errors.Is(err, repository.ErrPartRequestStatusNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "part request status not found"})
	case errors.Is(err, repository.ErrPartRequestUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "user not found"})
	case errors.Is(err, repository.ErrPartRequestLocked):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "part request cannot be changed after approval or rejection"})
	case errors.Is(err, repository.ErrPartRequestInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "not enough part quantity in stock"})
	case errors.Is(err, repository.ErrPartRequestRepairContextRequired):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "vehicle_id, mechanic_shift_id and planned_replacement_at are required to complete repair"})
	case errors.Is(err, repository.ErrPartRequestRepairCompletionForbidden):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "part request must be approved before repair can be completed"})
	case errors.Is(err, repository.ErrVehiclePartInstallationVehicleNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "vehicle not found"})
	case errors.Is(err, repository.ErrVehiclePartInstallationMechanicShiftNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "mechanic shift not found"})
	case errors.Is(err, repository.ErrVehiclePartInstallationUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "installer user not found"})
	case errors.Is(err, repository.ErrVehiclePartInstallationActiveDuplicate):
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": "non-consumable part is already active on this vehicle"})
	case errors.Is(err, service.ErrPartRequestRejectionCommentRequired):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "rejection_comment is required when status is rejected"})
	case errors.Is(err, service.ErrPartRequestRejectForbidden):
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "only warehouse manager can reject part request"})
	case errors.Is(err, service.ErrPartRequestStatusChangeForbidden):
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "only warehouse manager can approve or reject part request"})
	case errors.Is(err, service.ErrPartRequestViewForbidden):
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "part request list is not available for this role"})
	case errors.Is(err, service.ErrPartRequestRepairStatusUnsupported):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "only completed repair status is supported"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	}
}
