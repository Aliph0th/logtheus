package utils

import (
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"strings"
)

func SplitCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			values = append(values, item)
		}
	}

	if len(values) == 0 {
		return nil
	}
	return values
}

func MapTimeBucket(raw string) logEngineProto.TimeBucket {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "5m":
		return logEngineProto.TimeBucket_TIME_BUCKET_5M
	case "1h":
		return logEngineProto.TimeBucket_TIME_BUCKET_1H
	case "5h":
		return logEngineProto.TimeBucket_TIME_BUCKET_5H
	case "10h":
		return logEngineProto.TimeBucket_TIME_BUCKET_10H
	case "24h":
		return logEngineProto.TimeBucket_TIME_BUCKET_24H
	default:
		return logEngineProto.TimeBucket_TIME_BUCKET_1M
	}
}
