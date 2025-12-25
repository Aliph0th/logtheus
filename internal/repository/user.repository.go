package repository

import (
	"logtheus/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAll returns all users
func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
