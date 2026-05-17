package services

import (
	"context"
	"errors"

	"logtheus/shared/pkg/clients"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/shared/pkg/types"
	"logtheus/shared/pkg/utils"
	"logtheus/user/internal/config"
	userConsts "logtheus/user/internal/consts"
	"logtheus/user/internal/models"
	"logtheus/user/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo         *repository.UserRepository
	tokenService *TokenService
	mailProducer *clients.MailEventProducer
	cfg          *config.AppConfig
}

func NewUserService(
	repo *repository.UserRepository,
	tokenService *TokenService,
	cfg *config.AppConfig,
	mailProducer *clients.MailEventProducer,
) *UserService {
	return &UserService{
		repo:         repo,
		tokenService: tokenService,
		mailProducer: mailProducer,
		cfg:          cfg,
	}
}

func (s *UserService) CreateUser(req *userProto.RegisterUserRequest) (*models.User, string, string, error) {
	candidate, _ := s.repo.GetByEmail(req.Email)
	if candidate != nil {
		return nil, "", "", grpc.WithAlreadyExists("User with email already exists").WithSlug(consts.ERROR_CODE_USER_EXISTS)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_PASSWORD_HASH_FAILED)
	}

	user := &models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(passwordHash),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, "", "", grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	token, err := s.tokenService.IssueEmailVerificationToken(user.ID)
	if err != nil {
		return nil, "", "", grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_VERIFICATION_TOKEN_ISSUE_FAILED)
	}

	err = s.sendVerifyEmail(user.Email, user.Username, token)
	if err != nil {
		return nil, "", "", grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_SEND_EMAIL_FAILED)
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
		var grpcErr *grpc.GRPCError
		if errors.As(err, &grpcErr) {
			return "", "", grpcErr
		}
		return "", "", grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	if err := s.repo.VerifyEmail(auth.UserID); err != nil {
		return "", "", grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	accessToken, refreshToken := s.tokenService.SignAuthTokens(&types.UserAuthPayload{
		UserID:          auth.UserID,
		IsEmailVerified: true,
	})

	return accessToken, refreshToken, nil
}

func (s *UserService) ValidateToken(token string) (*types.UserAuthClaims, error) {
	payload, err := s.tokenService.VerifyAccessToken(token)
	if err != nil {
		return nil, grpc.WithUnauthenticated("Invalid token").WithSlug(consts.ERROR_CODE_UNAUTHENTICATED)
	}
	return payload, nil
}

func (s *UserService) GetMe(ctx context.Context) (*models.User, error) {
	auth := utils.MustUserData(ctx)
	if auth == nil {
		return nil, grpc.WithInvalidArgument("No authenticated user data").WithSlug(consts.ERROR_CODE_UNAUTHENTICATED)
	}
	user, err := s.repo.GetByID(auth.UserID)
	if err != nil {
		//TODO: maybe distinguish not found and other errors?
		return nil, grpc.WithNotFound("User not found")
	}
	return user, nil
}

func (s *UserService) sendVerifyEmail(email, username, code string) error {
	ctx := context.Background()

	event := &types.VerifyEmailEvent{
		Email:             email,
		Username:          username,
		Code:              code,
		ExpirationMinutes: uint8(userConsts.TTL_VERIFY_TOKEN.Minutes()),
	}

	return s.mailProducer.PublishVerifyEmail(ctx, event)
}

func (s *UserService) GetUserByIdentifier(req *userProto.GetUserRequest) (*models.User, error) {
	switch req.GetIdentifier().(type) {
	case *userProto.GetUserRequest_UserId:
		return s.repo.GetByID(req.GetUserId())
	case *userProto.GetUserRequest_Email:
		return s.repo.GetByEmail(req.GetEmail())
	}
	return nil, grpc.WithInvalidArgument("Invalid identifier type").WithSlug(consts.ERROR_CODE_UNKNOWN)
}

func (s *UserService) GetUsersByIDs(ids []uint64) ([]*models.User, error) {
	return s.repo.GetByIDs(ids)
}
