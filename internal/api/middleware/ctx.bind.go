package middleware

import (
	consts "logtheus/internal/constants"
	"net/http"

	"github.com/gin-gonic/gin"
)

func BindDTO[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var dto T
		if err := c.ShouldBind(&dto); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set(consts.DTO_KEY, dto)
		c.Next()
	}
}
