package api

import (
	"context"
	"logtheus/logengine/internal/services"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
)

type LogEngineHandler struct {
	logEngineProto.UnimplementedLogEngineServiceServer
	logEngineService *services.LogEngineService
}

func NewLogEngineHandler(logEngineService *services.LogEngineService) *LogEngineHandler {
	return &LogEngineHandler{
		logEngineService: logEngineService,
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
