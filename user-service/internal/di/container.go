package di

import (
	"logtheus/user/internal/api"
	"logtheus/user/internal/api/interceptors"
	"logtheus/user/internal/config"
	"logtheus/user/internal/repository"
	service "logtheus/user/internal/services"
	"logtheus/user/internal/storages"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

func Build(cfg *config.AppConfig, db *gorm.DB) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })
	_ = c.Provide(func() *gorm.DB { return db })
	_ = c.Provide(storages.NewRedisClient)

	// Repositories
	_ = c.Provide(repository.NewTokenRepository)
	_ = c.Provide(repository.NewUserRepository)

	// Services
	_ = c.Provide(service.NewTokenService)
	_ = c.Provide(service.NewUserService)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewUserHandler)

	return c
}
