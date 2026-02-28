package api

import (
	"context"

	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/user/internal/services"
	"logtheus/user/internal/utils"

	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	userProto.UnimplementedUserServiceServer
	userService  *services.UserService
	tokenService *services.TokenService
}

func NewUserHandler(userService *services.UserService, tokenService *services.TokenService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		tokenService: tokenService,
	}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *userProto.RegisterUserRequest) (*userProto.SuccessAuthUserResponse, error) {
	user, accessToken, refreshToken, err := h.userService.CreateUser(req)
	if err != nil {
		return nil, err
	}
	return &userProto.SuccessAuthUserResponse{
		User:         utils.ToUserProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *userProto.LoginUserRequest) (*userProto.SuccessAuthUserResponse, error) {
	user, accessToken, refreshToken, err := h.userService.LoginUser(req)
	if err != nil {
		return nil, err
	}
	return &userProto.SuccessAuthUserResponse{
		User:         utils.ToUserProto(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) VerifyUserEmail(ctx context.Context, req *userProto.VerifyUserEmailRequest) (*userProto.VerifyUserEmailResponse, error) {
	//TODO: refresh token!
	accessToken, _, err := h.userService.VerifyUserEmail(ctx, req)
	if err != nil {
		return nil, err
	}
	return &userProto.VerifyUserEmailResponse{
		AccessToken: accessToken,
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *userProto.ValidateTokenRequest) (*userProto.ValidateTokenResponse, error) {
	claims, err := h.userService.ValidateToken(req.Token)
	if err != nil {
		return nil, err
	}
	return &userProto.ValidateTokenResponse{
		UserId:          claims.UserID,
		IsEmailVerified: claims.IsEmailVerified,
	}, nil
}

func (h *UserHandler) GetMe(ctx context.Context, req *emptypb.Empty) (*userProto.User, error) {
	user, err := h.userService.GetMe(ctx)
	if err != nil {
		return nil, err
	}
	return utils.ToUserProto(user), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userProto.GetUserRequest) (*userProto.User, error) {
	user, err := h.userService.GetUserByIdentifier(req)
	if err != nil {
		return nil, err
	}
	return utils.ToUserProto(user), nil
}

func (h *UserHandler) IssueInviteToken(ctx context.Context, req *emptypb.Empty) (*userProto.InviteTokenResponse, error) {
	token, err := h.tokenService.IssueInviteToken()
	if err != nil {
		return nil, err
	}
	return &userProto.InviteTokenResponse{Token: token}, nil
}
