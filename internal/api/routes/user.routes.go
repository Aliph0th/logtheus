package routes

import (
	"logtheus/internal/api/controllers"
	"logtheus/internal/api/dto"
	"logtheus/internal/api/middleware"
	"logtheus/internal/api/validators"
	"logtheus/internal/repository"
	"logtheus/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterUserRoutes(api *gin.RouterGroup, db *gorm.DB) {
	repository := repository.NewUserRepository(db)
	service := service.NewUserService(repository)
	controller := controllers.NewUserController(service)

	users := api.Group("/users")
	{
		users.POST("/register",
			append(
				validators.RegisterValidators,
				middleware.BindDTO[dto.RegisterRequest](),
				controller.CreateUser,
			)...,
		)
	}
}
