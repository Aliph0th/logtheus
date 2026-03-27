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
