package service

import (
	"context"
	"errors"
	"strings"

	"auto_park/modules/user_module/models"
	"auto_park/modules/user_module/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UsersUpdateRepo interface {
	EnsureRoleExists(ctx context.Context, roleID int64) error
	EmailBelongsToOther(ctx context.Context, email string, userID int64) (bool, error)
	UpdateUserAdmin(ctx context.Context, p repository.UpdateUserAdminParams) (*models.UserPublic, error)
}

type UpdateUserRequest struct {
	Email      *string
	FirstName  *string
	LastName   *string
	MiddleName *string
	Password   *string
	Phone      *string
	RoleID     *int64
	IIN        *string
}

type UsersUpdateService struct {
	repo UsersUpdateRepo
}

func NewUsersUpdateService(repo UsersUpdateRepo) *UsersUpdateService {
	return &UsersUpdateService{repo: repo}
}

func (s *UsersUpdateService) UpdateUserAdmin(ctx context.Context, targetID int64, req UpdateUserRequest) (*models.UserPublic, error) {
	var (
		emailPtr, firstPtr, lastPtr, middlePtr, phonePtr, iinPtr *string
		rolePtr                                                  *int64
		passHashPtr                                              *string
	)

	if req.Email != nil {
		e := strings.TrimSpace(*req.Email)
		if e == "" {
			return nil, errors.New("email cannot be empty")
		}
		if err := validateEmail(e); err != nil {
			return nil, err
		}
		inUse, err := s.repo.EmailBelongsToOther(ctx, e, targetID)
		if err != nil {
			return nil, err
		}
		if inUse {
			return nil, ErrEmailExists
		}
		emailPtr = &e
	}

	if req.FirstName != nil {
		v := strings.TrimSpace(*req.FirstName)
		if v == "" {
			return nil, errors.New("first_name cannot be empty")
		}
		firstPtr = &v
	}
	if req.LastName != nil {
		v := strings.TrimSpace(*req.LastName)
		if v == "" {
			return nil, errors.New("last_name cannot be empty")
		}
		lastPtr = &v
	}
	if req.MiddleName != nil {
		v := strings.TrimSpace(*req.MiddleName)
		middlePtr = &v // пустая строка останется пустой (как у тебя было)
	}

	if req.Phone != nil {
		v := strings.TrimSpace(*req.Phone)
		if v != "" {
			if err := validatePhone(v); err != nil {
				return nil, err
			}
		}
		phonePtr = &v
	}

	if req.IIN != nil {
		v := strings.TrimSpace(*req.IIN)
		if err := validateIIN(v); err != nil {
			return nil, err
		}
		iinPtr = &v
	}

	if req.RoleID != nil {
		if err := s.repo.EnsureRoleExists(ctx, *req.RoleID); err != nil {
			if errors.Is(err, repository.ErrRoleNotFound) {
				return nil, ErrRoleNotFound
			}
			return nil, err
		}
		rolePtr = req.RoleID
	}

	if req.Password != nil {
		p := strings.TrimSpace(*req.Password)
		if len(p) < 4 {
			return nil, errors.New("password too short")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("cannot hash password")
		}
		h := string(hash)
		passHashPtr = &h
	}

	updated, err := s.repo.UpdateUserAdmin(ctx, repository.UpdateUserAdminParams{
		ID:         targetID,
		Email:      emailPtr,
		FirstName:  firstPtr,
		LastName:   lastPtr,
		MiddleName: middlePtr,
		IIN:        iinPtr,
		Phone:      phonePtr,
		RoleID:     rolePtr,
		PassHash:   passHashPtr,
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			return nil, ErrUserNotFound
		case errors.Is(err, repository.ErrEmailExists):
			return nil, ErrEmailExists
		case errors.Is(err, repository.ErrRoleNotFound):
			return nil, ErrRoleNotFound
		default:
			return nil, err
		}
	}

	return updated, nil
}
