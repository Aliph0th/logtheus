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

func RegisterInvitesRoutes(api *gin.RouterGroup, container *dig.Container) {
	controller := utils.MustResolve[*controllers.InviteController](container)

	invites := api.Group("/invites")
	{
		invites.POST("/", append(
			validators.CreateInviteValidators,
			middleware.BindDTO[*dto.InviteCreateRequest](),
			controller.CreateInvite,
		)...)
	}
}
