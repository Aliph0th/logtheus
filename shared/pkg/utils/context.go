package utils

import (
	"context"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/types"
)

func MustUserData(c context.Context) *types.UserAuthClaims {
	return c.Value(consts.AUTH_CONTEXT_KEY).(*types.UserAuthClaims)
}
