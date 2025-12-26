package utils

import (
	"logtheus/internal/consts"

	"github.com/gin-gonic/gin"
)

func MustDTO[T any](c *gin.Context) T {
	return c.MustGet(consts.DTO_KEY).(T)
}
