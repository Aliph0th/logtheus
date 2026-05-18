package api

import (
	"context"
	"logtheus/project/internal/services"
	sharedGrpc "logtheus/shared/pkg/grpc"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	"logtheus/shared/pkg/utils"

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

func (h *ProjectHandler) UpdateProject(ctx context.Context, req *projectProto.UpdateProjectRequest) (*projectProto.Project, error) {
	project, err := h.projectService.UpdateProject(ctx, req)
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

func (h *ProjectHandler) DeleteProject(ctx context.Context, req *projectProto.DeleteProjectRequest) (*emptypb.Empty, error) {
	if err := h.projectService.DeleteProject(ctx, req.ProjectId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
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

func (h *ProjectHandler) GetProjectById(ctx context.Context, req *projectProto.GetProjectByIdRequest) (*projectProto.Project, error) {
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

func (h *ProjectHandler) GetMemberRole(ctx context.Context, req *projectProto.GetMemberRoleRequest) (*projectProto.GetMemberRoleResponse, error) {
	role, err := h.projectService.GetMemberRole(req.UserId, req.ProjectId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, sharedGrpc.WithNotFound("Project member not found")
	}

	return &projectProto.GetMemberRoleResponse{
		Role: utils.HttpRoleToGRPCRole(*role),
	}, nil
}

func (h *ProjectHandler) GetProjectMembers(ctx context.Context, req *projectProto.GetProjectMembersRequest) (*projectProto.GetProjectMembersResponse, error) {
	members, err := h.projectService.GetProjectMembers(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	return &projectProto.GetProjectMembersResponse{Members: members}, nil
}

func (h *ProjectHandler) CreateInvite(ctx context.Context, req *projectProto.InviteUserRequest) (*projectProto.InviteUserResponse, error) {
	err := h.invitesService.CreateInvite(ctx, req)
	if err != nil {
		return nil, err
	}
	return &projectProto.InviteUserResponse{Ok: true}, nil
}
