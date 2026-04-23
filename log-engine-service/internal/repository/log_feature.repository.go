package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"logtheus/logengine/internal/models"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"logtheus/shared/pkg/storages"
	"time"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClusteringFeatureRow struct {
	LogID      string          `gorm:"column:log_id"`
	Embedding  pgvector.Vector `gorm:"column:embedding"`
	Attributes json.RawMessage `gorm:"column:attributes"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
}

type LogFeatureRepository struct {
	db *gorm.DB
}

func NewLogFeatureRepository(db *storages.Database) *LogFeatureRepository {
	return &LogFeatureRepository{db: db.DB}
}

func (r *LogFeatureRepository) UpsertBatch(ctx context.Context, features []*models.LogFeature) error {
	if len(features) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "log_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"application_id", "project_id", "embedding", "attributes"}),
		}).
		Create(features).Error
}

func (r *LogFeatureRepository) ListByFilterWithLimit(
	ctx context.Context,
	filter *logEngineProto.ClusteringFilter,
	maxPoints uint32,
) ([]ClusteringFeatureRow, bool, error) {
	if filter == nil {
		return nil, false, fmt.Errorf("filter is required")
	}
	if filter.From == nil || filter.To == nil {
		return nil, false, fmt.Errorf("from and to are required")
	}

	limit := int(maxPoints)
	if limit <= 0 {
		limit = 1
	}

	query := r.db.WithContext(ctx).
		Model(&models.LogFeature{}).
		Select("log_id, embedding, attributes, created_at").
		Where("project_id = ?", filter.ProjectId).
		Where("created_at >= ?", filter.From.AsTime().UTC()).
		Where("created_at <= ?", filter.To.AsTime().UTC()).
		Order("created_at ASC")

	if filter.ApplicationId != nil {
		query = query.Where("application_id = ?", filter.GetApplicationId())
	}

	rows := make([]ClusteringFeatureRow, 0)
	if err := query.Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, false, err
	}

	truncated := len(rows) > limit
	if truncated {
		rows = rows[:limit]
	}

	return rows, truncated, nil
}
