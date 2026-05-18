package api

import (
	"logtheus/gateway/internal/api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func NewRouter(container *dig.Container, origin string) *gin.Engine {
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{origin}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	api := router.Group("/api/v1")
	{
		routes.RegisterUserRoutes(api, container)
		routes.RegisterProjectRoutes(api, container)
		routes.RegisterApplicationRoutes(api, container)
		routes.RegisterLogRoutes(api, container)
	}
	return router
}
