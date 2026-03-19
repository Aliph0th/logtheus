package validators

import (
	"fmt"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/consts"

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
		return fmt.Sprintf(
			"Description must be between %d and %d characters long",
			consts.MIN_PROJECT_DESCRIPTION_LEN,
			consts.MAX_PROJECT_DESCRIPTION_LEN,
		)
	}).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_PROJECT_DESCRIPTION_LEN,
		Max: &consts.MAX_PROJECT_DESCRIPTION_LEN,
	}).Validate(),
	middleware.ValidationMiddleware,
}

var UpdateProjectValidators = []gin.HandlerFunc{
	gv.NewBody(
		"name",
		func(_, _, _ string) string {
			return fmt.Sprintf(
				"Project name must be between %d and %d characters",
				consts.MIN_PROJECT_NAME_LEN,
				consts.MAX_PROJECT_NAME_LEN,
			)
		},
	).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_PROJECT_NAME_LEN,
		Max: &consts.MAX_PROJECT_NAME_LEN,
	}).Validate(),
	gv.NewBody("description", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Description must be between %d and %d characters long",
			consts.MIN_PROJECT_DESCRIPTION_LEN,
			consts.MAX_PROJECT_DESCRIPTION_LEN,
		)
	}).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_PROJECT_DESCRIPTION_LEN,
		Max: &consts.MAX_PROJECT_DESCRIPTION_LEN,
	}).Validate(),
	middleware.ValidationMiddleware,
}
