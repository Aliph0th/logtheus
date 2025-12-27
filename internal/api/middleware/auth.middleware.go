package middleware

import (
	"logtheus/internal/api/dto"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/consts"
	"logtheus/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(allowNotVerifiedEmail bool, tokenService *service.TokenService) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			excepts.RespondError(ctx, excepts.WithUnauthorized("Authorization header missing"))
			return
		}

		prefix, token, found := strings.Cut(header, " ")
		if !found || strings.ToLower(prefix) != "bearer" {
			excepts.RespondError(ctx, excepts.WithUnauthorized("Invalid authorization header format"))
			return
		}
		claims, err := tokenService.VerifyAccessToken(token)
		if err != nil {
			excepts.RespondError(ctx, excepts.WithUnauthorized(err.Error()))
			return
		}
		var payload dto.UserAuthPayload = claims.UserAuthPayload
		if !allowNotVerifiedEmail && !payload.IsEmailVerified {
			excepts.RespondError(ctx, excepts.WithUnauthorized("Email is not verified"))
			return
		}

		ctx.Set(consts.AUTH_PAYLOAD_KEY, payload)
		ctx.Next()
	}
}
