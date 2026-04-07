package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/utils"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	sharedUtils "logtheus/shared/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	var timestamp *timestamppb.Timestamp
	var expiresAt string
	if data.ExpiresAt != nil {
		expiresAt = *data.ExpiresAt
	}
	expiration, err := time.Parse(time.RFC3339, expiresAt)
	if err == nil {
		timestamp = timestamppb.New(expiration)
	}
	response, err := c.projectClient.CreateInvite(grpcCtx, &projectProto.InviteUserRequest{
		ProjectId: data.ProjectID,
		Email:     data.Email,
		Role:      sharedUtils.HttpRoleToGRPCRole(data.Role),
		ExpiresAt: timestamp,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"ok": response.Ok})
}
