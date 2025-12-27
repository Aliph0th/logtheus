package service

import (
	"context"
	"fmt"
	"log/slog"
	"logtheus/internal/api/dto"
	"logtheus/internal/config"
	"logtheus/internal/consts"
	"logtheus/internal/models"
	"logtheus/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo         *repository.UserRepository
	tokenService *TokenService
	mailService  *MailService
	cfg          *config.AppConfig
}

func NewUserService(
	repo *repository.UserRepository,
	tokenService *TokenService,
	mailService *MailService,
	cfg *config.AppConfig,
) *UserService {
	return &UserService{
		repo:         repo,
		tokenService: tokenService,
		mailService:  mailService,
		cfg:          cfg,
	}
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

	token, err := s.tokenService.IssueToken(user.ID, consts.TYPE_VERIFY_TOKEN, consts.VERIFY_TOKEN_TTL)
	if err != nil {
		return nil, fmt.Errorf("Failed to issue verification token: %w", err)
	}
	go func(email, username, domain, code string) {
		if err := s.mailService.SendVerifyEmail(email, username, domain, code); err != nil {
			slog.Error("Verification email failed", "email", email, "err", err)
		}
	}(user.Email, user.Username, s.cfg.AppDomain, token)

	return user, nil
}
