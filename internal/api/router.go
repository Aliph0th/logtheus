package api

import (
	"logtheus/internal/api/middleware"
	"logtheus/internal/api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func NewRouter(container *dig.Container) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(middleware.ErrorHandler())

	api := router.Group("/api/v1")
	{
		routes.RegisterUserRoutes(api, container)
	}
	return router
}
