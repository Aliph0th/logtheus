package models

import (
	"time"

	"github.com/google/uuid"
)

type ClusteringAssignment struct {
	JobID     uuid.UUID     `gorm:"primaryKey;type:uuid"`
	Job       ClusteringJob `gorm:"foreignKey:JobID;references:JobID;constraint:OnDelete:CASCADE"`
	LogID     string        `gorm:"primaryKey;size:64"`
	ClusterID int32         `gorm:"not null;index"`
	CreatedAt time.Time     `gorm:"not null"`
}

func (ClusteringAssignment) TableName() string {
	return "clustering_assignments"
}

type ClusteringClusterSummary struct {
	JobID     uuid.UUID     `gorm:"primaryKey;type:uuid"`
	Job       ClusteringJob `gorm:"foreignKey:JobID;references:JobID;constraint:OnDelete:CASCADE"`
	ClusterID int32         `gorm:"primaryKey"`
	Size      uint32        `gorm:"not null"`
	CreatedAt time.Time     `gorm:"not null"`
}

func (ClusteringClusterSummary) TableName() string {
	return "clustering_cluster_summaries"
}
