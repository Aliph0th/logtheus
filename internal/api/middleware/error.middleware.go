package middleware

import (
	excepts "logtheus/internal/api/exceptions"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			excepts.RespondError(ctx, ctx.Errors.Last().Err)
		}
	}
}
