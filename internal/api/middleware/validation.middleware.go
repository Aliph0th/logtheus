package middleware

import (
	"log/slog"
	sl "logtheus/internal/utils/logger"
	"net/http"

	gv "github.com/bube054/ginvalidator"
	"github.com/gin-gonic/gin"
)

func ValidationMiddleware(ctx *gin.Context) {
	result, err := gv.ValidationResult(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "The server encountered an unexpected error.",
		})
		slog.Error("Internal error", sl.Error(err))
		return
	}

	if len(result) != 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"errors": result,
		})
		return
	}

	ctx.Next()
}
