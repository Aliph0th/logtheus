package services

import (
	"context"
	"fmt"
	"logtheus/ingestion/internal/config"
	"logtheus/ingestion/internal/consts"
	"logtheus/ingestion/internal/types"
	sharedConsts "logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	applicationProto "logtheus/shared/pkg/pb/v1/application"
	ingestionProto "logtheus/shared/pkg/pb/v1/ingestion"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type IngestionService struct {
	applicationClient applicationProto.ApplicationServiceClient
	logEngineClient   logEngineProto.LogEngineServiceClient
	cfg               *config.AppConfig
}

func NewIngestionService(
	applicationClient applicationProto.ApplicationServiceClient,
	logEngineClient logEngineProto.LogEngineServiceClient,
	cfg *config.AppConfig,
) *IngestionService {
	return &IngestionService{
		applicationClient: applicationClient,
		logEngineClient:   logEngineClient,
		cfg:               cfg,
	}
}

func (s *IngestionService) IngestLogs(ctx context.Context, req *ingestionProto.IngestLogRequest) ([]*types.NormalizedLog, error) {
	if len(req.Logs) == 0 {
		return nil, grpc.WithInvalidArgument("Batch must contain at least one log").WithSlug(sharedConsts.ERROR_CODE_VALIDATION_FAILED)
	}

	totalBatchSize := 0
	for i, logItem := range req.Logs {
		if logItem == nil {
			return nil, grpc.WithInvalidArgument(fmt.Sprintf("logs[%d] item must be an object", i)).WithSlug(sharedConsts.ERROR_CODE_VALIDATION_FAILED)
		}

		if len(logItem.RawData) == 0 {
			return nil, grpc.WithInvalidArgument(fmt.Sprintf("logs[%d].raw_data is required", i)).WithSlug(sharedConsts.ERROR_CODE_VALIDATION_FAILED)
		}

		totalBatchSize += len(logItem.RawData)
	}

	if totalBatchSize > consts.MAX_INGESTION_BYTES {
		msg := fmt.Sprintf("Batch payload exceeds max allowed size: %d bytes, received: %d bytes", consts.MAX_INGESTION_BYTES, totalBatchSize)
		return nil, grpc.WithResourceExhausted(msg).WithSlug(sharedConsts.ERROR_CODE_TOO_LARGE_PAYLOAD)
	}

	appInfo, err := s.applicationClient.ValidateApiKey(ctx, &applicationProto.ValidateApiKeyRequest{ApiKey: req.ApiKey})
	if err != nil {
		return nil, err
	}

	normalizedLogs := make([]*types.NormalizedLog, 0, len(req.Logs))
	logEngineLogs := make([]*logEngineProto.LogItem, 0, len(req.Logs))
	for _, logItem := range req.Logs {
		format := DetectFormat(logItem.RawData)

		normalized := &types.NormalizedLog{
			APIKey:       req.ApiKey,
			KeySignature: extractAPIKeySignature(req.ApiKey),
			Format:       format,
			SourceIP:     req.SourceIp,
			ReceivedAt:   time.Now().UTC(),
			Attributes:   map[string]string{},
		}

		if err := ParseIntoNormalized(logItem.RawData, format, normalized); err != nil {
			return nil, err
		}

		logEngineLogs = append(logEngineLogs, &logEngineProto.LogItem{
			ApplicationId:   appInfo.ApplicationId,
			ApplicationName: appInfo.ApplicationName,
			ProjectId:       appInfo.ProjectId,
			Format:          string(normalized.Format),
			SourceIp:        normalized.SourceIP,
			ReceivedAt:      timestamppb.New(normalized.ReceivedAt),
			RawData:         logItem.RawData,
			Attributes:      normalized.Attributes,
		})

		normalizedLogs = append(normalizedLogs, normalized)
	}

	if _, err := s.logEngineClient.SaveLogs(ctx, &logEngineProto.SaveLogsRequest{Logs: logEngineLogs}); err != nil {
		return nil, err
	}

	return normalizedLogs, nil
}

func extractAPIKeySignature(apiKey string) string {
	parts := strings.Split(apiKey, "_")
	if len(parts) >= 3 {
		return parts[1]
	}
	return ""
}
