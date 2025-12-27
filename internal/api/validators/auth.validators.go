package validators

import (
	"fmt"
	"logtheus/internal/api/middleware"
	"logtheus/internal/consts"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var RegisterValidators = []gin.HandlerFunc{
	gv.NewBody("username", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Username must be between %d and %d characters",
			consts.MIN_USERNAME_LEN,
			consts.MAX_USERNAME_LEN,
		)
	},
	).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Length(&vgo.IsLengthOpts{
		Min: consts.MIN_USERNAME_LEN,
		Max: &consts.MAX_USERNAME_LEN,
	}).Validate(),

	gv.NewBody("email", func(value, _, _ string) string {
		return fmt.Sprintf("Invalid email address: \"%s\"", value)
	}).Chain().Email(&vgo.IsEmailOpts{}).Validate(),

	gv.NewBody("password", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Password must be at least %d characters long",
			consts.MIN_PASSWORD_LEN,
		)
	}).Chain().Length(&vgo.IsLengthOpts{Min: consts.MIN_PASSWORD_LEN}).Validate(),
	middleware.ValidationMiddleware,
}

var VerifyEmailValidators = []gin.HandlerFunc{
	gv.NewBody("code", func(_, _, _ string) string {
		return "Invalid verification code"
	}).Chain().Numeric(&vgo.IsNumericOpts{}).Validate(),
	middleware.ValidationMiddleware,
}

var LoginValidators = []gin.HandlerFunc{
	gv.NewBody("email", func(value, _, _ string) string {
		return fmt.Sprintf("Invalid email address: \"%s\"", value)
	}).Chain().Email(&vgo.IsEmailOpts{}).Validate(),

	gv.NewBody("password", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Password must be at least %d characters long",
			consts.MIN_PASSWORD_LEN,
		)
	}).Chain().Length(&vgo.IsLengthOpts{Min: consts.MIN_PASSWORD_LEN}).Validate(),
	middleware.ValidationMiddleware,
}
