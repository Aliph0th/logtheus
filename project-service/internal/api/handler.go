package api

import (
	"context"
	"logtheus/project/internal/services"
	projectProto "logtheus/shared/pkg/pb/v1/project"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProjectHandler struct {
	projectProto.UnimplementedProjectServiceServer
	projectService *services.ProjectService
	invitesService *services.InvitesService
}

func NewProjectHandler(projectService *services.ProjectService, invitesService *services.InvitesService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		invitesService: invitesService,
	}
}

func (h *ProjectHandler) CreateProject(ctx context.Context, req *projectProto.CreateProjectRequest) (*projectProto.CreateProjectResponse, error) {
	project, max, err := h.projectService.CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}
	return &projectProto.CreateProjectResponse{
		Project: &projectProto.Project{
			Id:          project.ID,
			Name:        project.Name,
			Description: *project.Description,
			CreatedAt:   timestamppb.New(project.CreatedAt),
			UpdatedAt:   timestamppb.New(project.UpdatedAt),
		},
		Max: uint32(max),
	}, nil
}

func (h *ProjectHandler) GetMyProjects(ctx context.Context, req *emptypb.Empty) (*projectProto.GetMyProjectsResponse, error) {
	projects, max, err := h.projectService.GetMyProjects(ctx)
	if err != nil {
		return nil, err
	}
	protoProjects := make([]*projectProto.Project, len(projects))
	for i, p := range projects {
		protoProjects[i] = &projectProto.Project{
			Id:          p.ID,
			Name:        p.Name,
			Description: *p.Description,
			CreatedAt:   timestamppb.New(p.CreatedAt),
			UpdatedAt:   timestamppb.New(p.UpdatedAt),
		}
	}
	return &projectProto.GetMyProjectsResponse{
		Projects: protoProjects,
		Max:      uint32(max),
	}, nil
}

func (h *ProjectHandler) GetProjectByID(ctx context.Context, req *projectProto.GetProjectByIdRequest) (*projectProto.Project, error) {
	project, err := h.projectService.GetProjectByID(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &projectProto.Project{
		Id:          project.ID,
		Name:        project.Name,
		Description: *project.Description,
		CreatedAt:   timestamppb.New(project.CreatedAt),
		UpdatedAt:   timestamppb.New(project.UpdatedAt),
	}, nil
}

func (h *ProjectHandler) CreateInvite(ctx context.Context, req *projectProto.InviteUserRequest) (*projectProto.InviteUserResponse, error) {
	err := h.invitesService.CreateInvite(ctx, req)
	if err != nil {
		return nil, err
	}
	return &projectProto.InviteUserResponse{Ok: true}, nil
}
