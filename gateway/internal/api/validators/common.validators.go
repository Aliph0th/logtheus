package validators

import (
	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var DatabaseID = func(field string) gin.HandlerFunc {
	return gv.NewParam(field, func(_, _, _ string) string {
		return "ID must be numeric"
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Validate()
}
