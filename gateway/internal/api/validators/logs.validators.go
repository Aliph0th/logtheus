package validators

import (
	"encoding/json"
	"fmt"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/consts"
	"net/http"
	"strings"

	gv "github.com/bube054/ginvalidator"
	vgo "github.com/bube054/validatorgo"
	"github.com/gin-gonic/gin"
)

var IngestLogsValidators = []gin.HandlerFunc{
	gv.NewBody("logs", func(_, _, validatorName string) string {
		switch validatorName {
		case "IsEmpty":
			return "logs is required"
		case "CustomValidator":
			return "logs must be a non-empty array of strings"
		default:
			return "logs must be a valid array"
		}
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			var logs []string
			if err := json.Unmarshal([]byte(sanitizedValue), &logs); err != nil {
				return false
			}

			if len(logs) == 0 {
				return false
			}

			for _, rawData := range logs {
				if strings.TrimSpace(rawData) == "" {
					return false
				}
			}

			return true
		},
	).Validate(),

	gv.NewBody("logs", func(_, _, _ string) string {
		return fmt.Sprintf("Total logs payload exceeds %d bytes", consts.MAX_INGESTION_BYTES)
	}).Chain().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			var logs []string
			if err := json.Unmarshal([]byte(sanitizedValue), &logs); err != nil {
				return true
			}

			totalPayloadSize := 0
			for _, rawData := range logs {
				totalPayloadSize += len(rawData)
			}

			return totalPayloadSize <= consts.MAX_INGESTION_BYTES
		},
	).Validate(),

	middleware.ValidationMiddleware,
}
