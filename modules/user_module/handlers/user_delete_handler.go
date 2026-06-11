package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/internal/auditlog"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	"auto_park/modules/user_module/models"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type UsersDeleteHandler struct {
	svc      *service.UsersDeleteService
	auditSvc *auditlogservice.Service
}

func NewUsersDeleteHandler(svc *service.UsersDeleteService, auditSvc *auditlogservice.Service) *UsersDeleteHandler {
	return &UsersDeleteHandler{svc: svc, auditSvc: auditSvc}
}

// DELETE /api/users/:id (admin only) <-- роль проверяется в router через middleware.RequireRoles(...)

// DeleteUser godoc
// @Summary      Delete user
// @Description  Deletes user by id (admin only). JWT required.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} models.DeleteUserResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/{id} [delete]
func (h *UsersDeleteHandler) DeleteUser(c *gin.Context) {
	requesterID, err := middleware.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	target, _ := h.svc.GetUserBeforeDelete(c.Request.Context(), targetID)

	err = h.svc.DeleteUserAdmin(c.Request.Context(), requesterID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotDeleteYourself):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "cannot delete yourself"})
			return
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "user not found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"warning",
		"user",
		"active",
		"deleted",
		auditlog.Actor(middleware.CurrentEmail(c), requesterID),
		auditlog.Message(
			"id", targetID,
			"email", userAuditEmail(target),
			"role_id", userAuditRoleID(target),
			"name", userAuditFullName(target),
		),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user deleted successfully",
		"id":      targetID,
	})
}

func userAuditEmail(u *models.UserPublic) string {
	if u == nil {
		return ""
	}
	return u.Email
}

func userAuditRoleID(u *models.UserPublic) int64 {
	if u == nil {
		return 0
	}
	return u.RoleID
}

func userAuditFullName(u *models.UserPublic) string {
	if u == nil {
		return ""
	}
	return strings.TrimSpace(u.LastName + " " + u.FirstName + " " + auditStringValue(u.MiddleName))
}

func auditStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
