package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"auto_park/internal/config"
	"auto_park/user_module/models"
	"auto_park/user_module/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrRoleNotFound = errors.New("role not found")
)

type UserRepo interface {
	EnsureRoleExists(ctx context.Context, roleID int64) error
	GetRoleNameByID(ctx context.Context, id int64) (string, error)
	EmailInUse(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, p repository.CreateUserParams) (*models.User, error)
}

type CreateUserRequest struct {
	Email      string
	FirstName  string
	LastName   string
	MiddleName *string
	IIN        string
	Phone      *string
	RoleID     int64
}

type CreateUserResult struct {
	User     *models.User
	RawPass  string
	RoleName string
}

type UserService struct {
	cfg    *config.Config
	repo   UserRepo
	mailer Mailer
}

func NewUserService(cfg *config.Config, repo UserRepo, mailer Mailer) *UserService {
	return &UserService{cfg: cfg, repo: repo, mailer: mailer}
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResult, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := validateIIN(req.IIN); err != nil {
		return nil, err
	}
	if req.Phone != nil {
		if err := validatePhone(*req.Phone); err != nil {
			return nil, err
		}
	}

	if err := s.repo.EnsureRoleExists(ctx, req.RoleID); err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	roleName, err := s.repo.GetRoleNameByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	exists, err := s.repo.EmailInUse(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	rawPass, err := generatePassword(14)
	if err != nil {
		return nil, errors.New("cannot generate password")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(rawPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("cannot hash password")
	}

	u, err := s.repo.CreateUser(ctx, repository.CreateUserParams{
		Email:      strings.TrimSpace(req.Email),
		FirstName:  strings.TrimSpace(req.FirstName),
		LastName:   strings.TrimSpace(req.LastName),
		MiddleName: normalizePtr(req.MiddleName),
		IIN:        strings.TrimSpace(req.IIN),
		Phone:      normalizePtr(req.Phone),
		RoleID:     req.RoleID,
		PassHash:   string(hash),
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	// best-effort email
	if s.mailer != nil {
		to := u.Email
		pass := rawPass
		rn := roleName
		go func() { _ = s.mailer.SendWelcome(context.Background(), to, pass, rn) }()
	}

	return &CreateUserResult{User: u, RawPass: rawPass, RoleName: roleName}, nil
}

func normalizePtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func validateEmail(e string) error {
	_, err := mail.ParseAddress(strings.TrimSpace(e))
	if err != nil {
		return errors.New("invalid email format")
	}
	return nil
}

func validateIIN(iin string) error {
	iin = strings.TrimSpace(iin)
	if len(iin) != 12 {
		return errors.New("iin must be 12 digits")
	}
	for _, r := range iin {
		if r < '0' || r > '9' {
			return errors.New("iin must contain only digits")
		}
	}
	return nil
}

func validatePhone(p string) error {
	re := regexp.MustCompile(`^\+?\d{10,15}$`)
	if !re.MatchString(strings.TrimSpace(p)) {
		return errors.New("invalid phone format")
	}
	return nil
}

func generatePassword(n int) (string, error) {
	if n < 10 {
		n = 10
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	s := base64.RawURLEncoding.EncodeToString(buf)
	if len(s) > n {
		s = s[:n]
	}
	return s, nil
}
