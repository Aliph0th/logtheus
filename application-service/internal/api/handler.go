package api

import (
	"context"
	"logtheus/application/internal/services"
	applicationProto "logtheus/shared/pkg/pb/v1/application"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ApplicationHandler struct {
	applicationProto.UnimplementedApplicationServiceServer
	applicationService *services.ApplicationService
}

func NewApplicationHandler(applicationService *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
	}
}

func (h *ApplicationHandler) CreateApplication(ctx context.Context, req *applicationProto.CreateApplicationRequest) (*applicationProto.CreateApplicationResponse, error) {
	app, max, apiKey, err := h.applicationService.CreateApplication(ctx, req)
	if err != nil {
		return nil, err
	}
	return &applicationProto.CreateApplicationResponse{
		Application: &applicationProto.Application{
			Id:          app.ID,
			Name:        app.Name,
			Description: *app.Description,
			ProjectId:   app.ProjectID,
			CreatedAt:   timestamppb.New(app.CreatedAt),
			UpdatedAt:   timestamppb.New(app.UpdatedAt),
		},
		Max:    uint32(max),
		ApiKey: apiKey,
	}, nil
}

func (h *ApplicationHandler) UpdateApplication(ctx context.Context, req *applicationProto.UpdateApplicationRequest) (*applicationProto.Application, error) {
	app, err := h.applicationService.UpdateApplication(ctx, req)
	if err != nil {
		return nil, err
	}

	return &applicationProto.Application{
		Id:          app.ID,
		Name:        app.Name,
		Description: *app.Description,
		ProjectId:   app.ProjectID,
		CreatedAt:   timestamppb.New(app.CreatedAt),
		UpdatedAt:   timestamppb.New(app.UpdatedAt),
	}, nil
}

func (h *ApplicationHandler) DeleteApplication(ctx context.Context, req *applicationProto.DeleteApplicationRequest) (*emptypb.Empty, error) {
	if err := h.applicationService.DeleteApplication(ctx, req.ApplicationId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *ApplicationHandler) GetApplicationsByProjectId(ctx context.Context, req *applicationProto.GetApplicationsByProjectIdRequest) (*applicationProto.GetApplicationsByProjectIdResponse, error) {
	apps, max, err := h.applicationService.GetApplicationsByProjectID(ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	protoApps := make([]*applicationProto.Application, len(apps))
	for i, app := range apps {
		protoApps[i] = &applicationProto.Application{
			Id:          app.ID,
			Name:        app.Name,
			Description: *app.Description,
			ProjectId:   app.ProjectID,
			CreatedAt:   timestamppb.New(app.CreatedAt),
			UpdatedAt:   timestamppb.New(app.UpdatedAt),
		}
	}

	return &applicationProto.GetApplicationsByProjectIdResponse{Applications: protoApps, Max: max}, nil
}

func (h *ApplicationHandler) GetApplicationById(ctx context.Context, req *applicationProto.GetApplicationByIdRequest) (*applicationProto.Application, error) {
	app, err := h.applicationService.GetApplicationByID(ctx, req.ApplicationId)
	if err != nil {
		return nil, err
	}

	return &applicationProto.Application{
		Id:          app.ID,
		Name:        app.Name,
		Description: *app.Description,
		ProjectId:   app.ProjectID,
		CreatedAt:   timestamppb.New(app.CreatedAt),
		UpdatedAt:   timestamppb.New(app.UpdatedAt),
	}, nil
}
