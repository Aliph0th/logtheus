package repository

import (
	"errors"

	"logtheus/shared/pkg/grpc"
	"logtheus/user/internal/models"

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

func (r *UserRepository) GetByID(id uint64) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpc.WithNotFound("User not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, grpc.WithNotFound("User not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) VerifyEmail(userID uint64) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("is_email_verified", true).Error
}
