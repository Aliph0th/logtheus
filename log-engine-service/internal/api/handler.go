package api

import (
	"context"
	"logtheus/logengine/internal/services"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
)

type LogEngineHandler struct {
	logEngineProto.UnimplementedLogEngineServiceServer
	logEngineService  *services.LogEngineService
	clusteringService *services.ClusteringService
}

func NewLogEngineHandler(logEngineService *services.LogEngineService, clusteringService *services.ClusteringService) *LogEngineHandler {
	return &LogEngineHandler{
		logEngineService:  logEngineService,
		clusteringService: clusteringService,
	}
}

func (h *LogEngineHandler) SaveLogs(ctx context.Context, req *logEngineProto.SaveLogsRequest) (*logEngineProto.SaveLogsResponse, error) {
	err := h.logEngineService.SaveLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	return &logEngineProto.SaveLogsResponse{
		Ok:         true,
		SavedCount: uint32(len(req.Logs)),
	}, nil
}

func (h *LogEngineHandler) CheckSimilarLogs(ctx context.Context, req *logEngineProto.CheckSimilarLogsRequest) (*logEngineProto.CheckSimilarLogsResponse, error) {
	return h.logEngineService.CheckSimilarLogs(ctx, req)
}

func (h *LogEngineHandler) GetVolumeSeries(ctx context.Context, req *logEngineProto.GetVolumeSeriesRequest) (*logEngineProto.GetVolumeSeriesResponse, error) {
	points, err := h.logEngineService.GetVolumeSeries(ctx, req)
	if err != nil {
		return nil, err
	}

	return &logEngineProto.GetVolumeSeriesResponse{
		Points: points,
	}, nil
}

func (h *LogEngineHandler) GetAggregationByField(ctx context.Context, req *logEngineProto.GetAggregationRequest) (*logEngineProto.GetAggregationResponse, error) {
	items, err := h.logEngineService.GetAggregationByField(ctx, req)
	if err != nil {
		return nil, err
	}

	return &logEngineProto.GetAggregationResponse{
		Items: items,
	}, nil
}

func (h *LogEngineHandler) GetLatencyStats(ctx context.Context, req *logEngineProto.GetLatencyStatsRequest) (*logEngineProto.GetLatencyStatsResponse, error) {
	stats, err := h.logEngineService.GetLatencyStats(ctx, req)
	if err != nil {
		return nil, err
	}

	return &logEngineProto.GetLatencyStatsResponse{
		Stats: stats,
	}, nil
}

func (h *LogEngineHandler) StartClusteringJob(ctx context.Context, req *logEngineProto.StartClusteringJobRequest) (*logEngineProto.StartClusteringJobResponse, error) {
	return h.clusteringService.StartClusteringJob(ctx, req)
}

func (h *LogEngineHandler) GetClusteringJobStatus(ctx context.Context, req *logEngineProto.GetClusteringJobStatusRequest) (*logEngineProto.GetClusteringJobStatusResponse, error) {
	return h.clusteringService.GetClusteringJobStatus(ctx, req)
}

func (h *LogEngineHandler) GetClusteringJobs(ctx context.Context, req *logEngineProto.GetClusteringJobsRequest) (*logEngineProto.GetClusteringJobsResponse, error) {
	return h.clusteringService.GetClusteringJobs(ctx, req)
}

func (h *LogEngineHandler) GetClusteringJobResult(ctx context.Context, req *logEngineProto.GetClusteringJobResultRequest) (*logEngineProto.GetClusteringJobResultResponse, error) {
	return h.clusteringService.GetClusteringJobResult(ctx, req)
}

func (h *LogEngineHandler) CancelClusteringJob(ctx context.Context, req *logEngineProto.CancelClusteringJobRequest) (*logEngineProto.CancelClusteringJobResponse, error) {
	return h.clusteringService.CancelClusteringJob(ctx, req)
}
