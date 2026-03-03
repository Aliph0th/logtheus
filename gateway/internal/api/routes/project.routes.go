package routes

import (
	"logtheus/gateway/internal/api/controllers"
	"logtheus/gateway/internal/api/dto"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/api/validators"
	"logtheus/shared/pkg/utils"

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
			middleware.BindDTO[*dto.ProjectCreateRequest](),
			controller.CreateProject,
		)...)

		projects.GET("/my", controller.GetMyProjects)
		projects.GET(":id", validators.DatabaseID("id"), controller.GetProjectByID)
		// projects.PUT(":id")

		RegisterInvitesRoutes(projects, container)
	}
}
