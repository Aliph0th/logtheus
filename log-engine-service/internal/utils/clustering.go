package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"logtheus/logengine/internal/consts"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"math"
	"strings"
	"time"
)

type ClusteringParams struct {
	Eps       float64
	MinPoints uint32
	MaxPoints uint32
	TTLHours  uint32
}

func NormalizeClusteringParams(req *logEngineProto.StartClusteringJobRequest) *ClusteringParams {
	eps := req.GetParams().GetEps()
	if eps == 0 {
		eps = consts.DEFAULT_CLUSTERING_EPS
	}

	minPoints := req.GetParams().GetMinPoints()
	if minPoints == 0 {
		minPoints = consts.DEFAULT_CLUSTERING_MIN_POINTS
	}

	maxPoints := req.GetParams().GetMaxPoints()
	if maxPoints == 0 {
		maxPoints = consts.DEFAULT_CLUSTERING_MAX_POINTS
	}

	ttlHours := req.GetTtlHours()
	if ttlHours == 0 {
		ttlHours = consts.DEFAULT_CLUSTERING_TTL_HOURS
	}

	return &ClusteringParams{
		Eps:       eps,
		MinPoints: minPoints,
		MaxPoints: maxPoints,
		TTLHours:  ttlHours,
	}
}

func BuildClusteringRequestHash(req *logEngineProto.StartClusteringJobRequest, params *ClusteringParams) (string, error) {
	clusterBy := NormalizeClusterBy(req.GetClusterBy())
	normalized := map[string]any{
		"project_id":     req.GetFilter().GetProjectId(),
		"application_id": req.GetFilter().GetApplicationId(),
		"from":           req.GetFilter().GetFrom().AsTime().UTC().Format(time.RFC3339Nano),
		"to":             req.GetFilter().GetTo().AsTime().UTC().Format(time.RFC3339Nano),
		"cluster_by":     clusterBy,
		"eps":            params.Eps,
		"min_points":     params.MinPoints,
		"max_points":     params.MaxPoints,
		"request_key":    strings.TrimSpace(req.GetRequestKey()),
	}

	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:]), nil
}

func NormalizeClusterBy(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return consts.DEFAULT_CLUSTERING_TARGET
	}
	return value
}

func ClusteringStatusToProto(status consts.ClusteringJobStatus, expiresAt time.Time) logEngineProto.ClusteringJobStatus {
	if time.Now().UTC().After(expiresAt) {
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_EXPIRED
	}

	switch status {
	case consts.CLUSTERING_STATUS_QUEUED:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_QUEUED
	case consts.CLUSTERING_STATUS_RUNNING:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_RUNNING
	case consts.CLUSTERING_STATUS_SUCCEEDED:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_SUCCEEDED
	case consts.CLUSTERING_STATUS_FAILED:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_FAILED
	case consts.CLUSTERING_STATUS_CANCELED:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_CANCELED
	case consts.CLUSTERING_STATUS_EXPIRED:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_EXPIRED
	default:
		return logEngineProto.ClusteringJobStatus_CLUSTERING_JOB_STATUS_UNSPECIFIED
	}
}

func RunDBSCANCosine(vectors [][]float32, eps float64, minPoints int) []int32 {
	if len(vectors) == 0 {
		return nil
	}
	if minPoints < 2 {
		minPoints = 2
	}

	n := len(vectors)
	labels := make([]int32, n)
	for i := range labels {
		labels[i] = math.MinInt32
	}

	clusterID := int32(0)
	for i := 0; i < n; i++ {
		if labels[i] != math.MinInt32 {
			continue
		}

		neighbors := cosineNeighbors(vectors, i, eps)
		if len(neighbors) < minPoints {
			labels[i] = -1
			continue
		}

		labels[i] = clusterID
		seed := make([]int, len(neighbors))
		copy(seed, neighbors)

		for j := 0; j < len(seed); j++ {
			point := seed[j]
			if labels[point] == -1 {
				labels[point] = clusterID
			}
			if labels[point] != math.MinInt32 {
				continue
			}
			labels[point] = clusterID

			pointNeighbors := cosineNeighbors(vectors, point, eps)
			if len(pointNeighbors) >= minPoints {
				seed = append(seed, pointNeighbors...)
			}
		}

		clusterID++
	}

	for i := range labels {
		if labels[i] == math.MinInt32 {
			labels[i] = -1
		}
	}

	return labels
}

func ClusterByExactValue(values []string) []int32 {
	labels := make([]int32, len(values))
	nextClusterID := int32(0)
	clusterByValue := make(map[string]int32)

	for idx, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			labels[idx] = -1
			continue
		}

		clusterID, exists := clusterByValue[normalized]
		if !exists {
			clusterID = nextClusterID
			clusterByValue[normalized] = clusterID
			nextClusterID++
		}

		labels[idx] = clusterID
	}

	return labels
}

func cosineNeighbors(vectors [][]float32, index int, eps float64) []int {
	neighbors := make([]int, 0, 16)
	target := vectors[index]
	for i := 0; i < len(vectors); i++ {
		distance := cosineDistance(target, vectors[i])
		if distance <= eps {
			neighbors = append(neighbors, i)
		}
	}
	return neighbors
}

func cosineDistance(a []float32, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 2
	}

	var dot float64
	var normA float64
	var normB float64
	for i := 0; i < len(a); i++ {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		normA += av * av
		normB += bv * bv
	}

	if normA == 0 || normB == 0 {
		return 2
	}

	cosine := dot / (math.Sqrt(normA) * math.Sqrt(normB))
	if cosine > 1 {
		cosine = 1
	}
	if cosine < -1 {
		cosine = -1
	}
	return 1 - cosine
}
