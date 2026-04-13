package repository

import (
	"context"
	"logtheus/logengine/internal/models"
	"logtheus/shared/pkg/storages"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
