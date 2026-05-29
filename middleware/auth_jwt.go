package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auto_park/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	UserID      int64  `json:"uid"`
	AccountType string `json:"account_type"`
	DriverID    int64  `json:"driver_id,omitempty"`
	Email       string `json:"email"`
	RoleID      int64  `json:"role_id"`
	RoleName    string `json:"role_name"`
	IIN         string `json:"iin"`
	jwt.RegisteredClaims
}

func AuthJWT(cfg *config.Config) gin.HandlerFunc {
	secret := []byte(cfg.Auth.JWTSecret)

	return func(c *gin.Context) {
		tokenStr := ""

		// ВАЖНО:
		// Authorization Bearer должен быть приоритетнее cookie.
		// Иначе в тестах и в браузере cookie может перебивать нужный token.
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}

		if tokenStr == "" {
			if cookie, err := c.Cookie("session_token"); err == nil && cookie != "" {
				tokenStr = cookie
			}
		}

		// Для WebSocket, где нельзя поставить Authorization header.
		if tokenStr == "" {
			tokenStr = strings.TrimSpace(c.Query("access_token"))
		}
		if tokenStr == "" {
			tokenStr = strings.TrimSpace(c.Query("token"))
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "missing auth token"})
			return
		}

		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid or expired token"})
			return
		}

		if claims.UserID <= 0 || claims.RoleID <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid token claims"})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleIDKey, claims.RoleID)
		c.Set(ContextRoleNameKey, claims.RoleName)
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextIINKey, claims.IIN)
		if strings.TrimSpace(claims.AccountType) == "" {
			claims.AccountType = "user"
		}
		c.Set(ContextAccountTypeKey, claims.AccountType)
		c.Set(ContextDriverIDKey, claims.DriverID)
		if claims.ExpiresAt != nil {
			c.Set(ContextExpiresAtKey, claims.ExpiresAt.Unix())
		}

		c.Next()
	}
}

func RequireRoles(allowedRoles ...int64) gin.HandlerFunc {
	allowed := make(map[int64]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextRoleIDKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "role not found in context"})
			return
		}

		roleID, ok := roleIDFromAny(roleVal)
		if !ok || roleID <= 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "invalid role type"})
			return
		}

		// Администратор всегда имеет полный доступ.
		if roleID == 1 {
			c.Next()
			return
		}

		if _, ok := allowed[roleID]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func roleIDFromAny(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case uint:
		if uint64(v) > uint64(^uint64(0)>>1) {
			return 0, false
		}
		return int64(v), true
	case uint64:
		if v > uint64(^uint64(0)>>1) {
			return 0, false
		}
		return int64(v), true
	case float64:
		roleID := int64(v)
		if v != float64(roleID) {
			return 0, false
		}
		return roleID, true
	case string:
		roleID, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return roleID, err == nil
	default:
		return 0, false
	}
}
