package services

import (
	"context"
	"log/slog"
	"logtheus/logengine/internal/consts"
	"logtheus/logengine/internal/repository"
	"time"
)

type ClusteringCleanupService struct {
	clusteringRepo *repository.ClusteringJobRepository
}

func NewClusteringCleanupService(clusteringRepo *repository.ClusteringJobRepository) *ClusteringCleanupService {
	return &ClusteringCleanupService{clusteringRepo: clusteringRepo}
}

func (s *ClusteringCleanupService) Start(ctx context.Context) {
	slog.Info("Starting clustering cleanup service")
	go s.run(ctx)
}

func (s *ClusteringCleanupService) run(ctx context.Context) {
	ticker := time.NewTicker(consts.CLUSTERING_CLEANUP_INTERVAL)
	defer ticker.Stop()

	s.cleanupOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupOnce(ctx)
		}
	}
}

func (s *ClusteringCleanupService) cleanupOnce(ctx context.Context) {
	deleted, err := s.clusteringRepo.PurgeExpired(ctx, time.Now().UTC())
	if err != nil {
		slog.Error("Failed to purge expired clustering jobs", "error", err)
		return
	}
	if deleted > 0 {
		slog.Info("Purged expired clustering jobs", "deleted_jobs", deleted)
	}
}
