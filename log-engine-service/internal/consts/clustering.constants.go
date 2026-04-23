package consts

import "time"

type ClusteringJobStatus string

const (
	CLUSTERING_STATUS_QUEUED    ClusteringJobStatus = "queued"
	CLUSTERING_STATUS_RUNNING   ClusteringJobStatus = "running"
	CLUSTERING_STATUS_SUCCEEDED ClusteringJobStatus = "succeeded"
	CLUSTERING_STATUS_FAILED    ClusteringJobStatus = "failed"
	CLUSTERING_STATUS_CANCELED  ClusteringJobStatus = "canceled"
	CLUSTERING_STATUS_EXPIRED   ClusteringJobStatus = "expired"
)

const (
	CLUSTERING_TARGET_EMBEDDING = "embedding"
)

const (
	DEFAULT_CLUSTERING_TTL_HOURS        uint32  = 24
	DEFAULT_CLUSTERING_EPS              float64 = 0.2
	DEFAULT_CLUSTERING_MIN_POINTS       uint32  = 5
	DEFAULT_CLUSTERING_MAX_POINTS       uint32  = 5000
	DEFAULT_CLUSTERING_RESULT_PAGE_SIZE uint32  = 200
	MAX_CLUSTERING_RESULT_PAGE_SIZE     uint32  = 1000
	DEFAULT_CLUSTERING_TARGET           string  = CLUSTERING_TARGET_EMBEDDING
)

const CLUSTERING_CLEANUP_INTERVAL = 1 * time.Hour
