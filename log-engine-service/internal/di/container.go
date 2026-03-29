package di

import (
	"logtheus/logengine/internal/api"
	"logtheus/logengine/internal/config"
	"logtheus/logengine/internal/repository"
	"logtheus/logengine/internal/services"
	"logtheus/logengine/internal/storages"
	"logtheus/shared/pkg/interceptors"

	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })
	_ = c.Provide(func(cfg *config.AppConfig) (*storages.ClickHouse, error) {
		return storages.NewClickHouseStorage(cfg)
	})

	//Repositories
	_ = c.Provide(repository.NewLogRepository)

	//Services
	_ = c.Provide(services.NewLogEngineService)
	_ = c.Provide(services.NewS3Service)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewLogEngineHandler)

	return c
}
