package repository

import (
	"logtheus/internal/models"

	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

func (r *TokenRepository) GetByID(id uint) (*models.Token, error) {
	var token models.Token
	if err := r.db.First(&token, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepository) GetTokensByUserID(userID uint) ([]models.Token, error) {
	var tokens []models.Token
	if err := r.db.Find(&tokens, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *TokenRepository) DeleteByID(id uint) error {
	return r.db.Delete(&models.Token{}, id).Error
}

func (r *TokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Token{}).Error
}
