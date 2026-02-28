package repository

import (
	"errors"
	"logtheus/project/internal/consts"
	"logtheus/project/internal/models"
	"logtheus/shared/pkg/grpc"
	"logtheus/shared/pkg/storages"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *storages.Database) *ProjectRepository {
	return &ProjectRepository{db.DB}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) GetByID(id uint64) (*models.Project, error) {
	var project models.Project
	if err := r.db.First(&project, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpc.WithNotFound("Project not found")
		}
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) CountUserProjects(userID uint64) (uint8, error) {
	var count int64
	if err := r.db.Model(&models.Project{}).Where("owner_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint8(count), nil
}

func (r *ProjectRepository) CountProjectMembers(projectID uint64) (uint8, error) {
	var count int64
	if err := r.db.Model(&models.ProjectMember{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint8(count), nil
}

func (r *ProjectRepository) GetByUserID(userID uint64) ([]*models.Project, error) {
	var projects []*models.Project
	if err := r.db.Where("owner_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) GetMemberRole(userID, projectID uint64) (*consts.ProjectRole, error) {
	var member models.ProjectMember
	if err := r.db.First(&member, "user_id = ? AND project_id = ? AND joined_at IS NOT NULL", userID, projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &member.Role, nil
}
