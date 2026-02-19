package interceptors

import (
	"context"
	"errors"
	"fmt"
	"logtheus/shared/pkg/grpc"
	"logtheus/user/internal/consts"
	service "logtheus/user/internal/services"
	"strings"

	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct {
	tokenService *service.TokenService
}

func NewAuthInterceptor(tokenService *service.TokenService) *AuthInterceptor {
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
		if consts.PUBLIC_METHODS[info.FullMethod] {
			return handler(ctx, req)
		}

		token, err := i.extractTokenFromContext(ctx)
		if err != nil {
			// Return proper gRPC status error
			return nil, grpc.WithUnauthenticated(fmt.Sprintf("authentication failed: %v", err)).ToGRPCStatus()
		}

		claims, err := i.tokenService.VerifyAccessToken(token)
		if err != nil {
			return nil, grpc.WithUnauthenticated("invalid or expired token").ToGRPCStatus()
		}

		newCtx := context.WithValue(ctx, consts.AUTH_CONTEXT_KEY, claims)

		return handler(newCtx, req)
	}
}

func (i *AuthInterceptor) extractTokenFromContext(ctx context.Context) (string, error) {
	data, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	authHeader := data.Get("authorization")
	if len(authHeader) == 0 {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization format: expected 'Bearer <token>'")
	}

	return parts[1], nil
}
