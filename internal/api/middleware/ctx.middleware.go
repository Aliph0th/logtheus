package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	excepts "logtheus/internal/api/exceptions"
	"logtheus/internal/consts"

	"github.com/gin-gonic/gin"
)

func BindDTO[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var dto T
		if err := ctx.ShouldBind(&dto); err != nil {
			var typeError *json.UnmarshalTypeError
			if errors.As(err, &typeError) && typeError != nil {
				expected := typeError.Type.String()
				excepts.RespondError(
					ctx,
					excepts.WithBadRequest(fmt.Sprintf("Field '%s' expects %s, got %s", typeError.Field, expected, typeError.Value)),
				)
				return
			}
			excepts.RespondError(ctx, excepts.WithBadRequest("Invalid request body"))
			return
		}
		ctx.Set(consts.DTO_KEY, dto)
		ctx.Next()
	}
}
