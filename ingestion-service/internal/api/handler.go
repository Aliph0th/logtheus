package api

import (
	"context"
	"logtheus/ingestion/internal/services"
	ingestionProto "logtheus/shared/pkg/pb/v1/ingestion"
)

type IngestionHandler struct {
	ingestionProto.UnimplementedIngestionServiceServer
	ingestionService *services.IngestionService
}

func NewIngestionHandler(ingestionService *services.IngestionService) *IngestionHandler {
	return &IngestionHandler{
		ingestionService: ingestionService,
	}
}

func (h *IngestionHandler) IngestLogs(ctx context.Context, req *ingestionProto.IngestLogRequest) (*ingestionProto.IngestLogResponse, error) {
	normalized, err := h.ingestionService.IngestLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	return &ingestionProto.IngestLogResponse{Success: true, AcceptedCount: uint32(len(normalized))}, nil
}
