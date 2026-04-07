package di

import (
	"logtheus/mail/internal/config"
	"logtheus/mail/internal/services"

	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })

	// Services
	_ = c.Provide(services.NewMailService)
	_ = c.Provide(services.NewMailConsumer)

	return c
}
