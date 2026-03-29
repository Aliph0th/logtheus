package utils

import "strings"

func SplitBrokers(brokers string) []string {
	parts := strings.Split(brokers, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		broker := strings.TrimSpace(item)
		if broker != "" {
			result = append(result, broker)
		}
	}
	if len(result) == 0 {
		panic("kafka brokers are not configured")
	}
	return result
}
