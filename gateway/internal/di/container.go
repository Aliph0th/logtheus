package di

import (
	"logtheus/gateway/internal/api/controllers"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/clients"
	"logtheus/gateway/internal/config"
	userProto "logtheus/shared/pkg/pb/v1/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })

	// Clients
	_ = c.Provide(clients.NewUserClient)

	//Controllers
	_ = c.Provide(controllers.NewUserController)

	// Middlewares
	_ = c.Provide(func(userClient userProto.UserServiceClient) gin.HandlerFunc {
		return middleware.AuthMiddleware(false, userClient)
	})

	return c
}
