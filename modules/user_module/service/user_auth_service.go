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
	UpdateSession(ctx context.Context, userID int64, token string, when time.Time) error
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
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	RoleID int64  `json:"role_id"`
	IIN    string `json:"iin"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidEmail    = errors.New("invalid credentials")
	ErrInvalidPassword = errors.New("invalid credentials")
	ErrInvalidIIN      = errors.New("invalid credentials")
)

func (s *AuthService) Login(ctx context.Context, email, password, iin string) (*models.AuthResponse, error) {
	ua, err := s.repo.GetAuthByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidEmail
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ua.PassHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}
	if ua.IIN != iin {
		return nil, ErrInvalidIIN
	}

	now := time.Now()
	expires := now.Add(s.tokenTTL)

	claims := jwtClaims{
		UserID: ua.ID,
		Email:  ua.Email,
		RoleID: ua.RoleID,
		IIN:    ua.IIN,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
			Subject:   strconv.FormatInt(ua.ID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateSession(ctx, ua.ID, jwtString, now); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:     jwtString,
		UserID:    ua.ID,
		Email:     ua.Email,
		RoleID:    ua.RoleID,
		LastSeen:  now.UTC(),
		ExpiresAt: expires.Unix(),
	}, nil
}
