package service

import (
	"fmt"
	"log/slog"
	"logtheus/internal/api/dto"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/config"
	"logtheus/internal/consts"
	"logtheus/internal/models"
	"logtheus/internal/repository"
	"logtheus/internal/utils"
	sl "logtheus/internal/utils/logger"

	"github.com/gin-gonic/gin"
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

func (s *UserService) GetUserByID(ctx *gin.Context, id uint) (*models.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) CreateUser(ctx *gin.Context, req *dto.RegisterRequest) (*models.User, string, string, error) {

	candidate, _ := s.repo.GetByEmail(req.Email)
	if candidate != nil {
		return nil, "", "", excepts.WithConflict("User with email already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", fmt.Errorf("Failed to hash password: %w", err)
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(passwordHash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, "", "", excepts.Wrap(err, 500, "USER_CREATE_FAILED")
	}

	token, err := s.tokenService.IssueToken(user.ID, consts.TYPE_VERIFY_TOKEN, consts.VERIFY_TOKEN_TTL)
	if err != nil {
		return nil, "", "", excepts.Wrap(err, 500, "TOKEN_ISSUE_FAILED")
	}
	go func() {
		if err := s.mailService.SendVerifyEmail(user.Email, user.Username, s.cfg.AppDomain, token); err != nil {
			slog.Error("Verification email failed", "email", user.Email, sl.Error(err))
		}
	}()

	accessToken, refreshToken := s.tokenService.SignAuthTokens(&dto.UserAuthPayload{
		UserID:          user.ID,
		IsEmailVerified: false,
	})

	return user, accessToken, refreshToken, nil
}

func (s *UserService) VerifyUserEmail(ctx *gin.Context, code string) (string, string, error) {
	auth := utils.MustAuth(ctx)
	if auth.IsEmailVerified {
		return "", "", excepts.WithBadRequest("Email is already verified")
	}
	token, err := s.tokenService.ConsumeToken(auth.UserID, code, consts.TYPE_VERIFY_TOKEN)
	if err != nil {
		return "", "", err
	}
	if err := s.repo.VerifyEmail(token.UserID); err != nil {
		return "", "", excepts.Wrap(err, 500, "USER_VERIFY_EMAIL_FAILED")
	}

	accessToken, refreshToken := s.tokenService.SignAuthTokens(&dto.UserAuthPayload{
		UserID:          auth.UserID,
		IsEmailVerified: true,
	})

	return accessToken, refreshToken, nil
}

func (s *UserService) Login(ctx *gin.Context, req *dto.LoginRequest) (*models.User, string, string, error) {
	user, _ := s.repo.GetByEmail(req.Email)
	if user == nil {
		return nil, "", "", excepts.WithUnauthorized("Invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", "", excepts.WithUnauthorized("Invalid email or password")
	}
	accessToken, refreshToken := s.tokenService.SignAuthTokens(&dto.UserAuthPayload{
		UserID:          user.ID,
		IsEmailVerified: user.IsEmailVerified,
	})
	return user, accessToken, refreshToken, nil
}
