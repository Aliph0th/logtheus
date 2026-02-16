package di

import (
	"logtheus/internal/api/controllers"
	"logtheus/internal/api/middleware"
	"logtheus/internal/config"
	"logtheus/internal/repository"
	"logtheus/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

func Build(cfg *config.AppConfig, db *gorm.DB) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })
	_ = c.Provide(func() *gorm.DB { return db })

	// Repositories
	_ = c.Provide(repository.NewUserRepository)
	_ = c.Provide(repository.NewTokenRepository)
	_ = c.Provide(repository.NewProjectRepository)
	_ = c.Provide(repository.NewInvitesRepository)

	// Services
	_ = c.Provide(service.NewTokenService)
	_ = c.Provide(service.NewMailService)
	_ = c.Provide(service.NewUserService)
	_ = c.Provide(service.NewProjectService)
	_ = c.Provide(service.NewInvitesService)

	// Controllers
	_ = c.Provide(controllers.NewUserController)
	_ = c.Provide(controllers.NewProjectController)
	_ = c.Provide(controllers.NewInvitesController)

	// Middlewares
	_ = c.Provide(func(tokenService *service.TokenService) gin.HandlerFunc {
		return middleware.AuthMiddleware(false, tokenService)
	})

	return c
}
