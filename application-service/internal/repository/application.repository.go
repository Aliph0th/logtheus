package repository

import (
	"errors"
	"logtheus/application/internal/models"
	"logtheus/shared/pkg/grpc"
	"logtheus/shared/pkg/storages"

	"gorm.io/gorm"
)

type ApplicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *storages.Database) *ApplicationRepository {
	return &ApplicationRepository{db.DB}
}

func (r *ApplicationRepository) Create(app *models.Application) error {
	return r.db.Create(app).Error
}

func (r *ApplicationRepository) CountProjectApplications(projectID uint64) (uint8, error) {
	var count int64
	if err := r.db.Model(&models.Application{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return 0, err
	}
	return uint8(count), nil
}

func (r *ApplicationRepository) GetByID(id uint64) (*models.Application, error) {
	var app models.Application
	if err := r.db.First(&app, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpc.WithNotFound("Application not found")
		}
		return nil, err
	}
	return &app, nil
}

func (r *ApplicationRepository) Update(app *models.Application) error {
	return r.db.Save(app).Error
}

func (r *ApplicationRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Application{}, "id = ?", id).Error
}

func (r *ApplicationRepository) GetByProjectID(projectID uint64) ([]*models.Application, error) {
	var apps []*models.Application
	if err := r.db.Where("project_id = ?", projectID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *ApplicationRepository) SaveApiKey(key *models.ApplicationKey) error {
	return r.db.Create(key).Error
}
