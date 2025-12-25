package controllers

import (
	"logtheus/internal/api/dto"
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

	user, err := c.userService.CreateUser(ctx, &data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}
