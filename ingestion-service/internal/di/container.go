package di

import (
	"logtheus/ingestion/internal/api"
	"logtheus/ingestion/internal/config"
	"logtheus/ingestion/internal/services"
	"logtheus/shared/pkg/clients"
	"logtheus/shared/pkg/interceptors"
	applicationProto "logtheus/shared/pkg/pb/v1/application"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	projectProto "logtheus/shared/pkg/pb/v1/project"

	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })

	// gRPC Clients
	_ = c.Provide(func(cfg *config.AppConfig) applicationProto.ApplicationServiceClient {
		return clients.NewApplicationClient(cfg.Services.Application)
	})
	_ = c.Provide(func(cfg *config.AppConfig) projectProto.ProjectServiceClient {
		return clients.NewProjectClient(cfg.Services.Project)
	})
	_ = c.Provide(func(cfg *config.AppConfig) logEngineProto.LogEngineServiceClient {
		return clients.NewLogEngineClient(cfg.Services.LogEngine)
	})

	//Services
	_ = c.Provide(services.NewIngestionService)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewIngestionHandler)

	return c
}
