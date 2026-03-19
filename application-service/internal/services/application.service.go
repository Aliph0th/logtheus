package services

import (
	"context"
	"fmt"
	"logtheus/application/internal/config"
	"logtheus/application/internal/consts"
	"logtheus/application/internal/models"
	"logtheus/application/internal/repository"
	"logtheus/shared/pkg/grpc"
	applicationProto "logtheus/shared/pkg/pb/v1/application"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	"logtheus/shared/pkg/utils"
)

type ApplicationService struct {
	repo          *repository.ApplicationRepository
	projectClient projectProto.ProjectServiceClient
	apiKeyService *APIKeyService
	cfg           *config.AppConfig
}

func NewApplicationService(
	repo *repository.ApplicationRepository,
	projectClient projectProto.ProjectServiceClient,
	apiKeyService *APIKeyService,
	cfg *config.AppConfig,
) *ApplicationService {
	return &ApplicationService{
		repo:          repo,
		projectClient: projectClient,
		apiKeyService: apiKeyService,
		cfg:           cfg,
	}
}

func (s *ApplicationService) CreateApplication(
	ctx context.Context,
	dto *applicationProto.CreateApplicationRequest,
) (*models.Application, uint8, string, error) {
	auth := utils.MustUserData(ctx)

	roleResp, err := s.projectClient.GetMemberRole(ctx, &projectProto.GetMemberRoleRequest{
		UserId:    auth.UserID,
		ProjectId: dto.ProjectId,
	})
	if err != nil || roleResp == nil {
		return nil, 0, "", grpc.WithNotFound("Project not found or you are not a member")
	}

	existingCount, _ := s.repo.CountProjectApplications(dto.ProjectId)
	max := s.cfg.Settings.MaxApplicationsPerProject
	if existingCount >= max {
		err := grpc.WithResourceExhausted(fmt.Sprintf("Application limit of %d per project reached", max))
		return nil, 0, "", err
	}

	application := &models.Application{
		Name:        dto.Name,
		Description: &dto.Description,
		ProjectID:   dto.ProjectId,
	}
	err = s.repo.Create(application)
	if err != nil {
		return nil, 0, "", err
	}

	apiKey, err := s.apiKeyService.CreateAPIKey(consts.PREFIX_API_KEY, application.ID)
	if err != nil {
		return nil, 0, "", err
	}
	return application, max, apiKey, nil
}

func (s *ApplicationService) UpdateApplication(
	ctx context.Context,
	dto *applicationProto.UpdateApplicationRequest,
) (*models.Application, error) {
	auth := utils.MustUserData(ctx)

	app, err := s.repo.GetByID(dto.ApplicationId)
	if err != nil {
		return nil, err
	}

	roleResp, err := s.projectClient.GetMemberRole(ctx, &projectProto.GetMemberRoleRequest{
		UserId:    auth.UserID,
		ProjectId: app.ProjectID,
	})
	if err != nil || roleResp == nil {
		return nil, grpc.WithNotFound("Project not found or you are not a member")
	}
	if roleResp.Role == projectProto.Role_VIEWER {
		return nil, grpc.WithPermissionDenied("Insufficient permissions to update application")
	}

	if dto.Name != nil {
		app.Name = *dto.Name
	}
	if dto.Description != nil {
		app.Description = dto.Description
	}

	if err := s.repo.Update(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, applicationID uint64) error {
	auth := utils.MustUserData(ctx)

	app, err := s.repo.GetByID(applicationID)
	if err != nil {
		return err
	}

	roleResp, err := s.projectClient.GetMemberRole(ctx, &projectProto.GetMemberRoleRequest{
		UserId:    auth.UserID,
		ProjectId: app.ProjectID,
	})
	if err != nil || roleResp == nil {
		return grpc.WithNotFound("Project not found or you are not a member")
	}
	if roleResp.Role == projectProto.Role_VIEWER {
		return grpc.WithPermissionDenied("Insufficient permissions to delete application")
	}

	return s.repo.Delete(applicationID)
}

func (s *ApplicationService) GetApplicationsByProjectID(ctx context.Context, projectID uint64) ([]*models.Application, uint32, error) {
	auth := utils.MustUserData(ctx)

	roleResp, err := s.projectClient.GetMemberRole(ctx, &projectProto.GetMemberRoleRequest{
		UserId:    auth.UserID,
		ProjectId: projectID,
	})
	if err != nil || roleResp == nil {
		return nil, 0, grpc.WithNotFound("Project not found or you are not a member")
	}

	apps, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, 0, err
	}
	return apps, uint32(s.cfg.Settings.MaxApplicationsPerProject), nil
}

func (s *ApplicationService) GetApplicationByID(ctx context.Context, applicationID uint64) (*models.Application, error) {
	auth := utils.MustUserData(ctx)

	app, err := s.repo.GetByID(applicationID)
	if err != nil {
		return nil, err
	}

	roleResp, err := s.projectClient.GetMemberRole(ctx, &projectProto.GetMemberRoleRequest{
		UserId:    auth.UserID,
		ProjectId: app.ProjectID,
	})
	if err != nil || roleResp == nil {
		return nil, grpc.WithNotFound("Application not found")
	}

	return app, nil
}
