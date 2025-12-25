package validators

import (
	"fmt"
	"logtheus/internal/api/middleware"
	"logtheus/internal/constants"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var RegisterValidators = []gin.HandlerFunc{
	gv.NewBody("username", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Username must be between %d and %d characters",
			constants.MIN_USERNAME_LEN,
			constants.MAX_USERNAME_LEN,
		)
	},
	).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Length(&vgo.IsLengthOpts{
		Min: constants.MIN_USERNAME_LEN,
		Max: &constants.MAX_USERNAME_LEN,
	}).Validate(),

	gv.NewBody("email", func(value, _, _ string) string {
		return fmt.Sprintf("Invalid email address: \"%s\"", value)
	}).Chain().Email(&vgo.IsEmailOpts{}).Validate(),

	gv.NewBody("password", func(_, _, _ string) string {
		return fmt.Sprintf(
			"Password must be at least %d characters long",
			constants.MIN_PASSWORD_LEN,
		)
	}).Chain().Length(&vgo.IsLengthOpts{Min: constants.MIN_PASSWORD_LEN}).Validate(),
	middleware.ValidationMiddleware,
}
