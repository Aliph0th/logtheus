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

type SimilarLogMatch struct {
	LogID      string  `gorm:"column:log_id"`
	Similarity float32 `gorm:"column:similarity"`
}

func NewLogFeatureRepository(db *storages.Database) *LogFeatureRepository {
	return &LogFeatureRepository{db: db.DB}
}

func (r *LogFeatureRepository) UpsertBatch(ctx context.Context, features []*models.LogFeature) error {
	if len(features) == 0 {
		return nil
	}

	uniqueFeatures := dedupeFeaturesByLogID(features)

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "log_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"application_id", "project_id", "embedding", "attributes", "similar_count"}),
		}).
		Create(uniqueFeatures).Error
}

func (r *LogFeatureRepository) FindMostSimilarInWindow(
	ctx context.Context,
	projectID uint64,
	applicationID uint64,
	embedding pgvector.Vector,
	minSimilarity float32,
	fromTime time.Time,
) (*SimilarLogMatch, error) {
	row := &SimilarLogMatch{}

	err := r.db.WithContext(ctx).
		Raw(
			`SELECT log_id, (1 - (embedding <=> ?)) AS similarity
			 FROM log_features
			 WHERE project_id = ?
			   AND application_id = ?
			   AND created_at >= ?
			 ORDER BY embedding <=> ? ASC
			 LIMIT 1`,
			embedding, projectID, applicationID, fromTime, embedding,
		).
		Scan(row).Error
	if err != nil {
		return nil, err
	}

	if row.LogID == "" || row.Similarity < minSimilarity {
		return nil, nil
	}

	return row, nil
}

func (r *LogFeatureRepository) IncrementSimilarCounter(ctx context.Context, logID string, delta uint64) error {
	return r.db.WithContext(ctx).
		Model(&models.LogFeature{}).
		Where("log_id = ?", logID).
		UpdateColumn("similar_count", gorm.Expr("similar_count + ?", delta)).Error
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

func dedupeFeaturesByLogID(features []*models.LogFeature) []*models.LogFeature {
	if len(features) <= 1 {
		return features
	}

	unique := make([]*models.LogFeature, 0, len(features))
	seen := make(map[string]int, len(features))
	for _, feature := range features {
		if index, exists := seen[feature.LogID]; exists {
			unique[index] = feature
			continue
		}
		seen[feature.LogID] = len(unique)
		unique = append(unique, feature)
	}

	return unique
}
