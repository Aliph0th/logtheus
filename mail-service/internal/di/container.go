package di

import (
	"logtheus/mail/internal/api"
	"logtheus/mail/internal/config"
	"logtheus/mail/internal/services"
	sharedInterceptors "logtheus/shared/pkg/interceptors"

	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })

	// Services
	_ = c.Provide(services.NewMailService)

	//Interceptors
	_ = c.Provide(sharedInterceptors.NewErrorInterceptor)

	//handlers
	_ = c.Provide(api.NewMailHandler)

	return c
}
