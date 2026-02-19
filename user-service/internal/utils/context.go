package utils

import (
	"context"
	"logtheus/shared/pkg/consts"
	"logtheus/user/types"
)

func MustUserData(c context.Context) *types.UserAuthPayload {
	return c.Value(consts.AUTH_PAYLOAD_KEY).(*types.UserAuthPayload)
}
