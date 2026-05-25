package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"auto_park/internal/config"
	"auto_park/modules/user_module/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepo interface {
	GetAuthByEmail(ctx context.Context, email string) (*models.UserAuth, error)
	UpdateSession(ctx context.Context, accountType string, accountID int64, token string, when time.Time) error
}

type AuthService struct {
	repo      AuthRepo
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(cfg *config.Config, repo AuthRepo) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(cfg.Auth.JWTSecret),
		tokenTTL:  cfg.Auth.TokenTTL,
	}
}

func (s *AuthService) TokenTTL() time.Duration { return s.tokenTTL }

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

var (
	ErrInvalidEmail    = errors.New("invalid credentials")
	ErrInvalidPassword = errors.New("invalid credentials")
)

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	ua, err := s.repo.GetAuthByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidEmail
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ua.PassHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	now := time.Now()
	expires := now.Add(s.tokenTTL)
	accountType := ua.AccountType
	if accountType == "" {
		accountType = "user"
	}
	var driverID int64
	if ua.DriverID != nil {
		driverID = *ua.DriverID
	}

	claims := jwtClaims{
		UserID:      ua.ID,
		AccountType: accountType,
		DriverID:    driverID,
		Email:       ua.Email,
		RoleID:      ua.RoleID,
		RoleName:    ua.RoleName,
		IIN:         ua.IIN,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
			Subject:   accountType + ":" + strconv.FormatInt(ua.ID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateSession(ctx, accountType, ua.ID, jwtString, now); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:       jwtString,
		UserID:      ua.ID,
		AccountType: accountType,
		DriverID:    ua.DriverID,
		Email:       ua.Email,
		RoleID:      ua.RoleID,
		RoleName:    ua.RoleName,
		LastSeen:    now.UTC(),
		ExpiresAt:   expires.Unix(),
	}, nil
}
