package api

import (
	"logtheus/internal/api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	api := router.Group("/api/v1")
	{
		routes.RegisterUserRoutes(api, db)
	}
	return router
}
