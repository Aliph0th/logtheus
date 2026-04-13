package models

import (
	"encoding/json"
	"time"
)

type LogRecord struct {
	LogID           string          `json:"log_id" gorm:"not null;size:64"`
	ApplicationID   uint64          `json:"application_id" gorm:"not null"`
	ApplicationName string          `json:"application_name" gorm:"not null;size:255"`
	ProjectID       uint64          `json:"project_id" gorm:"not null"`
	Format          string          `json:"format" gorm:"not null;type:LowCardinality(String);size:20"`
	SourceIP        string          `json:"source_ip" gorm:"not null;type:IPv4"`
	ReceivedAt      time.Time       `json:"received_at" gorm:"not null"`
	Attributes      json.RawMessage `json:"attributes" gorm:"type:JSON;not null"`
	S3Key           string          `json:"s3_key" gorm:"not null;"`
}

func (LogRecord) TableName() string {
	return "logs"
}
