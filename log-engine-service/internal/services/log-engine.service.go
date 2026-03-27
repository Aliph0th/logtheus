package services

import (
	"context"
	"encoding/json"
	"logtheus/logengine/internal/config"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/repository"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
)

type LogEngineService struct {
	cfg  *config.AppConfig
	repo *repository.LogRepository
	s3   *S3Service
}

func NewLogEngineService(
	cfg *config.AppConfig,
	repo *repository.LogRepository,
	s3 *S3Service,
) *LogEngineService {
	return &LogEngineService{
		cfg:  cfg,
		repo: repo,
		s3:   s3,
	}
}

func (s *LogEngineService) SaveLogs(ctx context.Context, req *logEngineProto.SaveLogsRequest) error {
	if len(req.Logs) == 0 {
		return grpc.WithInvalidArgument("Batch must contain at least one log").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
	}

	//TODO:
	// first := req.Logs[0]
	// for _, item := range req.Logs[1:] {
	// 	if item.ProjectId != first.ProjectId || item.ApplicationId != first.ApplicationId || item.ApplicationName != first.ApplicationName {
	// 		return grpc.WithInvalidArgument("All logs in batch must belong to the same project and application").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
	// 	}
	// }

	s3Key, err := s.s3.UploadBatch(ctx, req.Logs)
	if err != nil {
		return grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_UPLOAD_FAILED)
	}

	logRecords := make([]*models.LogRecord, 0, len(req.Logs))
	for _, item := range req.Logs {
		if item.ReceivedAt == nil {
			return grpc.WithInvalidArgument("received_at is required for each log").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
		}

		attributesJSON, marshalErr := json.Marshal(item.Attributes)
		if marshalErr != nil {
			return grpc.WithInvalidArgument("Invalid attributes payload").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
		}

		logRecords = append(logRecords, &models.LogRecord{
			ApplicationID:   item.ApplicationId,
			ApplicationName: item.ApplicationName,
			ProjectID:       item.ProjectId,
			Format:          item.Format,
			SourceIP:        item.SourceIp,
			ReceivedAt:      item.ReceivedAt.AsTime().UTC(),
			Attributes:      json.RawMessage(attributesJSON),
			S3Key:           s3Key,
		})
	}

	if err := s.repo.SaveBatch(ctx, logRecords); err != nil {
		_ = s.s3.DeleteObject(ctx, s3Key)
		return grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	return nil
}
