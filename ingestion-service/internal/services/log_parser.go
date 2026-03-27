package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"logtheus/ingestion/internal/consts"
	"logtheus/ingestion/internal/types"
	"regexp"
	"unicode/utf8"
)

var nginxCombinedLogRegex = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "([^"]*)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)"`)

func ParseIntoNormalized(raw []byte, format consts.LogFormat, normalized *types.NormalizedLog) error {
	switch format {
	case consts.LOG_FORMAT_JSON:
		return parseJSON(raw, normalized)
	case consts.LOG_FORMAT_NGINX:
		return parseNginx(raw, normalized)
	case consts.LOG_FORMAT_TEXT:
		normalized.Message = string(raw)
		return nil
	case consts.LOG_FORMAT_BINARY:
		normalized.RawBase64 = base64.StdEncoding.EncodeToString(raw)
		return nil
	default:
		return fmt.Errorf("unsupported log format: %s", format)
	}
}

func DetectFormat(raw []byte) consts.LogFormat {
	if json.Valid(raw) {
		return consts.LOG_FORMAT_JSON
	}

	if utf8.Valid(raw) {
		if nginxCombinedLogRegex.MatchString(string(raw)) {
			return consts.LOG_FORMAT_NGINX
		}
		return consts.LOG_FORMAT_TEXT
	}

	return consts.LOG_FORMAT_BINARY
}

func parseJSON(raw []byte, normalized *types.NormalizedLog) error {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	attributes := make(map[string]string, len(payload))
	for key, value := range payload {
		attributes[key] = fmt.Sprint(value)
	}

	normalized.Attributes = attributes
	if msg, ok := extractFirstString(payload, "message", "msg", "log"); ok {
		normalized.Message = msg
	}
	return nil
}

func parseNginx(raw []byte, normalized *types.NormalizedLog) error {
	line := string(raw)
	matches := nginxCombinedLogRegex.FindStringSubmatch(line)
	if len(matches) != 8 {
		normalized.Message = line
		return nil
	}

	normalized.Message = matches[3]
	normalized.Attributes = map[string]string{
		"remote_addr":     matches[1],
		"time_local":      matches[2],
		"request":         matches[3],
		"status":          matches[4],
		"body_bytes_sent": matches[5],
		"http_referer":    matches[6],
	}
	return nil
}

func extractFirstString(payload map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			return fmt.Sprint(value), true
		}
	}
	return "", false
}
