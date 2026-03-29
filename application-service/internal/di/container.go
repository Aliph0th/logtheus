package di

import (
	"logtheus/application/internal/api"
	"logtheus/application/internal/config"
	"logtheus/application/internal/repository"
	"logtheus/application/internal/services"
	"logtheus/shared/pkg/clients"
	"logtheus/shared/pkg/interceptors"
	projectProto "logtheus/shared/pkg/pb/v1/project"
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

	// gRPC Clients
	_ = c.Provide(func(cfg *config.AppConfig) projectProto.ProjectServiceClient {
		return clients.NewProjectClient(cfg.Services.Project)
	})
	_ = c.Provide(func(cfg *config.AppConfig) userProto.UserServiceClient {
		return clients.NewUserClient(cfg.Services.User)
	})

	//Repositories
	_ = c.Provide(repository.NewApplicationRepository)

	//Services
	_ = c.Provide(services.NewApplicationService)
	_ = c.Provide(services.NewAPIKeyService)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewApplicationHandler)

	return c
}
