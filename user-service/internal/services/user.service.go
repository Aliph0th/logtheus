package service

import (
	"fmt"
	"logtheus/shared/pkg/grpc"
	"logtheus/user/internal/config"
	"logtheus/user/internal/models"
	"logtheus/user/internal/repository"

	userProto "logtheus/shared/pkg/pb/v1/user"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
	cfg  *config.AppConfig
}

func NewUserService(
	repo *repository.UserRepository,
	cfg *config.AppConfig,
) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *UserService) GetUserByID(id uint64) (*models.User, error) {
	return s.repo.GetByID(id)
}
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *UserService) CreateUser(req *userProto.RegisterUserRequest) (*models.User, string, string, error) {
	candidate, _ := s.repo.GetByEmail(req.Email)
	if candidate != nil {
		return nil, "", "", grpc.WithAlreadyExists("User with email already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		//TODO: error handling
		return nil, "", "", fmt.Errorf("Failed to hash password: %w", err)
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(passwordHash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, "", "", grpc.WithInternal("USER_CREATE_FAILED")
	}

	// token, err := s.tokenService.IssueToken(user.ID, enums.TOKEN_TYPE_VERIFY, consts.VERIFY_TOKEN_TTL)
	// if err != nil {
	// 	return nil, "", "", excepts.Wrap(err, 500, "TOKEN_ISSUE_FAILED")
	// }
	// go func() {
	// 	if err := s.mailService.SendVerifyEmail(user.Email, user.Username, s.cfg.AppDomain, token); err != nil {
	// 		slog.Error("Verification email failed", "email", user.Email, sl.Error(err))
	// 	}
	// }()

	// accessToken, refreshToken := s.tokenService.SignAuthTokens(&dto.UserAuthPayload{
	// 	UserID:          user.ID,
	// 	IsEmailVerified: false,
	// })

	return user, "", "", nil
}

func (s *UserService) VerifyUserEmail(req *userProto.VerifyUserEmailRequest) (string, string, error) {
	// auth := utils.MustAuth(ctx)
	// if auth.IsEmailVerified {
	// 	return "", "", excepts.WithBadRequest("Email is already verified")
	// }
	// token, err := s.tokenService.ConsumeToken(auth.UserID, code, enums.TOKEN_TYPE_VERIFY)
	// if err != nil {
	// 	return "", "", err
	// }

	// if err := s.repo.VerifyEmail(token.UserID); err != nil {
	// 	return "", "", grpc.WithInternal("USER_VERIFY_EMAIL_FAILED")
	// }

	// accessToken, refreshToken := s.tokenService.SignAuthTokens(&dto.UserAuthPayload{
	// 	UserID:          auth.UserID,
	// 	IsEmailVerified: true,
	// })

	return "", "", nil
}

func (s *UserService) Login(req *userProto.LoginUserRequest) (*models.User, string, string, error) {
	user, _ := s.repo.GetByEmail(req.Email)
	if user == nil {
		return nil, "", "", grpc.WithUnauthenticated("Invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", "", grpc.WithUnauthenticated("Invalid email or password")
	}
	// accessToken, refreshToken := s.tokenService.SignAuthTokens(&dto.UserAuthPayload{
	// 	UserID:          user.ID,
	// 	IsEmailVerified: user.IsEmailVerified,
	// })
	return user, "", "", nil
}
