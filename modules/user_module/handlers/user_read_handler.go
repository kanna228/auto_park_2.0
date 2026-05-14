package handlers

import (
	"net/http"
	"strconv"

	"auto_park/middleware"
	"auto_park/modules/user_module/repository"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type UsersReadHandler struct {
	svc *service.UsersReadService
}

func NewUsersReadHandler(svc *service.UsersReadService) *UsersReadHandler {
	return &UsersReadHandler{svc: svc}
}

// GET /api/users

// ListUsers godoc
// @Summary      List users
// @Description  Returns list of users. JWT required. Access rules are applied in service layer (role-based filtering).
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Limit" default(50)
// @Param        offset query int false "Offset" default(0)
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Page size" default(50)
// @Success      200 {object} models.UsersListResponse
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users [get]
func (h *UsersReadHandler) ListUsers(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	users, total, err := h.svc.ListUsers(c.Request.Context(), roleID, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": users, "total": total, "limit": limit, "offset": offset})
}

// GET /api/users/:id

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Returns user by id. JWT required. Access rules are applied in service layer (admin or self).
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} models.UserResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users/{id} [get]
func (h *UsersReadHandler) GetUserByID(c *gin.Context) {
	roleID, userID, ok := getAuthFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid id"})
		return
	}

	u, err := h.svc.GetUserByID(c.Request.Context(), roleID, userID, targetID)
	if err != nil {
		switch err {
		case repository.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "user not found"})
			return
		case service.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "access denied"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": u})
}

// helper: достаём role_id и user_id, которые положил AuthJWT middleware
func getAuthFromContext(c *gin.Context) (roleID int64, userID int64, ok bool) {
	auth, err := middleware.CurrentAuth(c)
	if err != nil {
		return 0, 0, false
	}

	return auth.RoleID, auth.UserID, true
}
