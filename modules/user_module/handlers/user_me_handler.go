package handlers

import (
	"net/http"

	"auto_park/middleware"

	"github.com/gin-gonic/gin"
)

type CurrentUserResponse struct {
	ID          int64   `json:"id"`
	Email       string  `json:"email"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	MiddleName  *string `json:"middle_name"`
	RoleID      int64   `json:"role_id"`
	RoleName    string  `json:"role_name"`
	AccountType string  `json:"account_type"`
	DriverID    *int64  `json:"driver_id"`
	ExpiresAt   int64   `json:"expires_at"`
}

// Me godoc
// @Summary      Current user
// @Description  Returns current authenticated account from JWT claims.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} CurrentUserResponse
// @Failure      401 {object} ErrorResponse
// @Router       /api/users/me [get]
func (h *UsersReadHandler) Me(c *gin.Context) {
	auth, ok := middleware.CurrentAuthOrAbort(c)
	if !ok {
		return
	}

	var driverID *int64
	if auth.AccountType == "driver" && auth.DriverID > 0 {
		v := auth.DriverID
		driverID = &v
	}

	if auth.AccountType == "driver" {
		driver, err := h.svc.GetDriverByID(c.Request.Context(), auth.DriverID)
		if err == nil && driver != nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": CurrentUserResponse{
				ID:          auth.UserID,
				Email:       driver.Email,
				FirstName:   driver.FirstName,
				LastName:    driver.LastName,
				MiddleName:  driver.MiddleName,
				RoleID:      auth.RoleID,
				RoleName:    auth.RoleName,
				AccountType: auth.AccountType,
				DriverID:    driverID,
				ExpiresAt:   auth.ExpiresAt,
			}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": CurrentUserResponse{
			ID:          auth.UserID,
			Email:       auth.Email,
			RoleID:      auth.RoleID,
			RoleName:    auth.RoleName,
			AccountType: auth.AccountType,
			DriverID:    driverID,
			ExpiresAt:   auth.ExpiresAt,
		}})
		return
	}

	u, err := h.svc.GetUserByID(c.Request.Context(), auth.RoleID, auth.UserID, auth.UserID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": CurrentUserResponse{
			ID:          auth.UserID,
			Email:       auth.Email,
			RoleID:      auth.RoleID,
			RoleName:    auth.RoleName,
			AccountType: auth.AccountType,
			DriverID:    driverID,
			ExpiresAt:   auth.ExpiresAt,
		}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": CurrentUserResponse{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		MiddleName:  u.MiddleName,
		RoleID:      u.RoleID,
		RoleName:    auth.RoleName,
		AccountType: auth.AccountType,
		DriverID:    driverID,
		ExpiresAt:   auth.ExpiresAt,
	}})
}
