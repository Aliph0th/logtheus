package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"logtheus/application/internal/config"
	"logtheus/application/internal/consts"
	"logtheus/application/internal/models"
	"logtheus/application/internal/repository"
)

type APIKeyService struct {
	cfg  *config.AppConfig
	repo *repository.ApplicationRepository
}

func NewAPIKeyService(cfg *config.AppConfig, repo *repository.ApplicationRepository) *APIKeyService {
	return &APIKeyService{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *APIKeyService) CreateAPIKey(prefix consts.ApiKeyPrefix, applicationID uint64) (string, error) {
	token, err := s.generateLongToken()
	if err != nil {
		return "", err
	}
	signature := s.signature(token)
	hashedToken := s.hashToken(token)
	apiKey := &models.ApplicationKey{
		Signature:     signature,
		TokenHash:     hashedToken,
		Prefix:        prefix,
		ApplicationID: applicationID,
	}
	if err := s.repo.SaveApiKey(apiKey); err != nil {
		return "", fmt.Errorf("failed to save API key: %w", err)
	}
	return s.FormatApiKey(token, signature, prefix), nil
}

func (s *APIKeyService) FormatApiKey(token, signature string, prefix consts.ApiKeyPrefix) string {
	return fmt.Sprintf("%s_%s_%s", prefix, signature, token)
}

func (s *APIKeyService) generateLongToken() (string, error) {
	randomBytes := make([]byte, s.cfg.ApiKey.BytesLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	token := hex.EncodeToString(randomBytes)
	return token, nil
}

func (s *APIKeyService) hashToken(token string) string {
	sha256Hash := sha256.New()
	sha256Hash.Write([]byte(token))
	sha256Sum := sha256Hash.Sum(nil)
	return hex.EncodeToString(sha256Sum)
}

func (s *APIKeyService) signature(apiKey string) string {
	hmacKey := []byte(s.cfg.ApiKey.Secret)

	hmacHash := hmac.New(sha256.New, hmacKey)
	hmacHash.Write([]byte(apiKey))
	hmacSum := hmacHash.Sum(nil)
	return hex.EncodeToString(hmacSum)
}
