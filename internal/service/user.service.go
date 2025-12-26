package service

import (
	"context"
	"fmt"
	"logtheus/internal/api/dto"
	"logtheus/internal/models"
	"logtheus/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo}
}

func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.RegisterRequest) (*models.User, error) {

	candidate, _ := s.repo.GetByEmail(req.Email)

	if candidate != nil {
		return nil, fmt.Errorf("User with email %s already exists", req.Email)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Failed to hash password: %w", err)
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(passwordHash),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}
