package api

import (
	"logtheus/internal/api/controllers"
	"logtheus/internal/api/validators"
	"logtheus/internal/repository"
	"logtheus/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	router.Use(cors.Default())

	// TODO:
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.GET("/:id", userController.GetUserByID)
			users.POST("", append(validators.RegisterValidators, userController.CreateUser)...)
		}
	}
	return router
}
