package controllers

import (
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/utils"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	sharedUtils "logtheus/shared/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProjectController struct {
	projectClient projectProto.ProjectServiceClient
}

func NewProjectController(projectClient projectProto.ProjectServiceClient) *ProjectController {
	return &ProjectController{
		projectClient: projectClient,
	}
}

func (c *ProjectController) CreateProject(ctx *gin.Context) {
	data := utils.MustDTO[*dto.ProjectCreateRequest](ctx)

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.projectClient.CreateProject(grpcCtx, &projectProto.CreateProjectRequest{
		Name:        data.Name,
		Description: data.Description,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(
		http.StatusCreated,
		gin.H{"project": utils.FromGrpcToDTO(response.Project, &dto.ProjectDTO{}), "max": response.Max},
	)
}

func (c *ProjectController) GetMyProjects(ctx *gin.Context) {
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.projectClient.GetMyProjects(grpcCtx, &emptypb.Empty{})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	projects := make([]*dto.ProjectDTO, len(response.Projects))
	for i, p := range response.Projects {
		projects[i] = utils.FromGrpcToDTO(p, &dto.ProjectDTO{})
	}
	ctx.JSON(http.StatusOK, gin.H{"projects": projects, "max": response.Max})
}

func (c *ProjectController) GetProjectByID(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	project, err := c.projectClient.GetProjectById(grpcCtx, &projectProto.GetProjectByIdRequest{ProjectId: id})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"project": utils.FromGrpcToDTO(project, &dto.ProjectDTO{})})
}

func (c *ProjectController) GetProjectMembers(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	response, err := c.projectClient.GetProjectMembers(grpcCtx, &projectProto.GetProjectMembersRequest{ProjectId: id})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	members := make([]*dto.ProjectMemberDTO, len(response.Members))
	for i, member := range response.Members {
		item := &dto.ProjectMemberDTO{
			UserId: member.UserId,
			Role:   sharedUtils.GRPCRoleToHttpRole(member.Role),
		}
		if member.JoinedAt != nil {
			joinedAt := member.JoinedAt.AsTime().Unix()
			item.JoinedAt = &joinedAt
		}
		if member.User != nil {
			item.User = &dto.ProjectMemberUserDTO{
				Id:       member.User.Id,
				Email:    member.User.Email,
				Username: member.User.Username,
			}
		}
		members[i] = item
	}

	ctx.JSON(http.StatusOK, gin.H{"members": members})
}

func (c *ProjectController) UpdateProject(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	data := utils.MustDTO[*dto.ProjectUpdateRequest](ctx)

	if data.Name == nil && data.Description == nil {
		excepts.RespondError(ctx, excepts.WithBadRequest("At least one field (name or description) must be provided"))
		return
	}

	grpcCtx := utils.GetGRPCContextWithAuth(ctx)
	project, err := c.projectClient.UpdateProject(grpcCtx, &projectProto.UpdateProjectRequest{
		ProjectId:   id,
		Name:        data.Name,
		Description: data.Description,
	})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"project": utils.FromGrpcToDTO(project, &dto.ProjectDTO{})})
}

func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	grpcCtx := utils.GetGRPCContextWithAuth(ctx)

	_, err := c.projectClient.DeleteProject(grpcCtx, &projectProto.DeleteProjectRequest{ProjectId: id})
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})

}
