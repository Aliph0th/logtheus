package interceptors

import (
	"context"
	"logtheus/shared/pkg/types"
	"logtheus/user/internal/consts"
	"logtheus/user/internal/services"
	"strconv"

	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	tokenService *services.TokenService
}

func NewAuthInterceptor(tokenService *services.TokenService) *AuthInterceptor {
	return &AuthInterceptor{
		tokenService: tokenService,
	}
}

func (i *AuthInterceptor) Unary() grpcLib.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpcLib.UnaryServerInfo,
		handler grpcLib.UnaryHandler,
	) (any, error) {
		userID, isEmailVerified, found := i.extractUserDataFromMetadata(ctx)
		requestCtx := ctx
		if found {
			authPayload := &types.UserAuthPayload{
				UserID:          userID,
				IsEmailVerified: isEmailVerified,
			}
			requestCtx = context.WithValue(requestCtx, consts.AUTH_CONTEXT_KEY, authPayload)
		}
		return handler(requestCtx, req)

	}
}

func (i *AuthInterceptor) extractUserDataFromMetadata(ctx context.Context) (uint64, bool, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, false, false
	}

	var userID uint64
	var isEmailVerified bool

	if userIDStr := md.Get("x-user-id"); len(userIDStr) > 0 {
		if id, err := strconv.ParseUint(userIDStr[0], 10, 64); err == nil {
			userID = id
		}
	}

	if emailVerificationStr := md.Get("x-email-verified"); len(emailVerificationStr) > 0 {
		isEmailVerified = emailVerificationStr[0] == "true"
	}

	return userID, isEmailVerified, userID > 0
}
