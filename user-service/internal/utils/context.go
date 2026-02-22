package utils

import (
	"context"
	"logtheus/shared/pkg/types"
	"logtheus/user/internal/consts"
)

func MustUserData(c context.Context) *types.UserAuthClaims {
	return c.Value(consts.AUTH_CONTEXT_KEY).(*types.UserAuthClaims)
}
