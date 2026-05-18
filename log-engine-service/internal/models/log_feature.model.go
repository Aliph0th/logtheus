package models

import (
	"encoding/json"
	"time"

	"github.com/pgvector/pgvector-go"
)

type LogFeature struct {
	LogID         string          `gorm:"primaryKey;size:64" json:"log_id"`
	ApplicationID uint64          `gorm:"not null" json:"application_id"`
	ProjectID     uint64          `gorm:"not null" json:"project_id"`
	Embedding     pgvector.Vector `gorm:"type:vector(384);not null" json:"embedding"`
	SimilarCount  uint64          `gorm:"not null;default:0" json:"similar_count"`
	Attributes    json.RawMessage `gorm:"type:JSONB;not null" json:"attributes"`
	CreatedAt     time.Time       `gorm:"not null" json:"created_at"`
}

func (LogFeature) TableName() string {
	return "log_features"
}
