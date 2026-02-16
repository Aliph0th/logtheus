package validators

import (
	"fmt"
	"logtheus/internal/api/middleware"
	"logtheus/internal/consts/enums"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var availableRoles = []string{
	string(enums.PROJECT_ROLE_MEMBER),
	string(enums.PROJECT_ROLE_VIEWER),
}

var CreateInviteValidators = []gin.HandlerFunc{
	gv.NewBody("projectID", func(_, _, _ string) string {
		return "Project ID must be a valid number"
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Validate(),

	gv.NewBody("email", func(_, _, _ string) string {
		return "Invalid email address"
	}).Chain().Email(&vgo.IsEmailOpts{}).Validate(),

	gv.NewBody("role", func(_, _, _ string) string {
		return fmt.Sprintf("Role must be one of the following: %v", availableRoles)
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().In(availableRoles).Validate(),

	middleware.ValidationMiddleware,
}
