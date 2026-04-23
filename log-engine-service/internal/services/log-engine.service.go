package services

import (
	"context"
	"encoding/json"
	"logtheus/logengine/internal/config"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/repository"
	"logtheus/logengine/internal/utils"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type LogEngineService struct {
	cfg         *config.AppConfig
	repo        *repository.LogRepository
	featureRepo *repository.LogFeatureRepository
	s3          *S3Service
	logIdentity *LogIdentityService
}

func NewLogEngineService(
	cfg *config.AppConfig,
	repo *repository.LogRepository,
	featureRepo *repository.LogFeatureRepository,
	s3 *S3Service,
	logIdentity *LogIdentityService,
) *LogEngineService {
	return &LogEngineService{
		cfg:         cfg,
		repo:        repo,
		featureRepo: featureRepo,
		s3:          s3,
		logIdentity: logIdentity,
	}
}

func (s *LogEngineService) SaveLogs(ctx context.Context, req *logEngineProto.SaveLogsRequest) error {
	if len(req.Logs) == 0 {
		return grpc.WithInvalidArgument("Batch must contain at least one log").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
	}

	//TODO: validate?
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
		//TODO: validate?
		logID := s.logIdentity.BuildLogIDFromRawData(
			item.ProjectId,
			item.ApplicationId,
			item.SourceIp,
			item.RawData,
		)

		attributesJSON, marshalErr := json.Marshal(item.Attributes)
		if marshalErr != nil {
			return grpc.WithInvalidArgument("Invalid attributes payload").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
		}

		logRecords = append(logRecords, &models.LogRecord{
			LogID:           logID,
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

func (s *LogEngineService) GetVolumeSeries(ctx context.Context, req *logEngineProto.GetVolumeSeriesRequest) ([]*logEngineProto.TimeSeriesPoint, error) {
	rows, repoErr := s.repo.GetVolumeSeries(ctx, req.Filter, req.Bucket)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(consts.INTERNAL_ERROR_AGGREGATION_FAILED)
	}

	points := make([]*logEngineProto.TimeSeriesPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, &logEngineProto.TimeSeriesPoint{
			Timestamp: timestamppb.New(row.Bucket),
			Count:     row.Count,
		})
	}

	return points, nil
}

func (s *LogEngineService) GetAggregationByField(ctx context.Context, req *logEngineProto.GetAggregationRequest) ([]*logEngineProto.AggregationItem, error) {
	if _, ok := utils.AggregationFieldExpression(req.GetField()); !ok {
		return nil, grpc.WithInvalidArgument("unsupported aggregation field").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
	}

	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	rows, repoErr := s.repo.GetAggregationByField(ctx, req.Filter, req.Field, limit)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(consts.INTERNAL_ERROR_AGGREGATION_FAILED)
	}

	items := make([]*logEngineProto.AggregationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &logEngineProto.AggregationItem{
			Value: row.Value,
			Count: row.Count,
		})
	}

	return items, nil
}

func (s *LogEngineService) GetLatencyStats(ctx context.Context, req *logEngineProto.GetLatencyStatsRequest) (*logEngineProto.LatencyStats, error) {
	stats, repoErr := s.repo.GetLatencyStats(ctx, req.Filter)
	if repoErr != nil {
		return nil, grpc.WithInternal(repoErr).WithSlug(consts.INTERNAL_ERROR_AGGREGATION_FAILED)
	}

	return &logEngineProto.LatencyStats{
		P50: stats.P50,
		P95: stats.P95,
		P99: stats.P99,
	}, nil
}
