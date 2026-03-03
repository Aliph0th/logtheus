package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/utils"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InviteController struct {
	projectClient projectProto.ProjectServiceClient
}

func NewInvitesController(projectClient projectProto.ProjectServiceClient) *InviteController {
	return &InviteController{
		projectClient: projectClient,
	}
}

func (c *InviteController) CreateInvite(ctx *gin.Context) {
	data := utils.MustDTO[*dto.InviteCreateRequest](ctx)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.projectClient.CreateInvite(grpcCtx, &projectProto.InviteUserRequest{
		ProjectId: data.ProjectID,
		Email:     data.Email,
		Role:      utils.HttpRoleToGRPCRole(data.Role),
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"ok": response.Ok})
}
