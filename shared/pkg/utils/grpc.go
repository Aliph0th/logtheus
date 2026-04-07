package utils

import (
	"context"
	"fmt"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/types"

	"google.golang.org/grpc/metadata"
)

func GetGRPCContextWithAuth(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	authValue := ctx.Value(consts.AUTH_CONTEXT_KEY)
	authData, ok := authValue.(*types.UserAuthPayload)
	if !ok || authData == nil {
		return ctx
	}

	return metadata.AppendToOutgoingContext(
		ctx,
		consts.X_USER_ID_METADATA_KEY, fmt.Sprintf("%d", authData.UserID),
		consts.X_EMAIL_VERIFIED_METADATA_KEY, fmt.Sprintf("%v", authData.IsEmailVerified),
	)
}
