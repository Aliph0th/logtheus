package service

import (
	"fmt"
	"logtheus/internal/api/dto"
	"logtheus/internal/config"
	"logtheus/internal/consts"
	"logtheus/internal/models"
	"logtheus/internal/repository"
	"logtheus/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	repo *repository.TokenRepository
	cfg  *config.AppConfig
}

func NewTokenService(repo *repository.TokenRepository, cfg *config.AppConfig) *TokenService {
	return &TokenService{repo, cfg}
}

func (s *TokenService) SignAuthTokens(p *dto.UserAuthPayload) (string, string) {
	claims := dto.UserAuthClaims{
		UserID: p.UserID,
	}
	accessToken, _ := s.signJWTWithClaims(claims, s.cfg.JWT.AccessSecret, consts.ACCESS_TOKEN_TTL)
	refreshToken, _ := s.signJWTWithClaims(claims, s.cfg.JWT.RefreshSecret, consts.REFRESH_TOKEN_TTL)
	return accessToken, refreshToken
}

func (s *TokenService) VerifyAccessToken(token string) (*dto.UserAuthClaims, error) {
	return s.verifyJWT(token, s.cfg.JWT.AccessSecret)
}
func (s *TokenService) VerifyRefreshToken(token string) (*dto.UserAuthClaims, error) {
	return s.verifyJWT(token, s.cfg.JWT.RefreshSecret)
}

func (s *TokenService) IssueToken(userID uint, tokenType consts.TokenType, ttl time.Duration) (string, error) {
	token, err := s.generateToken(tokenType)
	if err != nil {
		return "", fmt.Errorf("Error generating token: %w", err)
	}
	s.repo.Create(&models.Token{
		UserID:    userID,
		Type:      tokenType,
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
	})
	return token, nil
}

func (s *TokenService) generateToken(tokenType consts.TokenType) (string, error) {
	switch tokenType {
	case consts.TYPE_RESET_TOKEN:
		return uuid.NewString(), nil
	case consts.TYPE_VERIFY_TOKEN:
		return utils.GenerateCryptoRandomInt(consts.VERIFY_TOKEN_LENGTH)
	default:
		return "", fmt.Errorf("Unsupported token type for generation: %s", tokenType)
	}
}

func (s *TokenService) verifyJWT(token, secret string) (*dto.UserAuthClaims, error) {
	var claims dto.UserAuthClaims
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(_ *jwt.Token) (interface{}, error) { return []byte(secret), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.cfg.JWT.Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %w", err)
	}
	if !parsedToken.Valid {
		return nil, fmt.Errorf("Token is not valid")
	}
	return &claims, nil
}

func (s *TokenService) signJWTWithClaims(claims dto.UserAuthClaims, secret string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    s.cfg.JWT.Issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
