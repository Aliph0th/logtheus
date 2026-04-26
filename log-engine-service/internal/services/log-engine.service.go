package services

import (
	"context"
	"encoding/json"
	"fmt"
	"logtheus/logengine/internal/config"
	internalConsts "logtheus/logengine/internal/consts"
	"logtheus/logengine/internal/models"
	"logtheus/logengine/internal/repository"
	"logtheus/shared/pkg/consts"
	"logtheus/shared/pkg/grpc"
	logEngineProto "logtheus/shared/pkg/pb/v1/logengine"
	projectProto "logtheus/shared/pkg/pb/v1/project"
	"time"

	"logtheus/logengine/internal/utils"

	"github.com/pgvector/pgvector-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LogEngineService struct {
	cfg           *config.AppConfig
	repo          *repository.LogRepository
	featureRepo   *repository.LogFeatureRepository
	projectClient projectProto.ProjectServiceClient
	s3            *S3Service
	logIdentity   *LogIdentityService
}

func NewLogEngineService(
	cfg *config.AppConfig,
	repo *repository.LogRepository,
	featureRepo *repository.LogFeatureRepository,
	projectClient projectProto.ProjectServiceClient,
	s3 *S3Service,
	logIdentity *LogIdentityService,
) *LogEngineService {
	return &LogEngineService{
		cfg:           cfg,
		repo:          repo,
		featureRepo:   featureRepo,
		projectClient: projectClient,
		s3:            s3,
		logIdentity:   logIdentity,
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

	normalLogs := make([]*logEngineProto.LogItem, 0, len(req.Logs))
	aggregateOnlyLogs := make([]*logEngineProto.LogItem, 0)
	for _, item := range req.Logs {
		if _, isAggregateOnly := item.Attributes[internalConsts.SimilarRefLogIDAttribute]; isAggregateOnly {
			aggregateOnlyLogs = append(aggregateOnlyLogs, item)
			continue
		}
		normalLogs = append(normalLogs, item)
	}

	groups := make(map[string][]*logEngineProto.LogItem)
	for _, item := range normalLogs {
		groupKey := buildGroupKey(item.ProjectId, item.ApplicationId)
		groups[groupKey] = append(groups[groupKey], item)
	}

	allRecords := make([]*models.LogRecord, 0, len(req.Logs))
	uploadedS3Keys := make([]string, 0, len(groups))

	for _, groupLogs := range groups {
		s3Key, err := s.s3.UploadBatch(ctx, groupLogs)
		if err != nil {
			for _, prevS3Key := range uploadedS3Keys {
				_ = s.s3.DeleteObject(ctx, prevS3Key)
			}
			return grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_UPLOAD_FAILED)
		}
		uploadedS3Keys = append(uploadedS3Keys, s3Key)

		for _, item := range groupLogs {
			//TODO: validate?
			logID := s.logIdentity.BuildLogIDFromRawData(
				item.ProjectId,
				item.ApplicationId,
				item.SourceIp,
				item.RawData,
			)

			attributesJSON, marshalErr := json.Marshal(item.Attributes)
			if marshalErr != nil {
				for _, prevS3Key := range uploadedS3Keys {
					_ = s.s3.DeleteObject(ctx, prevS3Key)
				}
				return grpc.WithInvalidArgument("Invalid attributes payload").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
			}

			allRecords = append(allRecords, &models.LogRecord{
				LogID:         logID,
				ApplicationID: item.ApplicationId,
				ProjectID:     item.ProjectId,
				Format:        item.Format,
				SourceIP:      item.SourceIp,
				ReceivedAt:    item.ReceivedAt.AsTime().UTC(),
				Attributes:    json.RawMessage(attributesJSON),
				S3Key:         s3Key,
			})
		}
	}

	for _, item := range aggregateOnlyLogs {
		logID := s.logIdentity.BuildLogIDFromRawData(
			item.ProjectId,
			item.ApplicationId,
			item.SourceIp,
			item.RawData,
		)

		attributesJSON, marshalErr := json.Marshal(item.Attributes)
		if marshalErr != nil {
			for _, prevS3Key := range uploadedS3Keys {
				_ = s.s3.DeleteObject(ctx, prevS3Key)
			}
			return grpc.WithInvalidArgument("Invalid attributes payload").WithSlug(consts.ERROR_CODE_VALIDATION_FAILED)
		}

		allRecords = append(allRecords, &models.LogRecord{
			LogID:         logID,
			ApplicationID: item.ApplicationId,
			ProjectID:     item.ProjectId,
			Format:        item.Format,
			SourceIP:      item.SourceIp,
			ReceivedAt:    item.ReceivedAt.AsTime().UTC(),
			Attributes:    json.RawMessage(attributesJSON),
			S3Key:         internalConsts.SimilarAggregateS3Key,
		})
	}

	batchSize := s.cfg.Persistence.LogsCHInsertBatchSize
	if err := s.repo.SaveBatchInChunks(ctx, allRecords, batchSize); err != nil {
		for _, s3Key := range uploadedS3Keys {
			_ = s.s3.DeleteObject(ctx, s3Key)
		}
		return grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
	}

	return nil
}

func (s *LogEngineService) CheckSimilarLogs(ctx context.Context, req *logEngineProto.CheckSimilarLogsRequest) (*logEngineProto.CheckSimilarLogsResponse, error) {
	if len(req.Candidates) == 0 {
		return &logEngineProto.CheckSimilarLogsResponse{Matches: []*logEngineProto.SimilarLogMatch{}}, nil
	}

	minSimilarity := req.MinSimilarity
	if minSimilarity <= 0 {
		minSimilarity = 0.98
	}

	windowDays := req.WindowDays
	if windowDays == 0 {
		windowDays = 30
	}
	windowFrom := time.Now().UTC().AddDate(0, 0, -int(windowDays))

	matches := make([]*logEngineProto.SimilarLogMatch, 0, len(req.Candidates))
	incrementByLogID := make(map[string]uint64)
	for _, candidate := range req.Candidates {
		if len(candidate.Embedding) == 0 {
			matches = append(matches, &logEngineProto.SimilarLogMatch{IsSimilar: false})
			continue
		}

		row, err := s.featureRepo.FindMostSimilarInWindow(
			ctx,
			candidate.ProjectId,
			candidate.ApplicationId,
			pgvector.NewVector(candidate.Embedding),
			minSimilarity,
			windowFrom,
		)
		if err != nil {
			return nil, grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
		}

		if row == nil {
			matches = append(matches, &logEngineProto.SimilarLogMatch{IsSimilar: false})
			continue
		}

		incrementByLogID[row.LogID] = incrementByLogID[row.LogID] + 1

		matches = append(matches, &logEngineProto.SimilarLogMatch{
			IsSimilar:    true,
			MatchedLogId: row.LogID,
			Similarity:   row.Similarity,
		})
	}

	for logID, increment := range incrementByLogID {
		if err := s.featureRepo.IncrementSimilarCounter(ctx, logID, increment); err != nil {
			return nil, grpc.WithInternal(err).WithSlug(consts.INTERNAL_ERROR_CODE_CREATION_FAILED)
		}
	}

	return &logEngineProto.CheckSimilarLogsResponse{Matches: matches}, nil
}

func buildGroupKey(projectID uint64, applicationID uint64) string {
	return fmt.Sprintf("%d:%d", projectID, applicationID)
}

func (s *LogEngineService) GetVolumeSeries(ctx context.Context, req *logEngineProto.GetVolumeSeriesRequest) ([]*logEngineProto.TimeSeriesPoint, error) {
	if accessErr := utils.EnsureProjectReadAccess(ctx, req.GetFilter().GetProjectId(), s.projectClient); accessErr != nil {
		return nil, accessErr
	}

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
	if accessErr := utils.EnsureProjectReadAccess(ctx, req.GetFilter().GetProjectId(), s.projectClient); accessErr != nil {
		return nil, accessErr
	}

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
	if accessErr := utils.EnsureProjectReadAccess(ctx, req.GetFilter().GetProjectId(), s.projectClient); accessErr != nil {
		return nil, accessErr
	}

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
