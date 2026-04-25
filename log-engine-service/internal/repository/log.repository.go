package repository

import (
	"context"
	"fmt"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/storages"
	"logtheus/logengine/internal/utils"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"time"
)

type VolumePoint struct {
	Bucket time.Time `gorm:"column:bucket"`
	Count  uint64    `gorm:"column:count"`
}

type AggregationItem struct {
	Value string `gorm:"column:value"`
	Count uint64 `gorm:"column:count"`
}

type LatencyStats struct {
	P50 float64 `gorm:"column:p50"`
	P95 float64 `gorm:"column:p95"`
	P99 float64 `gorm:"column:p99"`
}

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

func (r *LogRepository) SaveBatchInChunks(ctx context.Context, logs []*models.LogRecord, chunkSize int) error {
	if len(logs) == 0 {
		return nil
	}

	if chunkSize <= 0 {
		chunkSize = len(logs)
	}

	if err := r.ch.DB.WithContext(ctx).CreateInBatches(logs, chunkSize).Error; err != nil {
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

func (r *LogRepository) GetVolumeSeries(ctx context.Context, filter *logEngineProto.LogAggregationFilter, bucket logEngineProto.TimeBucket) ([]VolumePoint, error) {
	rows := make([]VolumePoint, 0)
	err := utils.ApplyAggregationFilter(r.ch.DB.WithContext(ctx).Table(models.LogRecord{}.TableName()), filter).
		Select(fmt.Sprintf("%s AS bucket, count() AS count", utils.BucketExpression(bucket))).
		Group("bucket").
		Order("bucket ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *LogRepository) GetAggregationByField(ctx context.Context, filter *logEngineProto.LogAggregationFilter, field string, limit uint32) ([]AggregationItem, error) {
	fieldExpr, ok := utils.AggregationFieldExpression(field)
	if !ok {
		return nil, fmt.Errorf("unsupported aggregation field: %s", field)
	}

	rows := make([]AggregationItem, 0)
	err := utils.ApplyAggregationFilter(r.ch.DB.WithContext(ctx).Table(models.LogRecord{}.TableName()), filter).
		Select(fmt.Sprintf("ifNull(nullIf(%s, ''), 'unknown') AS value, count() AS count", fieldExpr)).
		Group("value").
		Order("count DESC").
		Limit(int(limit)).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *LogRepository) GetLatencyStats(ctx context.Context, filter *logEngineProto.LogAggregationFilter) (LatencyStats, error) {
	stats := LatencyStats{}
	err := utils.ApplyAggregationFilter(r.ch.DB.WithContext(ctx).Table(models.LogRecord{}.TableName()), filter).
		Select("quantileTDigest(0.50)(toFloat64OrZero(toString(attributes.duration))) AS p50, quantileTDigest(0.95)(toFloat64OrZero(toString(attributes.duration))) AS p95, quantileTDigest(0.99)(toFloat64OrZero(toString(attributes.duration))) AS p99").
		Scan(&stats).Error
	if err != nil {
		return LatencyStats{}, err
	}

	return stats, nil
}
