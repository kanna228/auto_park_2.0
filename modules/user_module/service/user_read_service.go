package service

import (
	"context"
	"errors"

	"auto_park/modules/user_module/models"
	"auto_park/modules/user_module/repository"
)

var (
	ErrAccessDenied = errors.New("access denied")
)

type UsersReadService struct {
	repo repository.UsersReadRepo
}

func NewUsersReadService(repo repository.UsersReadRepo) *UsersReadService {
	return &UsersReadService{repo: repo}
}

func (s *UsersReadService) ListUsers(ctx context.Context, requesterRoleID int64, requesterUserID int64, limit int, offset int) ([]models.UserPublic, int64, error) {
	return s.repo.ListUsersForRole(ctx, requesterRoleID, requesterUserID, limit, offset)
}

// role=1 → может смотреть любого
// role=2 → может смотреть только пользователей с ролью 2 или 3
// role=3 → может смотреть только самого себя
func (s *UsersReadService) GetUserByID(ctx context.Context, requesterRoleID int64, requesterUserID int64, targetID int64) (*models.UserPublic, error) {
	target, err := s.repo.GetUserPublicByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	allowed := false
	switch requesterRoleID {
	case 1:
		allowed = true
	case 2:
		allowed = (target.RoleID == 2 || target.RoleID == 3)
	case 3:
		allowed = (target.ID == requesterUserID)
	}

	if !allowed {
		return nil, ErrAccessDenied
	}

	return target, nil
}
