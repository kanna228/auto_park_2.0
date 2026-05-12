package middleware

import (
	"errors"
	"net/http"
	"strings"

	"auto_park/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	RoleID int64  `json:"role_id"`
	IIN    string `json:"iin"`
	jwt.RegisteredClaims
}

func AuthJWT(cfg *config.Config) gin.HandlerFunc {
	secret := []byte(cfg.Auth.JWTSecret)

	return func(c *gin.Context) {
		tokenStr := ""

		if cookie, err := c.Cookie("session_token"); err == nil && cookie != "" {
			tokenStr = cookie
		}

		if tokenStr == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				tokenStr = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}

		// Useful for browser WebSocket clients because native WebSocket API
		// cannot set Authorization header. Prefer cookie/Bearer for normal HTTP.
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
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextIINKey, claims.IIN)
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

		roleID, ok := roleVal.(int64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "invalid role type"})
			return
		}

		if _, ok := allowed[roleID]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}
