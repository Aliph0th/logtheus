package models

import (
	"logtheus/logengine/internal/consts"
	"time"

	"github.com/google/uuid"
)

type ClusteringJob struct {
	JobID           uuid.UUID                  `gorm:"primaryKey;type:uuid"`
	RequestHash     string                     `gorm:"size:64;index:idx_clustering_jobs_request_hash_status"`
	ProjectID       uint64                     `gorm:"not null;index"`
	ApplicationID   *uint64                    `gorm:"index"`
	FromTime        time.Time                  `gorm:"column:from_time;not null;index"`
	ToTime          time.Time                  `gorm:"column:to_time;not null;index"`
	ClusterBy       string                     `gorm:"size:64;not null;default:embedding"`
	Eps             float64                    `gorm:"not null"`
	MinPoints       uint32                     `gorm:"not null"`
	MaxPoints       uint32                     `gorm:"not null"`
	Status          consts.ClusteringJobStatus `gorm:"size:16;not null;index:idx_clustering_jobs_request_hash_status"`
	ProgressPercent float64                    `gorm:"not null;default:0"`
	TotalPoints     uint32                     `gorm:"not null;default:0"`
	ClusterCount    uint32                     `gorm:"not null;default:0"`
	NoiseCount      uint32                     `gorm:"not null;default:0"`
	ErrorMessage    string                     `gorm:"type:text"`
	ExpiresAt       time.Time                  `gorm:"not null;index"`
	CreatedAt       time.Time                  `gorm:"not null"`
	UpdatedAt       time.Time                  `gorm:"not null;u"`
	StartedAt       *time.Time
	FinishedAt      *time.Time
}

func (ClusteringJob) TableName() string {
	return "clustering_jobs"
}
