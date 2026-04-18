package validators

import (
	"encoding/json"
	"fmt"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/consts"
	"logtheus/gateway/internal/utils"
	sharedConsts "logtheus/shared/pkg/consts"
	"net"
	"net/http"
	"strings"
	"time"

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

var LogMetricsCommonValidators = []gin.HandlerFunc{
	gv.NewQuery("project_id", func(_, _, _ string) string {
		return "project_id must be a positive integer"
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			raw := strings.TrimSpace(sanitizedValue)
			if raw == "" {
				return false
			}
			value := 0
			for _, ch := range raw {
				value = value*10 + int(ch-'0')
				if value > 0 {
					return true
				}
			}
			return false
		},
	).Validate(),

	gv.NewQuery("application_id", func(_, _, _ string) string {
		return "application_id must be a positive integer"
	}).Chain().Optional().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			raw := strings.TrimSpace(sanitizedValue)
			if raw == "" {
				return true
			}
			value := 0
			for _, ch := range raw {
				value = value*10 + int(ch-'0')
				if value > 0 {
					return true
				}
			}
			return false
		},
	).Validate(),

	gv.NewQuery("formats", func(_, _, _ string) string {
		return "formats must contain only: FORMAT_JSON, FORMAT_TEXT, FORMAT_BINARY"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			values := utils.SplitCSV(sanitizedValue)
			if len(values) == 0 {
				return true
			}

			for _, value := range values {
				switch strings.ToUpper(strings.TrimSpace(value)) {
				case "FORMAT_JSON", "FORMAT_TEXT", "FORMAT_BINARY":
					continue
				default:
					return false
				}
			}

			return true
		},
	).Validate(),

	gv.NewQuery("source_ips", func(_, _, _ string) string {
		return "source_ips must be a comma-separated list of valid IP addresses"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			values := utils.SplitCSV(sanitizedValue)
			if len(values) == 0 {
				return true
			}

			for _, value := range values {
				if net.ParseIP(strings.TrimSpace(value)) == nil {
					return false
				}
			}

			return true
		},
	).Validate(),

	gv.NewQuery("from", func(_, _, validatorName string) string {
		switch validatorName {
		case "IsEmpty":
			return "from is required"
		default:
			return "from must be RFC3339"
		}
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			_, err := time.Parse(time.RFC3339, strings.TrimSpace(sanitizedValue))
			return err == nil
		},
	).Validate(),

	gv.NewQuery("to", func(_, _, validatorName string) string {
		switch validatorName {
		case "IsEmpty":
			return "to is required"
		default:
			return "to must be RFC3339"
		}
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			_, err := time.Parse(time.RFC3339, strings.TrimSpace(sanitizedValue))
			return err == nil
		},
	).Validate(),

	validateLogMetricsRangeQuery,
}

var LogMetricsVolumeValidators = []gin.HandlerFunc{
	gv.NewQuery("bucket", func(_, _, _ string) string {
		return "bucket must be one of: 1m, 5m, 1h, 5h, 10h, 24h"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			bucket := strings.ToLower(strings.TrimSpace(sanitizedValue))
			if bucket == "" {
				return true
			}
			switch bucket {
			case "1m", "5m", "1h", "5h", "10h", "24h":
				return true
			default:
				return false
			}
		},
	).Validate(),
}

var LogMetricsAggregationValidators = []gin.HandlerFunc{
	gv.NewQuery("field", func(_, _, validatorName string) string {
		switch validatorName {
		case "IsEmpty":
			return "field is required"
		default:
			return "field must be one of canonical fields: service, level, timestamp, environment, event, error_message, status_code, duration, ip, method, path, useragent, hostname"
		}
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			_, exists := sharedConsts.CANONICAL_FIELDS_AGGREGATION_MAP[strings.TrimSpace(sanitizedValue)]
			return exists
		},
	).Validate(),

	gv.NewQuery("limit", func(_, _, _ string) string {
		return "limit must be a positive integer <= 100"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			raw := strings.TrimSpace(sanitizedValue)
			if raw == "" {
				return true
			}
			for _, ch := range raw {
				if ch < '0' || ch > '9' {
					return false
				}
			}
			value := 0
			for _, ch := range raw {
				value = value*10 + int(ch-'0')
				if value > 100 {
					return false
				}
			}
			return value > 0
		},
	).Validate(),
}

func validateLogMetricsRangeQuery(ctx *gin.Context) {
	fromRaw := strings.TrimSpace(ctx.Query("from"))
	toRaw := strings.TrimSpace(ctx.Query("to"))
	from, fromErr := time.Parse(time.RFC3339, fromRaw)
	to, toErr := time.Parse(time.RFC3339, toRaw)
	if fromErr != nil || toErr != nil {
		ctx.Next()
		return
	}

	if !from.Before(to) {
		excepts.RespondError(ctx, excepts.WithBadRequest("from must be before to"))
		return
	}

	ctx.Next()
}
