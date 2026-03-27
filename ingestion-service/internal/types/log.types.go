package types

import (
	"logtheus/ingestion/internal/consts"
	"time"
)

type NormalizedLog struct {
	APIKey       string
	KeySignature string
	Format       consts.LogFormat
	SourceIP     string
	ReceivedAt   time.Time
	EventTime    *time.Time
	Message      string
	Attributes   map[string]string
	RawBase64    string
}
