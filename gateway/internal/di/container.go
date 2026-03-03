package di

import (
	"logtheus/gateway/internal/api/controllers"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/config"
	"logtheus/shared/pkg/clients"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	userProto "logtheus/shared/pkg/pb/v1/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func Build(cfg *config.AppConfig) *dig.Container {
	c := dig.New()

	// Core singletons
	_ = c.Provide(func() *config.AppConfig { return cfg })

	// Clients
	_ = c.Provide(func(cfg *config.AppConfig) userProto.UserServiceClient {
		return clients.NewUserClient(cfg.Services.User)
	})
	_ = c.Provide(func(cfg *config.AppConfig) projectProto.ProjectServiceClient {
		return clients.NewProjectClient(cfg.Services.Project)
	})

	//Controllers
	_ = c.Provide(controllers.NewUserController)
	_ = c.Provide(controllers.NewProjectController)
	_ = c.Provide(controllers.NewInvitesController)

	// Middlewares
	_ = c.Provide(func(userClient userProto.UserServiceClient) gin.HandlerFunc {
		return middleware.AuthMiddleware(false, userClient)
	})

	return c
}
