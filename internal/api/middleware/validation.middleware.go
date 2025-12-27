package middleware

import (
	excepts "logtheus/internal/api/exceptions"
	"net/http"

	gv "github.com/bube054/ginvalidator"
	"github.com/gin-gonic/gin"
)

func ValidationMiddleware(ctx *gin.Context) {
	result, err := gv.ValidationResult(ctx)
	if err != nil {
		excepts.RespondError(ctx, err)
		return
	}

	if len(result) != 0 {
		excepts.RespondError(ctx, excepts.NewWithDetails(http.StatusBadRequest, "Validation failed", result))
		return
	}

	ctx.Next()
}
