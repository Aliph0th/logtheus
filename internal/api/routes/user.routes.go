package routes

import (
	"logtheus/internal/api/controllers"
	"logtheus/internal/api/dto"
	"logtheus/internal/api/middleware"
	"logtheus/internal/api/validators"
	"logtheus/internal/service"
	"logtheus/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func RegisterUserRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.UserController](container)
	tokenService := utils.MustResolve[*service.TokenService](container)

	users := api.Group("/users")
	{
		users.POST("/login",
			append(
				validators.LoginValidators,
				middleware.BindDTO[dto.LoginRequest](),
				controller.Login,
			)...,
		)
		users.POST("/register",
			append(
				validators.RegisterValidators,
				middleware.BindDTO[dto.RegisterRequest](),
				controller.CreateUser,
			)...,
		)
		users.POST("/verify",
			append(
				[]gin.HandlerFunc{middleware.AuthMiddleware(true, tokenService)},
				append(
					validators.VerifyEmailValidators,
					middleware.BindDTO[dto.VerifyEmailRequest](),
					controller.VerifyEmail,
				)...,
			)...,
		)
	}
}
