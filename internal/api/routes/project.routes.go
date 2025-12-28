package routes

import (
	"logtheus/internal/api/controllers"
	"logtheus/internal/api/dto"
	"logtheus/internal/api/middleware"
	"logtheus/internal/api/validators"
	"logtheus/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

func RegisterProjectRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.ProjectController](container)
	authMiddleware := utils.MustResolve[gin.HandlerFunc](container)

	projects := api.Group("/projects")
	{
		projects.Use(authMiddleware)

		projects.POST("/", append(
			validators.CreateProjectValidators,
			middleware.BindDTO[dto.ProjectCreateRequest](),
			controller.CreateProject,
		)...)

		projects.GET("/my", controller.GetMyProjects)
		// projects.GET(":id")
		// projects.PUT(":id")

	}
}
