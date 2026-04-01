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

		c.Set("user_id", claims.UserID)
		c.Set("role_id", claims.RoleID)
		c.Set("email", claims.Email)
		c.Set("iin", claims.IIN)
		c.Next()
	}
}

func RequireRoles(allowedRoles ...int64) gin.HandlerFunc {
	allowed := make(map[int64]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get("role_id")
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
