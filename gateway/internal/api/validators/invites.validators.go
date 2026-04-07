package validators

import (
	"fmt"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/shared/pkg/consts"
	"net/http"
	"time"

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

	gv.NewBody("expires_at", func(_, _, validatorName string) string {
		if validatorName == "CustomValidator" {
			return "Expiration date cannot be in the past"
		}
		return "Expiration date must be a valid ISO 8601 date string"
	}).Chain().Optional().ISO8601(&vgo.IsISO8601Opts{}).CustomValidator(func(r *http.Request, initialValue, sanitizedValue string) bool {
		timeValue, err := time.Parse(time.RFC3339, sanitizedValue)
		if err != nil {
			return false
		}
		return !timeValue.Before(time.Now())
	}).Validate(),

	middleware.ValidationMiddleware,
}
