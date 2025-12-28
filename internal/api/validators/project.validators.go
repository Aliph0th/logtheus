package validators

import (
	"fmt"
	"logtheus/internal/api/middleware"
	"logtheus/internal/consts"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var CreateProjectValidators = []gin.HandlerFunc{
	gv.NewBody(
		"name",
		func(_, _, _ string) string {
			return fmt.Sprintf(
				"Project name must be between %d and %d characters",
				consts.MIN_PROJECT_NAME_LEN,
				consts.MAX_PROJECT_NAME_LEN,
			)
		},
	).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_PROJECT_NAME_LEN,
		Max: &consts.MAX_PROJECT_NAME_LEN,
	}).Validate(),
	gv.NewBody("description", func(_, _, _ string) string {
		return "Description must be at most 500 characters long"
	}).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_PROJECT_DESCRIPTION_LEN,
		Max: &consts.MAX_PROJECT_DESCRIPTION_LEN,
	}).Validate(),
	middleware.ValidationMiddleware,
}
