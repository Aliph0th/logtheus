package api

import (
	"context"
	"log/slog"

	userProto "logtheus/shared/pkg/pb/v1/user"
	service "logtheus/user/internal/services"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userProto.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *userProto.RegisterUserRequest) (*userProto.SuccessAuthUserResponse, error) {
	slog.Info("RegisterUser called", "email", req.Email)
	user, accessToken, refreshToken, err := h.userService.CreateUser(req)
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
	return &userProto.VerifyUserEmailResponse{
		AccessToken: "access_token_placeholder",
	}, nil
}
