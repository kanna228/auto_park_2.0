package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"auto_park/internal/apierrors"
	"auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type UsersUpdateHandler struct {
	svc *service.UsersUpdateService
}

func NewUsersUpdateHandler(svc *service.UsersUpdateService) *UsersUpdateHandler {
	return &UsersUpdateHandler{svc: svc}
}

// PUT /api/users/:id (admin only) <-- роль проверяется в router через middleware.RequireRoles(...)

// UpdateUser godoc
// @Summary      Update user
// @Description  Updates user by id (admin only). JWT required.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        payload body dto.UpdateUserRequest true "Fields to update"
// @Success      200 {object} models.UserResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/{id} [put]
func (h *UsersUpdateHandler) UpdateUser(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid json: " + err.Error()})
		return
	}

	updated, err := h.svc.UpdateUserAdmin(c.Request.Context(), targetID, service.UpdateUserRequest{
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: req.MiddleName,
		Password:   req.Password,
		Phone:      req.Phone,
		RoleID:     req.RoleID,
		IIN:        req.IIN,
	})
	if err != nil {
		switch {
		case errors.Is(err, apierrors.ErrEntityArchived):
			c.JSON(http.StatusConflict, gin.H{"success": false, "code": apierrors.CodeEntityArchived, "error": "Нельзя изменить архивный объект"})
			return
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "user not found"})
			return
		case errors.Is(err, service.ErrEmailExists):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "email already exists"})
			return
		case errors.Is(err, service.ErrRoleNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": apierrors.CodeUnknownRole, "error": "role not found"})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated})
}
