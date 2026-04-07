package routes

import (
	"logtheus/gateway/internal/api/controllers"
	"logtheus/gateway/internal/api/dto"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/api/validators"
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/shared/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func RegisterUserRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.UserController](container)
	userClient := utils.MustResolve[userProto.UserServiceClient](container)
	authMiddleware := middleware.AuthMiddleware(true, userClient)
	users := api.Group("/users")
	{
		users.POST("/login",
			append(
				validators.LoginValidators,
				middleware.BindDTO[*dto.LoginRequest](),
				controller.Login,
			)...,
		)
		users.POST("/register",
			append(
				validators.RegisterValidators,
				middleware.BindDTO[*dto.RegisterRequest](),
				controller.CreateUser,
			)...,
		)
		users.POST("/verify",
			append(
				[]gin.HandlerFunc{authMiddleware},
				append(
					validators.VerifyEmailValidators,
					middleware.BindDTO[*dto.VerifyEmailRequest](),
					controller.VerifyEmail,
				)...,
			)...,
		)
		users.GET("/me", authMiddleware, controller.GetCurrentUser)
	}
}
