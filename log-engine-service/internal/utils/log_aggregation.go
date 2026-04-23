package utils

import (
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"strings"

	"gorm.io/gorm"
)

var canonicalAggregationFieldExpressions = map[string]string{
	"service":       "toString(attributes.service)",
	"level":         "lowerUTF8(toString(attributes.level))",
	"timestamp":     "toString(attributes.timestamp)",
	"environment":   "toString(attributes.environment)",
	"event":         "toString(attributes.event)",
	"error_message": "toString(attributes.error_message)",
	"status_code":   "toString(attributes.status_code)",
	"duration":      "toString(attributes.duration)",
	"ip":            "toString(attributes.ip)",
	"method":        "upperUTF8(toString(attributes.method))",
	"path":          "toString(attributes.path)",
	"useragent":     "toString(attributes.useragent)",
	"hostname":      "toString(attributes.hostname)",
}

var aggregationFieldAliases = map[string]string{
	"service":       "service",
	"level":         "level",
	"timestamp":     "timestamp",
	"time":          "timestamp",
	"environment":   "environment",
	"env":           "environment",
	"event":         "event",
	"message":       "event",
	"error":         "error_message",
	"errormessage":  "error_message",
	"error_message": "error_message",
	"status":        "status_code",
	"statuscode":    "status_code",
	"status_code":   "status_code",
	"duration":      "duration",
	"latency":       "duration",
	"ip":            "ip",
	"method":        "method",
	"path":          "path",
	"useragent":     "useragent",
	"user_agent":    "useragent",
	"hostname":      "hostname",
}

func BucketExpression(bucket logEngineProto.TimeBucket) string {
	switch bucket {
	case logEngineProto.TimeBucket_TIME_BUCKET_5M:
		return "toStartOfInterval(received_at, INTERVAL 5 minute)"
	case logEngineProto.TimeBucket_TIME_BUCKET_1H:
		return "toStartOfInterval(received_at, INTERVAL 1 hour)"
	case logEngineProto.TimeBucket_TIME_BUCKET_5H:
		return "toStartOfInterval(received_at, INTERVAL 5 hour)"
	case logEngineProto.TimeBucket_TIME_BUCKET_10H:
		return "toStartOfInterval(received_at, INTERVAL 10 hour)"
	case logEngineProto.TimeBucket_TIME_BUCKET_24H:
		return "toStartOfInterval(received_at, INTERVAL 24 hour)"
	default:
		return "toStartOfInterval(received_at, INTERVAL 1 minute)"
	}
}

func ApplyAggregationFilter(tx *gorm.DB, filter *logEngineProto.LogAggregationFilter) *gorm.DB {
	query := tx.
		Where("project_id = ?", filter.ProjectId).
		Where("received_at >= ?", filter.GetFrom().AsTime().UTC()).
		Where("received_at < ?", filter.GetTo().AsTime().UTC())

	if filter.ApplicationId != nil {
		query = query.Where("application_id = ?", *filter.ApplicationId)
	}

	if formats := filter.GetFormats(); len(formats) > 0 {
		query = query.Where("format IN ?", formats)
	}

	if sourceIPs := filter.GetSourceIps(); len(sourceIPs) > 0 {
		query = query.Where("source_ip IN ?", sourceIPs)
	}

	return query
}

func NormalizeAggregationField(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, " ", "")

	if canonical, ok := aggregationFieldAliases[normalized]; ok {
		return canonical
	}

	return ""
}

func AggregationFieldExpression(raw string) (string, bool) {
	canonical := NormalizeAggregationField(raw)
	if canonical == "" {
		return "", false
	}

	expr, ok := canonicalAggregationFieldExpressions[canonical]
	return expr, ok
}
