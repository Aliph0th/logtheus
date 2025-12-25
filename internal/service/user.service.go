package service

import (
	"context"
	"logtheus/internal/api/dto"
	"logtheus/internal/models"
	"logtheus/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.RegisterRequest) (*models.User, error) {
	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
