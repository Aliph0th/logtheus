package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/utils"
	userProto "logtheus/shared/pkg/pb/v1/user"

	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserController struct {
	userClient userProto.UserServiceClient
}

func NewUserController(userClient userProto.UserServiceClient) *UserController {
	return &UserController{
		userClient: userClient,
	}
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	data := utils.MustDTO[*dto.RegisterRequest](ctx)

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.userClient.RegisterUser(grpcCtx, &userProto.RegisterUserRequest{
		Email:    data.Email,
		Username: data.Username,
		Password: data.Password,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":         response.User,
		"accessToken":  response.AccessToken,
		"refreshToken": response.RefreshToken,
	})
}

func (c *UserController) VerifyEmail(ctx *gin.Context) {
	data := utils.MustDTO[*dto.VerifyEmailRequest](ctx)

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.userClient.VerifyUserEmail(grpcCtx, &userProto.VerifyUserEmailRequest{
		Code: data.Code,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"accessToken": response.AccessToken})
}

func (c *UserController) Login(ctx *gin.Context) {
	data := utils.MustDTO[*dto.LoginRequest](ctx)

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.userClient.LoginUser(grpcCtx, &userProto.LoginUserRequest{
		Email:    data.Email,
		Password: data.Password,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"user":         response.User,
		"accessToken":  response.AccessToken,
		"refreshToken": response.RefreshToken,
	})
}

func (c *UserController) GetCurrentUser(ctx *gin.Context) {
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.userClient.GetMe(grpcCtx, &emptypb.Empty{})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"user": response,
	})
}
