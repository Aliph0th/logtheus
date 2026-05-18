package validators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"logtheus/gateway/internal/api/dto"
	excepts "logtheus/gateway/internal/api/exceptions"
	"logtheus/gateway/internal/api/middleware"
	"logtheus/gateway/internal/consts"
	"logtheus/gateway/internal/utils"
	sharedConsts "logtheus/shared/pkg/consts"
	"net"
	"net/http"
	"strconv"
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
			value, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				return false
			}
			return value > 0
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
			value, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				return false
			}
			return value > 0 && value <= 100
		},
	).Validate(),
}

var LogClusteringStartBodyValidators = []gin.HandlerFunc{
	gv.NewBody("project_id", func(_, _, _ string) string {
		return "project_id must be a positive integer"
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value, err := strconv.ParseUint(sanitizedValue, 10, 64)
			if err != nil {
				return false
			}
			return value > 0
		},
	).Validate(),

	gv.NewBody("application_id", func(_, _, _ string) string {
		return "application_id must be a positive integer"
	}).Chain().Optional().Numeric(&vgo.IsNumericOpts{NoSymbols: true}).Bail().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			raw := strings.TrimSpace(sanitizedValue)
			if raw == "" {
				return true
			}

			value, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				return false
			}

			return value > 0
		},
	).Validate(),

	gv.NewBody("from", func(_, _, validatorName string) string {
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

	gv.NewBody("to", func(_, _, validatorName string) string {
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

	gv.NewBody("cluster_by", func(_, _, _ string) string {
		return "cluster_by must be 'embedding' or one of canonical fields"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.ToLower(strings.TrimSpace(sanitizedValue))
			if value == "" || value == "embedding" {
				return true
			}
			_, exists := sharedConsts.CANONICAL_FIELDS_AGGREGATION_MAP[value]
			return exists
		},
	).Validate(),

	gv.NewBody("eps", func(_, _, _ string) string {
		return "eps must be a float in range (0,2]"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return false
			}
			return parsed > 0 && parsed <= 2
		},
	).Validate(),

	gv.NewBody("min_points", func(_, _, _ string) string {
		return "min_points must be an integer >= 2"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return false
			}
			return parsed >= 2
		},
	).Validate(),

	gv.NewBody("max_points", func(_, _, _ string) string {
		return "max_points must be an integer in range [2,20000]"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return false
			}
			return parsed >= 2 && parsed <= 20000
		},
	).Validate(),

	gv.NewBody("ttl_hours", func(_, _, _ string) string {
		return "ttl_hours must be a positive integer"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return false
			}
			return parsed > 0
		},
	).Validate(),

	validateClusteringRangeBody,
}

var LogClusteringJobIDValidators = []gin.HandlerFunc{
	gv.NewParam("job_id", func(_, _, validatorName string) string {
		switch validatorName {
		case "IsEmpty":
			return "job_id is required"
		default:
			return "job_id must be a valid UUID"
		}
	}).Chain().Not().Empty(&vgo.IsEmptyOpts{IgnoreWhitespace: true}).Bail().UUID("4").Bail().Validate(),
}

var LogClusteringResultValidators = []gin.HandlerFunc{
	gv.NewQuery("offset", func(_, _, _ string) string {
		return "offset must be a non-negative integer"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			_, err := strconv.ParseUint(value, 10, 32)
			return err == nil
		},
	).Validate(),

	gv.NewQuery("limit", func(_, _, _ string) string {
		return "limit must be a positive integer <= 1000"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return false
			}
			return parsed > 0 && parsed <= 1000
		},
	).Validate(),
}

var LogClusteringJobsValidators = []gin.HandlerFunc{
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
			value, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				return false
			}
			return value > 0
		},
	).Validate(),

	gv.NewQuery("offset", func(_, _, _ string) string {
		return "offset must be a non-negative integer"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			_, err := strconv.ParseUint(value, 10, 32)
			return err == nil
		},
	).Validate(),

	gv.NewQuery("limit", func(_, _, _ string) string {
		return "limit must be a positive integer <= 1000"
	}).Chain().Optional().CustomValidator(
		func(_ *http.Request, _, sanitizedValue string) bool {
			value := strings.TrimSpace(sanitizedValue)
			if value == "" {
				return true
			}
			parsed, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return false
			}
			return parsed > 0 && parsed <= 1000
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

func validateClusteringRangeBody(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.Next()
		return
	}

	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

	req := &dto.LogClusteringStartRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		ctx.Next()
		return
	}

	fromRaw := strings.TrimSpace(req.From)
	toRaw := strings.TrimSpace(req.To)
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
