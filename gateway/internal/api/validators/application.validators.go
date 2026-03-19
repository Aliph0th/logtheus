package validators

import (
	"fmt"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/consts"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var CreateApplicationValidators = []gin.HandlerFunc{
	gv.NewBody("project_id", func(_, _, _ string) string {
		return "Project ID must be a valid number"
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Validate(),

	gv.NewBody(
		"name",
		func(_, _, _ string) string {
			return fmt.Sprintf(
				"Application name must be between %d and %d characters",
				consts.MIN_APPLICATION_NAME_LEN,
				consts.MAX_APPLICATION_NAME_LEN,
			)
		},
	).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_APPLICATION_NAME_LEN,
		Max: &consts.MAX_APPLICATION_NAME_LEN,
	}).Validate(),
	gv.NewBody("description", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Description must be between %d and %d characters long",
			consts.MIN_APPLICATION_DESCRIPTION_LEN,
			consts.MAX_APPLICATION_DESCRIPTION_LEN,
		)
	}).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_APPLICATION_DESCRIPTION_LEN,
		Max: &consts.MAX_APPLICATION_DESCRIPTION_LEN,
	}).Validate(),
	middleware.ValidationMiddleware,
}

var UpdateApplicationValidators = []gin.HandlerFunc{
	gv.NewBody(
		"name",
		func(_, _, _ string) string {
			return fmt.Sprintf(
				"Application name must be between %d and %d characters",
				consts.MIN_APPLICATION_NAME_LEN,
				consts.MAX_APPLICATION_NAME_LEN,
			)
		},
	).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_APPLICATION_NAME_LEN,
		Max: &consts.MAX_APPLICATION_NAME_LEN,
	}).Validate(),
	gv.NewBody("description", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Description must be between %d and %d characters long",
			consts.MIN_APPLICATION_DESCRIPTION_LEN,
			consts.MAX_APPLICATION_DESCRIPTION_LEN,
		)
	}).Chain().Optional().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_APPLICATION_DESCRIPTION_LEN,
		Max: &consts.MAX_APPLICATION_DESCRIPTION_LEN,
	}).Validate(),
	middleware.ValidationMiddleware,
}
