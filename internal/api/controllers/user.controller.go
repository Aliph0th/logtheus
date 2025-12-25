package controllers

import (
	"logtheus/internal/api/dto"
	"logtheus/internal/service"
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

func (c *UserController) GetUserByID(ctx *gin.Context) {
	id := ctx.Param("id")

	user, err := c.userService.GetUserByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	username := ctx.GetString("username")
	email := ctx.GetString("email")
	password := ctx.GetString("password")

	dto := dto.RegisterRequest{Email: email, Password: password, Username: username}

	user, err := c.userService.CreateUser(ctx, &dto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}
