package utils

import (
	"context"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/types"
)

func MustUserData(c context.Context) *types.UserAuthPayload {
	v, ok := c.Value(consts.AUTH_CONTEXT_KEY).(*types.UserAuthPayload)
	if !ok {
		return nil
	}
	return v
}
