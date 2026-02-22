package middleware

import (
	"context"
	excepts "logtheus/gateway/internal/api/exceptions"
	userProto "logtheus/shared/pkg/pb/v1/user"
	"logtheus/shared/pkg/types"

	"logtheus/gateway/internal/consts"
	sharedConsts "logtheus/shared/pkg/consts"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(allowNotVerifiedEmail bool, userClient userProto.UserServiceClient) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			excepts.RespondError(
				ctx,
				excepts.WithUnauthorized("Authorization header missing").WithCode(sharedConsts.ERROR_CODE_UNAUTHENTICATED),
			)
			return
		}

		prefix, token, found := strings.Cut(header, " ")
		if !found || strings.ToLower(prefix) != "bearer" {
			excepts.RespondError(
				ctx,
				excepts.WithUnauthorized("Invalid authorization header format").WithCode(sharedConsts.ERROR_CODE_UNAUTHENTICATED),
			)
			return
		}
		claims, err := userClient.ValidateToken(context.Background(), &userProto.ValidateTokenRequest{Token: token})
		if err != nil {
			excepts.RespondError(ctx, excepts.WithUnauthorized(err.Error()).WithCode(sharedConsts.ERROR_CODE_UNAUTHENTICATED))
			return
		}
		payload := &types.UserAuthPayload{
			UserID:          claims.UserId,
			IsEmailVerified: claims.IsEmailVerified,
		}
		if !allowNotVerifiedEmail && !payload.IsEmailVerified {
			excepts.RespondError(
				ctx,
				excepts.WithUnauthorized("Your email is not verified").WithCode(sharedConsts.ERROR_CODE_EMAIL_NOT_VERIFIED),
			)
			return
		}

		ctx.Set(consts.AUTH_PAYLOAD_KEY, payload)
		ctx.Next()
	}
}
