package services

import (
	"context"
	"fmt"
	"logtheus/project/internal/config"
	"logtheus/project/internal/models"
	"logtheus/project/internal/repository"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	"logtheus/shared/pkg/utils"
	"time"
)

type ProjectService struct {
	repo *repository.ProjectRepository
	cfg  *config.AppConfig
}

func NewProjectService(repo *repository.ProjectRepository, cfg *config.AppConfig) *ProjectService {
	return &ProjectService{repo, cfg}
}

func (s *ProjectService) CreateProject(ctx context.Context, dto *projectProto.CreateProjectRequest) (*models.Project, uint8, error) {
	auth := utils.MustUserData(ctx)

	existingCount, _ := s.repo.CountUserProjects(auth.UserID)
	max := s.cfg.Settings.MaxProjectsPerUser
	if existingCount >= max {
		err := grpc.WithResourceExhausted(fmt.Sprintf("Project limit of %d per user reached", max))
		return nil, 0, err.WithSlug(consts.ERROR_CODE_TOO_MANY_PROJECTS)
	}
	project := &models.Project{
		Name:        dto.Name,
		Description: &dto.Description,
		OwnerID:     auth.UserID,
	}
	err := s.repo.Create(project)
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	member := &models.ProjectMember{
		ProjectID: project.ID,
		UserID:    auth.UserID,
		Role:      consts.PROJECT_ROLE_OWNER,
		JoinedAt:  &now,
	}
	err = s.repo.AddMember(member)
	if err != nil {
		return nil, 0, err
	}

	return project, max, nil
}

func (s *ProjectService) GetMyProjects(ctx context.Context) ([]*models.Project, uint8, error) {
	auth := utils.MustUserData(ctx)

	projects, err := s.repo.GetByUserID(auth.UserID)
	if err != nil {
		return nil, 0, err
	}
	return projects, s.cfg.Settings.MaxProjectsPerUser, nil
}

func (s *ProjectService) GetProjectByID(ctx context.Context, id uint64) (*models.Project, error) {
	auth := utils.MustUserData(ctx)
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !s.IsMember(auth.UserID, id) {
		return nil, grpc.WithNotFound("Project not found")
	}
	return project, nil
}

func (s *ProjectService) GetMemberRole(userID, projectID uint64) (*consts.ProjectRole, error) {
	return s.repo.GetMemberRole(userID, projectID)
}

func (s *ProjectService) IsMember(userID, projectID uint64) bool {
	role, _ := s.repo.GetMemberRole(userID, projectID)
	return role != nil
}

func (s *ProjectService) CountMembers(projectID uint64) (uint8, error) {
	return s.repo.CountProjectMembers(projectID)
}

func (s *ProjectService) getByID(id uint64) (*models.Project, error) {
	return s.repo.GetByID(id)
}
