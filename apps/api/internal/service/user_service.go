package service

import (
	"context"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type UserService interface {
	FindUserByEmail(
		ctx context.Context,
		email string,
	) (*domain.User, error)
}

type UserServiceImpl struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &UserServiceImpl{
		users: users,
	}
}

func (s *UserServiceImpl) FindUserByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {
	return s.users.FindByEmail(ctx, email)
}
