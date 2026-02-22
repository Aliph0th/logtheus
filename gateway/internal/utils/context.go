package utils

import (
	"logtheus/gateway/internal/consts"
	"logtheus/shared/pkg/types"

	"github.com/gin-gonic/gin"
)

func MustDTO[T any](c *gin.Context) T {
	return c.MustGet(consts.DTO_KEY).(T)
}

func GetAuthPayload(c *gin.Context) *types.UserAuthPayload {
	if payload, exists := c.Get(consts.AUTH_PAYLOAD_KEY); exists {
		return payload.(*types.UserAuthPayload)
	}
	return nil
}
