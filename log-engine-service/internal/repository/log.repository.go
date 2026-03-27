package repository

import (
	"context"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/storages"
)

type LogRepository struct {
	ch *storages.ClickHouse
}

func NewLogRepository(ch *storages.ClickHouse) *LogRepository {
	return &LogRepository{ch: ch}
}

func (r *LogRepository) Save(ctx context.Context, log *models.LogRecord) error {
	if err := r.ch.DB.WithContext(ctx).Create(log).Error; err != nil {
		return err
	}
	return nil
}

func (r *LogRepository) SaveBatch(ctx context.Context, logs []*models.LogRecord) error {
	if len(logs) == 0 {
		return nil
	}

	if err := r.ch.DB.WithContext(ctx).Create(logs).Error; err != nil {
		return err
	}

	return nil
}

func (r *LogRepository) Delete(ctx context.Context, log *models.LogRecord) error {
	if err := r.ch.DB.WithContext(ctx).Delete(log).Error; err != nil {
		return err
	}
	return nil
}
