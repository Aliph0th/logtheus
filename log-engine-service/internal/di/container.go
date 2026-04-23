package di

import (
	"logtheus/logengine/internal/api"
	"logtheus/logengine/internal/config"
	"logtheus/logengine/internal/repository"
	"logtheus/logengine/internal/services"
	"logtheus/logengine/internal/storages"
	"logtheus/shared/pkg/interceptors"
	sharedStorages "logtheus/shared/pkg/storages"

	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })
	_ = c.Provide(func(cfg *config.AppConfig) (*storages.ClickHouse, error) {
		return storages.NewClickHouseStorage(cfg)
	})
	_ = c.Provide(func(cfg *config.AppConfig) (*sharedStorages.Database, error) {
		return sharedStorages.NewPostgres(cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Name)
	})

	//Repositories
	_ = c.Provide(repository.NewLogRepository)
	_ = c.Provide(repository.NewLogFeatureRepository)
	_ = c.Provide(repository.NewClusteringJobRepository)

	//Services
	_ = c.Provide(services.NewLogIdentityService)
	_ = c.Provide(services.NewLogEngineService)
	_ = c.Provide(services.NewLogFeatureService)
	_ = c.Provide(services.NewS3Service)
	_ = c.Provide(services.NewLogsConsumer)
	_ = c.Provide(services.NewLogFeaturesConsumer)
	_ = c.Provide(services.NewClusteringService)
	_ = c.Provide(services.NewClusteringCleanupService)

	//Interceptors
	_ = c.Provide(interceptors.NewAuthInterceptor)
	_ = c.Provide(interceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewLogEngineHandler)

	return c
}
