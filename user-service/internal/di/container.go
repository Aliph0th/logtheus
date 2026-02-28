package di

import (
	"logtheus/shared/pkg/clients"
	"logtheus/shared/pkg/interceptors"
	mailProto "logtheus/shared/pkg/pb/v1/mail"
	"logtheus/shared/pkg/storages"
	"logtheus/user/internal/api"
	"logtheus/user/internal/config"
	"logtheus/user/internal/repository"
	"logtheus/user/internal/services"

	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })
	_ = c.Provide(func() *storages.Database {
		db, _ := storages.NewPostgres(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
		return db
	})
	_ = c.Provide(func() *storages.RedisDatabase {
		db, _ := storages.NewRedisClient(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.Database)
		return db
	})

	// gRPC Clients
	_ = c.Provide(func(cfg *config.AppConfig) mailProto.MailServiceClient {
		return clients.NewMailClient(cfg.Services.Mail)
	})

	// Repositories
	_ = c.Provide(repository.NewTokenRepository)
	_ = c.Provide(repository.NewUserRepository)

	// Services
	_ = c.Provide(services.NewTokenService)
	_ = c.Provide(services.NewUserService)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewUserHandler)

	return c
}
