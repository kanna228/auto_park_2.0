package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"auto_park/middleware"
	"auto_park/modules/notification_module/dto"
	notificationservice "auto_park/modules/notification_module/service"
	ws "auto_park/modules/notification_module/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type NotificationHandler struct {
	svc      notificationservice.NotificationService
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func NewNotificationHandler(svc notificationservice.NotificationService, hub *ws.Hub) *NotificationHandler {
	return &NotificationHandler{
		svc: svc,
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// WebSocket godoc
// @Summary      Notifications websocket
// @Description  Opens websocket connection for realtime notifications. On connect server pushes unread snapshot and unread count. Browser can authenticate by session_token cookie; non-browser clients can use Authorization Bearer header.
// @Tags         Notifications
// @Security     BearerAuth
// @Router       /api/notifications/ws [get]
func (h *NotificationHandler) WebSocket(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, userID, conn)
	h.hub.Register(client)

	go client.WritePump()
	go func() {
		_ = h.svc.PushUnreadSnapshot(c.Request.Context(), userID)
	}()
	client.ReadPump()
}

// List godoc
// @Summary      List my notifications
// @Description  Returns current user's notifications. Use only_unread=true to get unread only.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        only_unread query bool false "Only unread notifications"
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} NotificationListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	onlyUnread := parseBoolQuery(c.Query("only_unread"))
	limit, ok := parseIntQuery(c, "limit", true)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}

	resp, err := h.svc.List(c.Request.Context(), userID, onlyUnread, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "unread_count": resp.UnreadCount, "data": resp})
}

// ListUnread godoc
// @Summary      List my unread notifications
// @Description  Returns current user's unread notifications.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} NotificationListResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/notifications/unread [get]
func (h *NotificationHandler) ListUnread(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	limit, ok := parseIntQuery(c, "limit", true)
	if !ok {
		return
	}
	offset, ok := parseIntQuery(c, "offset", true)
	if !ok {
		return
	}

	resp, err := h.svc.List(c.Request.Context(), userID, true, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// CountUnread godoc
// @Summary      Count my unread notifications
// @Description  Returns unread notifications count for current user.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} NotificationUnreadCountResponseWrap
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/notifications/unread/count [get]
func (h *NotificationHandler) CountUnread(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	count, err := h.svc.CountUnread(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": dto.NotificationUnreadCountResponse{Count: count}})
}

func (h *NotificationHandler) CountUnreadPlain(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	count, err := h.svc.CountUnread(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// MarkAsRead godoc
// @Summary      Mark notification as read
// @Description  Marks one current user's notification as read.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Notification ID"
// @Success      200 {object} NotificationResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	id, ok := parsePathID(c)
	if !ok {
		return
	}

	updated, err := h.svc.MarkAsRead(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !updated {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "notification not found or already read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}

// MarkAllAsRead godoc
// @Summary      Mark all notifications as read
// @Description  Marks all current user's unread notifications as read.
// @Tags         Notifications
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} NotificationMarkAllReadResponseWrap
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/notifications/read-all [patch]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := middleware.CurrentUserIDOrAbort(c)
	if !ok {
		return
	}

	updated, err := h.svc.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": dto.NotificationMarkAllReadResponse{Updated: updated}})
}

func parseBoolQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func parsePathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return 0, false
	}
	return id, true
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
