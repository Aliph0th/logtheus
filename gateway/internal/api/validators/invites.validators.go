package validators

import (
	"fmt"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/shared/pkg/consts"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var availableRoles = []string{
	string(consts.PROJECT_ROLE_MEMBER),
	string(consts.PROJECT_ROLE_VIEWER),
}

var CreateInviteValidators = []gin.HandlerFunc{
	gv.NewBody("project_id", func(_, _, _ string) string {
		return "Project ID must be a valid number"
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Validate(),

	gv.NewBody("email", func(_, _, _ string) string {
		return "Invalid email address"
	}).Chain().Email(&vgo.IsEmailOpts{}).Validate(),

	gv.NewBody("role", func(_, _, _ string) string {
		return fmt.Sprintf("Role must be one of the following: %v", availableRoles)
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().In(availableRoles).Validate(),

	gv.NewBody("expires_at", func(_, _, _ string) string {
		return "Expiration date must be a valid ISO 8601 date string"
	}).Chain().Optional().ISO8601(&vgo.IsISO8601Opts{}).Validate(),

	middleware.ValidationMiddleware,
}
