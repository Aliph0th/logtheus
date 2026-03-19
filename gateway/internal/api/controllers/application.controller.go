package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/utils"
	applicationProto "logtheus/shared/pkg/pb/v1/application"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ApplicationController struct {
	applicationClient applicationProto.ApplicationServiceClient
}

func NewApplicationController(applicationClient applicationProto.ApplicationServiceClient) *ApplicationController {
	return &ApplicationController{
		applicationClient: applicationClient,
	}
}

func (c *ApplicationController) CreateApplication(ctx *gin.Context) {
	data := utils.MustDTO[*dto.ApplicationCreateRequest](ctx)

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.applicationClient.CreateApplication(grpcCtx, &applicationProto.CreateApplicationRequest{
		Name:        data.Name,
		Description: data.Description,
		ProjectId:   data.ProjectID,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(
		http.StatusCreated,
		gin.H{
			"application": utils.FromGrpcToDTO(response.Application, &dto.ApplicationDTO{}),
			"max":         response.Max,
			"api_key":     response.ApiKey,
		},
	)
}

func (c *ApplicationController) UpdateApplication(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	data := utils.MustDTO[*dto.ApplicationUpdateRequest](ctx)

	if data.Name == nil && data.Description == nil {
		excepts.RespondError(ctx, excepts.WithBadRequest("At least one field (name or description) must be provided"))
		return
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	application, err := c.applicationClient.UpdateApplication(grpcCtx, &applicationProto.UpdateApplicationRequest{
		ApplicationId: id,
		Name:          data.Name,
		Description:   data.Description,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"application": utils.FromGrpcToDTO(application, &dto.ApplicationDTO{})})
}

func (c *ApplicationController) DeleteApplication(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)

	_, err := c.applicationClient.DeleteApplication(grpcCtx, &applicationProto.DeleteApplicationRequest{ApplicationId: id})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

func (c *ApplicationController) GetApplicationsByProjectID(ctx *gin.Context) {
	projectID, _ := strconv.ParseUint(ctx.Param("project_id"), 10, 64)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)

	response, err := c.applicationClient.GetApplicationsByProjectId(grpcCtx, &applicationProto.GetApplicationsByProjectIdRequest{ProjectId: projectID})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	applications := make([]*dto.ApplicationDTO, len(response.Applications))
	for i, app := range response.Applications {
		applications[i] = utils.FromGrpcToDTO(app, &dto.ApplicationDTO{})
	}

	ctx.JSON(http.StatusOK, gin.H{"applications": applications, "max": response.Max})
}

func (c *ApplicationController) GetApplicationByID(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)

	application, err := c.applicationClient.GetApplicationById(grpcCtx, &applicationProto.GetApplicationByIdRequest{ApplicationId: id})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"application": utils.FromGrpcToDTO(application, &dto.ApplicationDTO{})})
}
