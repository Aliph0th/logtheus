package api

import (
	"context"

	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/user/internal/services"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userProto.UnimplementedUserServiceServer
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *userProto.RegisterUserRequest) (*userProto.SuccessAuthUserResponse, error) {
	user, accessToken, refreshToken, err := h.userService.CreateUser(req)
	if err != nil {
		return nil, err
	}
	return &userProto.SuccessAuthUserResponse{
		User: &userProto.User{
			Id:              user.ID,
			Email:           user.Email,
			Username:        user.Username,
			IsEmailVerified: user.IsEmailVerified,
			CreatedAt:       timestamppb.New(user.CreatedAt),
			UpdatedAt:       timestamppb.New(user.UpdatedAt),
		},
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
		User: &userProto.User{
			Id:              user.ID,
			Email:           user.Email,
			IsEmailVerified: user.IsEmailVerified,
			CreatedAt:       timestamppb.New(user.CreatedAt),
			UpdatedAt:       timestamppb.New(user.UpdatedAt),
		},
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
	return &userProto.User{
		Id:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       timestamppb.New(user.CreatedAt),
		UpdatedAt:       timestamppb.New(user.UpdatedAt),
	}, nil
}
