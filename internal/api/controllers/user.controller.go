package controllers

import (
	"logtheus/internal/api/dto"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/service"
	utils "logtheus/internal/utils"

	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	data := utils.MustDTO[dto.RegisterRequest](ctx)

	user, accessToken, refreshToken, err := c.userService.CreateUser(ctx, &data)
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (c *UserController) VerifyEmail(ctx *gin.Context) {
	data := utils.MustDTO[dto.VerifyEmailRequest](ctx)

	accessToken, refreshToken, err := c.userService.VerifyUserEmail(ctx, data.Code)
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true, "accessToken": accessToken, "refreshToken": refreshToken})
}
