package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"auto_park/middleware"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type UsersDeleteHandler struct {
	svc *service.UsersDeleteService
}

func NewUsersDeleteHandler(svc *service.UsersDeleteService) *UsersDeleteHandler {
	return &UsersDeleteHandler{svc: svc}
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user deleted successfully",
		"id":      targetID,
	})
}
