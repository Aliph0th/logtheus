package middleware

import (
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/shared/pkg/consts"

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
		excepts.RespondError(ctx, excepts.WithBadRequest("Validation failed").WithDetails(result).WithCode(consts.ERROR_CODE_VALIDATION_FAILED))
		return
	}

	ctx.Next()
}
