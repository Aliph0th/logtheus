package di

import (
	"logtheus/project/internal/api"
	"logtheus/project/internal/config"
	"logtheus/project/internal/repository"
	"logtheus/project/internal/services"
	"logtheus/shared/pkg/clients"
	"logtheus/shared/pkg/interceptors"
	mailProto "logtheus/shared/pkg/pb/v1/mail"
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/shared/pkg/storages"

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
	_ = c.Provide(func(cfg *config.AppConfig) userProto.UserServiceClient {
		return clients.NewUserClient(cfg.Services.User)
	})

	//Repositories
	_ = c.Provide(repository.NewProjectRepository)
	_ = c.Provide(repository.NewInvitesRepository)

	//Services
	_ = c.Provide(services.NewProjectService)
	_ = c.Provide(services.NewInvitesService)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewProjectHandler)

	return c
}
