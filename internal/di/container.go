package di

import (
	"logtheus/internal/api/controllers"
	"logtheus/internal/config"
	"logtheus/internal/repository"
	"logtheus/internal/service"

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

	// Services
	_ = c.Provide(service.NewTokenService)
	_ = c.Provide(service.NewMailService)
	_ = c.Provide(service.NewUserService)

	// Controllers
	_ = c.Provide(controllers.NewUserController)

	return c
}
