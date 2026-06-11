package handlers

import (
	"fmt"
	"net/http"

	"auto_park/internal/apierrors"
	"auto_park/internal/auditlog"
	"auto_park/internal/config"
	"auto_park/middleware"
	auditlogservice "auto_park/modules/audit_log_module/service"
	dto "auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	cfg      *config.Config
	svc      *service.UserService
	auditSvc *auditlogservice.Service
}

func NewUsersHandler(cfg *config.Config, svc *service.UserService, auditSvc *auditlogservice.Service) *UsersHandler {
	return &UsersHandler{cfg: cfg, svc: svc, auditSvc: auditSvc}
}

// POST /api/users (admin only)  <-- роль проверяется в router через middleware.RequireRoles(...)

// CreateUser godoc
// @Summary      Create user
// @Description  Creates a new user (admin only). JWT required.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body dto.CreateUserRequest true "Create user payload"
// @Success      201 {object} models.CreateUserResponseWrap
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/users [post]
func (h *UsersHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid json: " + err.Error()})
		return
	}

	res, err := h.svc.CreateUser(c.Request.Context(), service.CreateUserRequest{
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: req.MiddleName,
		IIN:        req.IIN,
		Phone:      req.Phone,
		RoleID:     req.RoleID,
	})
	if err != nil {
		switch err {
		case service.ErrEmailExists:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "email already exists"})
			return
		case service.ErrRoleNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": apierrors.CodeUnknownRole, "error": "role not found"})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	actor := auditlog.Actor(middleware.CurrentEmail(c), 0)
	auditlog.Write(
		c.Request.Context(),
		h.auditSvc,
		"success",
		"user",
		"",
		"created",
		actor,
		auditlog.Message("id", res.User.ID, "email", res.User.Email, "role_id", req.RoleID, "name", fmt.Sprintf("%s %s", req.LastName, req.FirstName)),
	)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": dto.CreateUserResponse{
			ID:        res.User.ID,
			Email:     res.User.Email,
			RoleID:    res.User.RoleID,
			CreatedAt: res.User.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	})
}
