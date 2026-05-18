package repository

import (
	"context"
	"errors"
	"logtheus/logengine/internal/consts"
	"logtheus/logengine/internal/models"
	"logtheus/shared/pkg/storages"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClusteringJobRepository struct {
	db *gorm.DB
}

type ClusterRepresentativeRow struct {
	ClusterID int32  `gorm:"column:cluster_id"`
	LogID     string `gorm:"column:log_id"`
}

func NewClusteringJobRepository(db *storages.Database) *ClusteringJobRepository {
	return &ClusteringJobRepository{db: db.DB}
}

func dedupeAssignmentsByLogID(assignments []*models.ClusteringAssignment) []*models.ClusteringAssignment {
	if len(assignments) <= 1 {
		return assignments
	}

	unique := make([]*models.ClusteringAssignment, 0, len(assignments))
	seen := make(map[string]int, len(assignments))
	for _, assignment := range assignments {
		if index, exists := seen[assignment.LogID]; exists {
			unique[index] = assignment
			continue
		}
		seen[assignment.LogID] = len(unique)
		unique = append(unique, assignment)
	}

	return unique
}

func (r *ClusteringJobRepository) Create(ctx context.Context, job *models.ClusteringJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *ClusteringJobRepository) FindActiveByHash(ctx context.Context, requestHash string) (*models.ClusteringJob, error) {
	var job models.ClusteringJob
	err := r.db.WithContext(ctx).
		Where("request_hash = ?", requestHash).
		Where("status IN ?", []consts.ClusteringJobStatus{consts.CLUSTERING_STATUS_QUEUED, consts.CLUSTERING_STATUS_RUNNING}).
		Order("created_at DESC").
		First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ClusteringJobRepository) GetByID(ctx context.Context, jobID uuid.UUID) (*models.ClusteringJob, error) {
	var job models.ClusteringJob
	err := r.db.WithContext(ctx).Where("job_id = ?", jobID).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *ClusteringJobRepository) ListJobs(
	ctx context.Context,
	projectID uint64,
	applicationID *uint64,
	offset uint32,
	limit uint32,
) ([]*models.ClusteringJob, uint32, error) {
	base := r.db.WithContext(ctx).Model(&models.ClusteringJob{}).Where("project_id = ?", projectID)
	if applicationID != nil {
		base = base.Where("application_id = ?", *applicationID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]*models.ClusteringJob, 0)
	query := base.Order("created_at DESC").Offset(int(offset)).Limit(int(limit))
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, uint32(total), nil
}

func (r *ClusteringJobRepository) MarkRunning(ctx context.Context, jobID uuid.UUID, startedAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.ClusteringJob{}).
		Where("job_id = ?", jobID).
		Where("status = ?", consts.CLUSTERING_STATUS_QUEUED).
		Updates(map[string]any{
			"status":     consts.CLUSTERING_STATUS_RUNNING,
			"started_at": startedAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ClusteringJobRepository) MarkCanceled(ctx context.Context, jobID uuid.UUID, now time.Time) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.ClusteringJob{}).
		Where("job_id = ?", jobID).
		Where("status IN ?", []consts.ClusteringJobStatus{consts.CLUSTERING_STATUS_QUEUED, consts.CLUSTERING_STATUS_RUNNING}).
		Updates(map[string]any{
			"status":      consts.CLUSTERING_STATUS_CANCELED,
			"finished_at": now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ClusteringJobRepository) UpdateProgress(ctx context.Context, jobID uuid.UUID, progress float64, totalPoints uint32) error {
	return r.db.WithContext(ctx).
		Model(&models.ClusteringJob{}).
		Where("job_id = ?", jobID).
		Where("status = ?", consts.CLUSTERING_STATUS_RUNNING).
		Updates(map[string]any{
			"progress_percent": progress,
			"total_points":     totalPoints,
		}).Error
}

func (r *ClusteringJobRepository) MarkFailed(ctx context.Context, jobID uuid.UUID, errorMessage string, now time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.ClusteringJob{}).
		Where("job_id = ?", jobID).
		Where("status = ?", consts.CLUSTERING_STATUS_RUNNING).
		Updates(map[string]any{
			"status":           consts.CLUSTERING_STATUS_FAILED,
			"error_message":    errorMessage,
			"progress_percent": 100.0,
			"finished_at":      now,
		}).Error
}

func (r *ClusteringJobRepository) SaveSucceededResult(
	ctx context.Context,
	jobID uuid.UUID,
	assignments []*models.ClusteringAssignment,
	summaries []*models.ClusteringClusterSummary,
	clusterCount uint32,
	noiseCount uint32,
	now time.Time,
) error {
	assignments = dedupeAssignmentsByLogID(assignments)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", jobID).Delete(&models.ClusteringAssignment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("job_id = ?", jobID).Delete(&models.ClusteringClusterSummary{}).Error; err != nil {
			return err
		}

		if len(assignments) > 0 {
			if err := tx.Create(assignments).Error; err != nil {
				return err
			}
		}
		if len(summaries) > 0 {
			if err := tx.Create(summaries).Error; err != nil {
				return err
			}
		}

		return tx.Model(&models.ClusteringJob{}).
			Where("job_id = ?", jobID).
			Where("status = ?", consts.CLUSTERING_STATUS_RUNNING).
			Updates(map[string]any{
				"status":           consts.CLUSTERING_STATUS_SUCCEEDED,
				"progress_percent": 100.0,
				"cluster_count":    clusterCount,
				"noise_count":      noiseCount,
				"finished_at":      now,
			}).Error
	})
}

func (r *ClusteringJobRepository) GetAssignmentsPaged(ctx context.Context, jobID uuid.UUID, offset uint32, limit uint32) ([]models.ClusteringAssignment, uint32, error) {
	query := r.db.WithContext(ctx).Model(&models.ClusteringAssignment{}).Where("job_id = ?", jobID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]models.ClusteringAssignment, 0)
	if err := query.Order("log_id ASC").Offset(int(offset)).Limit(int(limit)).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, uint32(total), nil
}

func (r *ClusteringJobRepository) GetClusterSummaries(ctx context.Context, jobID uuid.UUID) ([]models.ClusteringClusterSummary, error) {
	rows := make([]models.ClusteringClusterSummary, 0)
	err := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("size DESC, cluster_id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ClusteringJobRepository) GetClusterRepresentativeLogIDs(ctx context.Context, jobID uuid.UUID) (map[int32]string, error) {
	rows := make([]ClusterRepresentativeRow, 0)
	err := r.db.WithContext(ctx).
		Raw(
			`SELECT cluster_id, MIN(log_id) AS log_id
			 FROM clustering_assignments
			 WHERE job_id = ? AND cluster_id >= 0
			 GROUP BY cluster_id`,
			jobID,
		).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int32]string, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.LogID) == "" {
			continue
		}
		result[row.ClusterID] = row.LogID
	}

	return result, nil
}

func (r *ClusteringJobRepository) PurgeExpired(ctx context.Context, now time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", now).
		Delete(&models.ClusteringJob{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
