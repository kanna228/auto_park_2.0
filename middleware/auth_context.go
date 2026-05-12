package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "user_id"
	ContextRoleIDKey = "role_id"
	ContextEmailKey  = "email"
	ContextIINKey    = "iin"
)

type AuthContext struct {
	UserID int64
	RoleID int64
	Email  string
	IIN    string
}

func CurrentUserID(c *gin.Context) (int64, error) {
	return int64FromContext(c, ContextUserIDKey)
}

func CurrentRoleID(c *gin.Context) (int64, error) {
	return int64FromContext(c, ContextRoleIDKey)
}

func CurrentEmail(c *gin.Context) string {
	value, _ := c.Get(ContextEmailKey)
	email, _ := value.(string)
	return email
}

func CurrentIIN(c *gin.Context) string {
	value, _ := c.Get(ContextIINKey)
	iin, _ := value.(string)
	return iin
}

func CurrentAuth(c *gin.Context) (AuthContext, error) {
	userID, err := CurrentUserID(c)
	if err != nil {
		return AuthContext{}, err
	}

	roleID, err := CurrentRoleID(c)
	if err != nil {
		return AuthContext{}, err
	}

	return AuthContext{
		UserID: userID,
		RoleID: roleID,
		Email:  CurrentEmail(c),
		IIN:    CurrentIIN(c),
	}, nil
}

func CurrentUserIDOrAbort(c *gin.Context) (int64, bool) {
	userID, err := CurrentUserID(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return 0, false
	}
	return userID, true
}

func CurrentRoleIDOrAbort(c *gin.Context) (int64, bool) {
	roleID, err := CurrentRoleID(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "role not found in context"})
		return 0, false
	}
	return roleID, true
}

func CurrentAuthOrAbort(c *gin.Context) (AuthContext, bool) {
	auth, err := CurrentAuth(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return AuthContext{}, false
	}
	return auth, true
}

func int64FromContext(c *gin.Context, key string) (int64, error) {
	value, exists := c.Get(key)
	if !exists || value == nil {
		return 0, fmt.Errorf("%s not found in context", key)
	}

	switch v := value.(type) {
	case int64:
		if v <= 0 {
			return 0, fmt.Errorf("%s is invalid", key)
		}
		return v, nil
	case int:
		if v <= 0 {
			return 0, fmt.Errorf("%s is invalid", key)
		}
		return int64(v), nil
	case float64:
		if v <= 0 || v != float64(int64(v)) {
			return 0, fmt.Errorf("%s is invalid", key)
		}
		return int64(v), nil
	default:
		return 0, fmt.Errorf("%s has invalid type", key)
	}
}
