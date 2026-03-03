package utils

import (
	"context"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/types"
)

func MustUserData(c context.Context) *types.UserAuthPayload {
	return c.Value(consts.AUTH_CONTEXT_KEY).(*types.UserAuthPayload)
}
