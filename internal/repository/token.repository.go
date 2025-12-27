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

func (r *TokenRepository) GetByToken(token string) (*models.Token, error) {
	var tokenModel models.Token
	if err := r.db.First(&tokenModel, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &tokenModel, nil
}

func (r *TokenRepository) GetTokensByUserID(userID uint) ([]models.Token, error) {
	var tokens []models.Token
	if err := r.db.Find(&tokens, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *TokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.Token{}).Error
}

func (r *TokenRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Token{}).Error
}
