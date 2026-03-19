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

func RegisterApplicationRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.ApplicationController](container)
	authMiddleware := utils.MustResolve[gin.HandlerFunc](container)

	applications := api.Group("/applications")
	{
		applications.Use(authMiddleware)

		applications.POST("/", append(
			validators.CreateApplicationValidators,
			middleware.BindDTO[*dto.ApplicationCreateRequest](),
			controller.CreateApplication,
		)...)

		applications.PATCH("/:id", append(
			[]gin.HandlerFunc{validators.DatabaseID("id")},
			append(
				validators.UpdateApplicationValidators,
				middleware.BindDTO[*dto.ApplicationUpdateRequest](),
				controller.UpdateApplication,
			)...,
		)...)

		applications.GET("/project/:project_id", validators.DatabaseID("project_id"), controller.GetApplicationsByProjectID)
		applications.GET("/:id", validators.DatabaseID("id"), controller.GetApplicationByID)

		applications.DELETE("/:id", validators.DatabaseID("id"), controller.DeleteApplication)
	}
}
