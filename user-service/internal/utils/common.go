package utils

import (
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/user/internal/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToUserProto(user *models.User) *userProto.User {
	return &userProto.User{
		Id:              user.ID,
		Email:           user.Email,
		Username:        user.Username,
		IsEmailVerified: user.IsEmailVerified,
		CreatedAt:       timestamppb.New(user.CreatedAt),
		UpdatedAt:       timestamppb.New(user.UpdatedAt),
	}
}
