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

func RegisterInvitesRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.InviteController](container)

	invites := api.Group("/invites")
	{
		invites.POST("/", append(
			validators.CreateInviteValidators,
			middleware.BindDTO[dto.InviteCreateRequest](),
			controller.CreateInvite,
		)...)
	}
}
