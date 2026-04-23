package dto

import (
	"fmt"
	"logtheus/gateway/internal/utils"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type IngestLogsRequest struct {
	Logs []string `json:"logs" binding:"required"`
}

type LogMetricsQuery struct {
	ProjectID     uint64 `form:"project_id" binding:"required"`
	ApplicationID uint64 `form:"application_id"`
	From          string `form:"from" binding:"required"`
	To            string `form:"to" binding:"required"`
	Formats       string `form:"formats"`
	SourceIPs     string `form:"source_ips"`
}

type LogMetricsVolumeQuery struct {
	LogMetricsQuery
	Bucket string `form:"bucket"`
}

type LogMetricsAggregationQuery struct {
	LogMetricsQuery
	Field string `form:"field" binding:"required"`
	Limit uint32 `form:"limit"`
}

type LogClusteringStartRequest struct {
	ProjectID     uint64  `json:"project_id" binding:"required"`
	ApplicationID uint64  `json:"application_id"`
	From          string  `json:"from" binding:"required"`
	To            string  `json:"to" binding:"required"`
	ClusterBy     string  `json:"cluster_by"`
	Eps           float64 `json:"eps"`
	MinPoints     uint32  `json:"min_points"`
	MaxPoints     uint32  `json:"max_points"`
	TtlHours      uint32  `json:"ttl_hours"`
	RequestKey    string  `json:"request_key"`
}

type LogClusteringResultQuery struct {
	Offset *uint32 `form:"offset"`
	Limit  *uint32 `form:"limit"`
}

func (q *LogMetricsQuery) ToProtoFilter() (*logEngineProto.LogAggregationFilter, error) {
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(q.From))
	if err != nil {
		return nil, fmt.Errorf("invalid from value")
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(q.To))
	if err != nil {
		return nil, fmt.Errorf("invalid to value")
	}

	filter := &logEngineProto.LogAggregationFilter{
		ProjectId: q.ProjectID,
		From:      timestamppb.New(from.UTC()),
		To:        timestamppb.New(to.UTC()),
		Formats:   utils.SplitCSV(q.Formats),
		SourceIps: utils.SplitCSV(q.SourceIPs),
	}

	for i := range filter.Formats {
		filter.Formats[i] = strings.ToUpper(strings.TrimSpace(filter.Formats[i]))
	}

	if q.ApplicationID > 0 {
		applicationID := q.ApplicationID
		filter.ApplicationId = &applicationID
	}

	return filter, nil
}

func (q *LogMetricsVolumeQuery) ToProtoRequest() (*logEngineProto.GetVolumeSeriesRequest, error) {
	filter, err := q.ToProtoFilter()
	if err != nil {
		return nil, err
	}

	return &logEngineProto.GetVolumeSeriesRequest{
		Filter: filter,
		Bucket: utils.MapTimeBucket(q.Bucket),
	}, nil
}

func (q *LogMetricsAggregationQuery) ToProtoRequest() (*logEngineProto.GetAggregationRequest, error) {
	filter, err := q.ToProtoFilter()
	if err != nil {
		return nil, err
	}

	return &logEngineProto.GetAggregationRequest{
		Filter: filter,
		Field:  q.Field,
		Limit:  q.Limit,
	}, nil
}

func (q *LogMetricsQuery) ToProtoLatencyRequest() (*logEngineProto.GetLatencyStatsRequest, error) {
	filter, err := q.ToProtoFilter()
	if err != nil {
		return nil, err
	}

	return &logEngineProto.GetLatencyStatsRequest{Filter: filter}, nil
}

func (q *LogClusteringResultQuery) ToProtoRequest(jobID string) *logEngineProto.GetClusteringJobResultRequest {
	offset := uint32(0)
	if q.Offset != nil {
		offset = *q.Offset
	}
	limit := uint32(100)
	if q.Limit != nil {
		limit = *q.Limit
	}
	return &logEngineProto.GetClusteringJobResultRequest{
		JobId:  strings.TrimSpace(jobID),
		Offset: offset,
		Limit:  limit,
	}
}

func (r *LogClusteringStartRequest) ToProtoRequest() (*logEngineProto.StartClusteringJobRequest, error) {
	from, err := time.Parse(time.RFC3339, strings.TrimSpace(r.From))
	if err != nil {
		return nil, fmt.Errorf("invalid from value")
	}
	to, err := time.Parse(time.RFC3339, strings.TrimSpace(r.To))
	if err != nil {
		return nil, fmt.Errorf("invalid to value")
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("from must be before to")
	}

	filter := &logEngineProto.ClusteringFilter{
		ProjectId: r.ProjectID,
		From:      timestamppb.New(from.UTC()),
		To:        timestamppb.New(to.UTC()),
	}

	if r.ApplicationID > 0 {
		filter.ApplicationId = &r.ApplicationID
	}

	params := &logEngineProto.ClusteringParams{
		Eps:       r.Eps,
		MinPoints: r.MinPoints,
		MaxPoints: r.MaxPoints,
	}

	return &logEngineProto.StartClusteringJobRequest{
		Filter:     filter,
		Params:     params,
		RequestKey: strings.TrimSpace(r.RequestKey),
		TtlHours:   r.TtlHours,
		ClusterBy:  strings.ToLower(strings.TrimSpace(r.ClusterBy)),
	}, nil
}
