package service

import (
	"fmt"
	"logtheus/internal/api/dto"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/consts"
	"logtheus/internal/models"
	"logtheus/internal/repository"
	"logtheus/internal/utils"

	"github.com/gin-gonic/gin"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo}
}

func (s *ProjectService) CreateProject(ctx *gin.Context, dto *dto.ProjectCreateRequest) (*models.Project, error) {
	auth := utils.MustAuth(ctx)

	existingCount, _ := s.repo.CountUserProjects(auth.UserID)
	if existingCount >= consts.MAX_PROJECTS_PER_USER {
		return nil, excepts.WithConflict(fmt.Sprintf("project limit of %d reached", consts.MAX_PROJECTS_PER_USER))
	}
	project := &models.Project{
		Name:        dto.Name,
		Description: &dto.Description,
	}
	err := s.repo.Create(project)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) GetMyProjects(ctx *gin.Context) ([]*models.Project, error) {
	auth := utils.MustAuth(ctx)

	projects, err := s.repo.GetByUserID(auth.UserID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}
