package service

import (
	"context"

	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	mailProto "logtheus/shared/pkg/pb/v1/mail"
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/user/internal/config"
	userConsts "logtheus/user/internal/consts"
	"logtheus/user/internal/models"
	"logtheus/user/internal/repository"
	"logtheus/user/internal/types"
	"logtheus/user/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/durationpb"
)

type UserService struct {
	repo         *repository.UserRepository
	tokenService *TokenService
	mailClient   mailProto.MailServiceClient
	cfg          *config.AppConfig
}

func NewUserService(
	repo *repository.UserRepository,
	tokenService *TokenService,
	cfg *config.AppConfig,
	mailClient mailProto.MailServiceClient,
) *UserService {
	return &UserService{
		repo:         repo,
		tokenService: tokenService,
		mailClient:   mailClient,
		cfg:          cfg,
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
		return nil, "", "", grpc.WithAlreadyExists("User with email already exists").WithSlug(consts.ERROR_CODE_USER_EXISTS)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", grpc.WithInternal().WithSlug(consts.INTERNAL_ERROR_CODE_PASSWORD_HASH_FAILED)
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(passwordHash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, "", "", grpc.WithInternal().WithSlug(consts.INTERNAL_ERROR_CODE_USER_CREATE_FAILED)
	}

	token, err := s.tokenService.IssueEmailVerificationToken(user.ID)
	if err != nil {
		return nil, "", "", grpc.WithInternal().WithSlug(consts.INTERNAL_ERROR_CODE_VERIFICATION_TOKEN_ISSUE_FAILED)
	}

	_, err = s.sendVerifyEmailAsync(user.Email, user.Username, token)
	if err != nil {
		return nil, "", "", grpc.WithInternal().WithSlug(consts.INTERNAL_ERROR_CODE_SEND_EMAIL_FAILED)
	}

	accessToken, refreshToken := s.tokenService.SignAuthTokens(&types.UserAuthPayload{
		UserID:          user.ID,
		IsEmailVerified: false,
	})

	return user, accessToken, refreshToken, nil
}

func (s *UserService) LoginUser(req *userProto.LoginUserRequest) (*models.User, string, string, error) {
	user, _ := s.repo.GetByEmail(req.Email)
	if user == nil {
		return nil, "", "", grpc.WithUnauthenticated("Invalid email or password").WithSlug(consts.ERROR_CODE_INVALID_CREDENTIALS)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", "", grpc.WithUnauthenticated("Invalid email or password").WithSlug(consts.ERROR_CODE_INVALID_CREDENTIALS)
	}
	accessToken, refreshToken := s.tokenService.SignAuthTokens(&types.UserAuthPayload{
		UserID:          user.ID,
		IsEmailVerified: user.IsEmailVerified,
	})
	return user, accessToken, refreshToken, nil
}

func (s *UserService) VerifyUserEmail(ctx context.Context, req *userProto.VerifyUserEmailRequest) (string, string, error) {
	auth := utils.MustUserData(ctx)
	if auth.IsEmailVerified {
		return "", "", grpc.WithInvalidArgument("Email is already verified").WithSlug(consts.ERROR_CODE_EMAIL_ALREADY_VERIFIED)
	}
	err := s.tokenService.UseEmailVerificationToken(auth.UserID, req.Code)
	if err != nil {
		return "", "", grpc.WithInternal().WithSlug(consts.INTERNAL_ERROR_CODE_USER_VERIFY_EMAIL_FAILED)
	}

	if err := s.repo.VerifyEmail(auth.UserID); err != nil {
		return "", "", grpc.WithInternal().WithSlug(consts.INTERNAL_ERROR_CODE_USER_VERIFY_EMAIL_FAILED)
	}

	accessToken, refreshToken := s.tokenService.SignAuthTokens(&types.UserAuthPayload{
		UserID:          auth.UserID,
		IsEmailVerified: true,
	})

	return accessToken, refreshToken, nil
}

func (s *UserService) sendVerifyEmailAsync(email, username, code string) (*mailProto.SuccessfulResponse, error) {
	ctx := context.Background()

	req := &mailProto.SendVerifyEmailRequest{
		Email:      email,
		Username:   username,
		Code:       code,
		Expiration: durationpb.New(userConsts.TTL_VERIFY_TOKEN),
	}

	response, err := s.mailClient.SendVerifyEmail(ctx, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
