package handlers

import (
	"net/http"
	"strings"
	"unicode"

	"auto_park/internal/config"
	dto "auto_park/modules/user_module/dto"
	"auto_park/modules/user_module/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg *config.Config
	svc *service.AuthService
}

func NewAuthHandler(cfg *config.Config, svc *service.AuthService) *AuthHandler {
	return &AuthHandler{cfg: cfg, svc: svc}
}

// Login godoc
// @Summary      Login
// @Description  Authenticate user and set cookie `session_token`. Returns JWT token in response. No auth required.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        payload body dto.LoginRequest true "Login payload"
// @Success      200 {object} models.AuthLoginResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Router       /api/users/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid json"})
		return
	}

	email := sanitize(req.Email)
	iin := sanitize(req.IIN)
	password := sanitizePassword(req.Password)

	resp, err := h.svc.Login(c.Request.Context(), email, password, iin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid credentials"})
		return
	}

	// cookie (из cfg)
	maxAge := int(h.cfg.Auth.TokenTTL.Seconds())
	c.SetSameSite(parseSameSite(h.cfg.Auth.CookieSameSite))
	c.SetCookie(
		"session_token",
		resp.Token,
		maxAge,
		"/",
		h.cfg.Auth.CookieDomain,
		h.cfg.Auth.CookieSecure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimFunc(s, func(r rune) bool { return unicode.IsSpace(r) })
}
func sanitizePassword(p string) string {
	p = strings.TrimSpace(p)
	return strings.TrimFunc(p, func(r rune) bool { return unicode.IsSpace(r) })
}
func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
