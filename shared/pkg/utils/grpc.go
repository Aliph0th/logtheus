package utils

import (
	"context"
	"fmt"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/types"

	"google.golang.org/grpc/metadata"
)

func GetGRPCContextWithAuth(ctx context.Context) context.Context {
	grpcCtx := context.Background()

	authData := ctx.Value(consts.AUTH_CONTEXT_KEY).(*types.UserAuthPayload)
	if authData == nil {
		return grpcCtx
	}

	grpcCtx = metadata.AppendToOutgoingContext(
		grpcCtx,
		consts.X_USER_ID_METADATA_KEY, fmt.Sprintf("%d", authData.UserID),
		consts.X_EMAIL_VERIFIED_METADATA_KEY, fmt.Sprintf("%v", authData.IsEmailVerified),
	)

	return grpcCtx
}
