package middleware

import (
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/consts"
	appProto "logtheus/shared/pkg/pb/v1/application"

	sharedConsts "logtheus/shared/pkg/consts"
	"strings"

	"github.com/gin-gonic/gin"
)

func ApiKeyMiddleware(appClient appProto.ApplicationServiceClient) gin.HandlerFunc {

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
		valid, err := appClient.ValidateApiKeyLight(ctx.Request.Context(), &appProto.ValidateApiKeyRequest{ApiKey: token})
		if err != nil || !valid.Valid {
			excepts.RespondError(
				ctx,
				excepts.WithUnauthorized("Invalid API key").WithCode(sharedConsts.ERROR_CODE_INVALID_API_KEY),
			)
			return
		}

		ctx.Set(consts.API_KEY, token)
		ctx.Next()
	}
}
