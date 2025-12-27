package utils

import (
	"logtheus/internal/api/dto"
	"logtheus/internal/consts"

	"github.com/gin-gonic/gin"
)

func MustDTO[T any](c *gin.Context) T {
	return c.MustGet(consts.DTO_KEY).(T)
}

func MustAuth(c *gin.Context) dto.UserAuthPayload {
	return c.MustGet(consts.AUTH_PAYLOAD_KEY).(dto.UserAuthPayload)
}
