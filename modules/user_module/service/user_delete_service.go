package service

import (
	"context"
	"errors"
)

var (
	ErrCannotDeleteYourself = errors.New("cannot delete yourself")
)

type UsersDeleteRepo interface {
	DeleteUserByID(ctx context.Context, id int64) error
}

type UsersDeleteService struct {
	repo UsersDeleteRepo
}

func NewUsersDeleteService(repo UsersDeleteRepo) *UsersDeleteService {
	return &UsersDeleteService{repo: repo}
}

// requesterID нужен только чтобы запретить self-delete
func (s *UsersDeleteService) DeleteUserAdmin(ctx context.Context, requesterID int64, targetID int64) error {
	if requesterID == targetID {
		return ErrCannotDeleteYourself
	}

	err := s.repo.DeleteUserByID(ctx, targetID)
	if err != nil {
		// пробрасываем как есть, handler сам решит 404/500
		return err
	}
	return nil
}
