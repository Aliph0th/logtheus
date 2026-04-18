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
