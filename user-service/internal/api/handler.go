package api

import (
	"context"

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
	return &userProto.SuccessAuthUserResponse{
		User: &userProto.User{
			Id:              1,
			Email:           req.Email,
			Username:        req.Username,
			IsEmailVerified: false,
			CreatedAt:       timestamppb.Now(),
			UpdatedAt:       timestamppb.Now(),
		},
		AccessToken:  "access_token_placeholder",
		RefreshToken: "refresh_token_placeholder",
	}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *userProto.LoginUserRequest) (*userProto.SuccessAuthUserResponse, error) {
	return &userProto.SuccessAuthUserResponse{
		User: &userProto.User{
			Id:              1,
			Email:           req.Email,
			IsEmailVerified: true,
			CreatedAt:       timestamppb.Now(),
			UpdatedAt:       timestamppb.Now(),
		},
		AccessToken:  "access_token_placeholder",
		RefreshToken: "refresh_token_placeholder",
	}, nil
}

func (h *UserHandler) VerifyUserEmail(ctx context.Context, req *userProto.VerifyUserEmailRequest) (*userProto.VerifyUserEmailResponse, error) {
	return &userProto.VerifyUserEmailResponse{
		AccessToken: "access_token_placeholder",
	}, nil
}
