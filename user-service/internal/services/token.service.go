package services

import (
	"fmt"

	sharedConsts "logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	"logtheus/shared/pkg/types"
	"logtheus/user/internal/config"
	"logtheus/user/internal/consts"
	"logtheus/user/internal/repository"
	"logtheus/user/internal/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenService struct {
	repo *repository.TokenRepository
	cfg  *config.AppConfig
}

func NewTokenService(repo *repository.TokenRepository, cfg *config.AppConfig) *TokenService {
	return &TokenService{repo, cfg}
}

func (s *TokenService) SignAuthTokens(p *types.UserAuthPayload) (string, string) {
	claims := types.UserAuthClaims{
		UserAuthPayload: *p,
	}
	accessToken, _ := s.signJWTWithClaims(&claims, s.cfg.JWT.AccessSecret, consts.TTL_ACCESS_TOKEN)
	refreshToken, _ := s.signJWTWithClaims(&claims, s.cfg.JWT.RefreshSecret, consts.TTL_REFRESH_TOKEN)
	return accessToken, refreshToken
}

func (s *TokenService) VerifyAccessToken(token string) (*types.UserAuthClaims, error) {
	return s.verifyJWT(token, s.cfg.JWT.AccessSecret)
}
func (s *TokenService) VerifyRefreshToken(token string) (*types.UserAuthClaims, error) {
	return s.verifyJWT(token, s.cfg.JWT.RefreshSecret)
}

func (s *TokenService) IssueEmailVerificationToken(userID uint64) (string, error) {
	token, err := s.generateToken(consts.TOKEN_TYPE_VERIFY)
	if err != nil {
		return "", fmt.Errorf("Error generating token: %w", err)
	}
	key := consts.REDIS_KEY_EMAIL_VERIFICATION(userID)
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Error hashing token: %w", err)
	}
	s.repo.CreateInRedis(key, string(hashedToken), consts.TTL_VERIFY_TOKEN)
	return token, nil
}

func (s *TokenService) UseEmailVerificationToken(userID uint64, token string) error {
	key := consts.REDIS_KEY_EMAIL_VERIFICATION(userID)
	storedToken, err := s.repo.GetFromRedis(key)
	if err != nil {
		return err
	}
	if storedToken == "" {
		return fmt.Errorf("Verification token not found or expired")
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedToken), []byte(token))
	if err != nil {
		return fmt.Errorf("Invalid verification token")
	}
	s.repo.DeleteFromRedis(key)
	return nil
}

func (s *TokenService) IssueInviteToken() (string, error) {
	token, err := s.generateToken(consts.TOKEN_TYPE_INVITE)
	if err != nil {
		return "", grpc.WithInternal().WithSlug(sharedConsts.INTERNAL_ERROR_CODE_TOKEN_GENERATION_FAILED)
	}
	return token, nil
}

func (s *TokenService) generateToken(tokenType consts.TokenType) (string, error) {
	switch tokenType {
	case consts.TOKEN_TYPE_PASSWORD_RESET, consts.TOKEN_TYPE_INVITE:
		return uuid.NewString(), nil
	case consts.TOKEN_TYPE_VERIFY:
		return utils.GenerateCryptoRandomInt(consts.VERIFY_EMAIL_TOKEN_LENGTH)
	default:
		return "", fmt.Errorf("Unsupported token type for generation: %s", tokenType)
	}
}

func (s *TokenService) verifyJWT(token, secret string) (*types.UserAuthClaims, error) {
	var claims types.UserAuthClaims
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(_ *jwt.Token) (any, error) { return []byte(secret), nil },
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

func (s *TokenService) signJWTWithClaims(claims *types.UserAuthClaims, secret string, expiration time.Duration) (string, error) {
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
